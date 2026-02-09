package request

import (
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return parseRequestLine(string(data))
}

func parseRequestLine(httpRequest string) (*Request, error) {
	lines := strings.Split(httpRequest, "\r\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("no request line found")
	}

	parts := strings.Fields(lines[0])
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid number of parts in request line")
	}

	method, target, version := parts[0], parts[1], parts[2]

	// Validate HTTP/1.1
	if version != "HTTP/1.1" {
		return nil, fmt.Errorf("invalid version: %s", version)
	}

	// Validate method (simple uppercase check for now)
	if !isValidMethod(method) {
		return nil, fmt.Errorf("invalid method: %s", method)
	}

	return &Request{
		RequestLine: RequestLine{
			Method:        method,
			RequestTarget: target,
			HttpVersion:   "1.1",
		},
	}, nil
}

func isValidMethod(method string) bool {
	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return false
		}
	}
	return true
}
