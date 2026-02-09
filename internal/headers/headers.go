package headers

import (
	"fmt"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Get(key string) string {
	// Lowercase the key for case-insensitive lookup
	return h[strings.ToLower(key)]
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

	// Validate: key contains only valid characters
	if !isValidHeaderKey(key) {
		return 0, false, fmt.Errorf("invalid header: invalid character in key")
	}

	// Lowercase the key
	key = strings.ToLower(key)

	// Trim whitespace from value only
	value = strings.TrimSpace(value)

	// Check if key already exists - append with comma if so
	if existing, exists := h[key]; exists {
		h[key] = existing + ", " + value
	} else {
		h[key] = value
	}

	// Return bytes consumed (line + CRLF)
	return idx + 2, false, nil
}


func isValidHeaderKey(key string) bool {
	// RFC 9110: token characters
	//token = 1*tchar
	// tchar = "!" / "#" / "$" / "%" / "&" / "'" / "*" / "+" / "-" / "." /
	//         "0"-"9" / "A"-"Z" / "^" / "_" / "`" / "a"-"z" / "|" / "~"
	for _, c := range key {
		if !isTokenChar(c) {
			return false
		}
	}
	return len(key) > 0
}

func isTokenChar(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '!' || c == '#' || c == '$' || c == '%' || c == '&' ||
		c == '\'' || c == '*' || c == '+' || c == '-' || c == '.' ||
		c == '^' || c == '_' || c == '`' || c == '|' || c == '~'
}

func (h Headers) Set(key, value string) {
	h[strings.ToLower(key)] = value
}
