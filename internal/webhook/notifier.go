// Package webhook handles notifications to n8n
package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Notifier sends webhook notifications
type Notifier struct {
	url     string
	headers map[string]string
	timeout int
	retries int
}

// NewNotifier creates a new webhook notifier
func NewNotifier(url string, headers map[string]string, timeout, retries int) *Notifier {
	return &Notifier{
		url:     url,
		headers: headers,
		timeout: timeout,
		retries: retries,
	}
}

// Payload represents a webhook notification
type Payload struct {
	Event     string                 `json:"event"`
	Timestamp string                 `json:"timestamp"`
	Pipeline  PipelineInfo           `json:"pipeline"`
	Backup    BackupInfo             `json:"backup,omitempty"`
	Restore   RestoreInfo            `json:"restore,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// PipelineInfo contains pipeline execution info
type PipelineInfo struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	Duration int    `json:"duration_seconds"`
}

// BackupInfo contains backup info
type BackupInfo struct {
	Status   string   `json:"status"`
	Duration int      `json:"duration_seconds,omitempty"`
	File     FileInfo `json:"file,omitempty"`
}

// RestoreInfo contains restore info
type RestoreInfo struct {
	Status   string `json:"status"`
	Duration int    `json:"duration_seconds,omitempty"`
	Target   string `json:"target,omitempty"`
}

// FileInfo contains file info
type FileInfo struct {
	Path      string `json:"path"`
	SizeBytes int64  `json:"size_bytes"`
	SizeHuman string `json:"size_human"`
}

// Send sends a webhook notification with retries
func (n *Notifier) Send(payload Payload) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= n.retries; attempt++ {
		if err := n.sendWithRetry(jsonData); err != nil {
			lastErr = err
			time.Sleep(time.Second * time.Duration(attempt+1))
			continue
		}
		return nil
	}

	return fmt.Errorf("failed after %d attempts: %w", n.retries, lastErr)
}

func (n *Notifier) sendWithRetry(jsonData []byte) error {
	client := &http.Client{
		Timeout: time.Duration(n.timeout) * time.Second,
	}

	req, err := http.NewRequest("POST", n.url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range n.headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}

// SendSuccess sends a success notification
func (n *Notifier) SendSuccess(pipelineID string, duration int, backup BackupInfo, restore RestoreInfo) error {
	payload := Payload{
		Event:     "backup.pipeline.completed",
		Timestamp: time.Now().Format(time.RFC3339),
		Pipeline: PipelineInfo{
			ID:       pipelineID,
			Status:   "success",
			Duration: duration,
		},
		Backup:  backup,
		Restore: restore,
	}

	return n.Send(payload)
}

// SendFailure sends a failure notification
func (n *Notifier) SendFailure(pipelineID string, phase string, errorMsg string) error {
	payload := Payload{
		Event:     "backup.pipeline.failed",
		Timestamp: time.Now().Format(time.RFC3339),
		Pipeline: PipelineInfo{
			ID:     pipelineID,
			Status: "failed",
		},
		Metadata: map[string]interface{}{
			"phase": phase,
			"error": errorMsg,
		},
	}

	return n.Send(payload)
}
