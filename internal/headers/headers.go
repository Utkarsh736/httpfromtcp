package headers

import (
	"fmt"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	dataStr := string(data)

	// Look for CRLF
	idx := strings.Index(dataStr, "\r\n")
	if idx == -1 {
		// Need more data
		return 0, false, nil
	}

	// If CRLF is at the start, headers are done
	if idx == 0 {
		return 2, true, nil
	}

	// Extract the line
	line := dataStr[:idx]

	// Find the colon
	colonIdx := strings.Index(line, ":")
	if colonIdx == -1 {
		return 0, false, fmt.Errorf("invalid header: no colon found")
	}

	key := line[:colonIdx]
	value := line[colonIdx+1:]

	// Validate: no whitespace before colon (RFC 9112)
	if strings.TrimSpace(key) != key {
		return 0, false, fmt.Errorf("invalid header: whitespace before colon")
	}

	// Trim whitespace from value only
	value = strings.TrimSpace(value)

	// Store in map
	h[key] = value

	// Return bytes consumed (line + CRLF)
	return idx + 2, false, nil
}

