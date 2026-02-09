package request

import (
	"fmt"
	"io"
	"strings"
)

const (
	stateInitialized = 0
	stateDone        = 1
	bufferSize       = 8
)

type Request struct {
	RequestLine RequestLine
	state       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0

	req := &Request{state: stateInitialized}

	for req.state != stateDone {
		// Grow buffer if full
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		// Read from reader
		n, err := reader.Read(buf[readToIndex:])
		if err == io.EOF {
			req.state = stateDone
			break
		}
		if err != nil {
			return nil, err
		}

		readToIndex += n

		// Parse what we've read so far
		parsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		// Remove parsed data from buffer
		if parsed > 0 {
			copy(buf, buf[parsed:readToIndex])
			readToIndex -= parsed
		}
	}

	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case stateInitialized:
		consumed, reqLine, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		}
		if consumed == 0 {
			// Need more data
			return 0, nil
		}
		r.RequestLine = reqLine
		r.state = stateDone
		return consumed, nil

	case stateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")

	default:
		return 0, fmt.Errorf("error: unknown state")
	}
}

func parseRequestLine(httpRequest string) (int, RequestLine, error) {
	// Find first \r\n
	idx := strings.Index(httpRequest, "\r\n")
	if idx == -1 {
		// Need more data
		return 0, RequestLine{}, nil
	}

	requestLine := httpRequest[:idx]
	parts := strings.Fields(requestLine)

	if len(parts) != 3 {
		return 0, RequestLine{}, fmt.Errorf("invalid number of parts in request line")
	}

	method, target, version := parts[0], parts[1], parts[2]

	// Validate HTTP/1.1
	if version != "HTTP/1.1" {
		return 0, RequestLine{}, fmt.Errorf("invalid version: %s", version)
	}

	// Validate method
	if !isValidMethod(method) {
		return 0, RequestLine{}, fmt.Errorf("invalid method: %s", method)
	}

	return idx + 2, RequestLine{
		Method:        method,
		RequestTarget: target,
		HttpVersion:   "1.1",
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
