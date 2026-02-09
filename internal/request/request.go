package request

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"httpfromtcp/internal/headers"
)

const (
	stateInitialized      = 0
	stateParsingHeaders   = 1
	stateParsingBody      = 2
	stateDone             = 3
	bufferSize            = 8
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
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

	req := &Request{
		state:   stateInitialized,
		Headers: headers.NewHeaders(),
		Body:    []byte{},
	}

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
			// Check if we were expecting more body data
			if req.state == stateParsingBody {
				contentLengthStr := req.Headers.Get("Content-Length")
				if contentLengthStr != "" {
					expectedLength, _ := strconv.Atoi(contentLengthStr)
					if len(req.Body) < expectedLength {
						return nil, fmt.Errorf("body shorter than reported content length")
					}
				}
			}
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
	totalBytesParsed := 0

	for r.state != stateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return totalBytesParsed, err
		}
		if n == 0 {
			// Need more data
			break
		}
		totalBytesParsed += n
	}

	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
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
		r.state = stateParsingHeaders
		return consumed, nil

	case stateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = stateParsingBody
		}
		return n, nil

	case stateParsingBody:
		contentLengthStr := r.Headers.Get("Content-Length")
		if contentLengthStr == "" {
			// No content length, done parsing
			r.state = stateDone
			return 0, nil
		}

		expectedLength, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			return 0, fmt.Errorf("invalid Content-Length: %s", contentLengthStr)
		}

		// Append all available data to body
		r.Body = append(r.Body, data...)

		// Check if we have enough
		if len(r.Body) > expectedLength {
			return 0, fmt.Errorf("body longer than reported content length")
		}

		if len(r.Body) == expectedLength {
			r.state = stateDone
		}

		// Consume all the data we were given
		return len(data), nil

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
