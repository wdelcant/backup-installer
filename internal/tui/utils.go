package tui

import (
	"strconv"
)

// parsePort converts a string to an integer port number
func parsePort(s string) int {
	port, err := strconv.Atoi(s)
	if err != nil || port < 1 || port > 65535 {
		return 5432 // default PostgreSQL port
	}
	return port
}

// parseInt converts a string to an integer
func parseInt(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 7 // default retention days
	}
	return n
}
