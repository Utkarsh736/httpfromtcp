package response

import (
	"fmt"
	"io"
	"strconv"

	"httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type writerState int

const (
	stateStatusLine writerState = 0
	stateHeaders    writerState = 1
	stateBody       writerState = 2
	stateDone       writerState = 3
)

type Writer struct {
	w     io.Writer
	state writerState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:     w,
		state: stateStatusLine,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != stateStatusLine {
		return fmt.Errorf("WriteStatusLine must be called first")
	}

	var reasonPhrase string
	switch statusCode {
	case StatusOK:
		reasonPhrase = "OK"
	case StatusBadRequest:
		reasonPhrase = "Bad Request"
	case StatusInternalServerError:
		reasonPhrase = "Internal Server Error"
	default:
		reasonPhrase = ""
	}

	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)
	_, err := w.w.Write([]byte(statusLine))
	if err != nil {
		return err
	}

	w.state = stateHeaders
	return nil
}

func (w *Writer) WriteHeaders(hdrs headers.Headers) error {
	if w.state != stateHeaders {
		return fmt.Errorf("WriteHeaders must be called after WriteStatusLine and before WriteBody")
	}

	for key, value := range hdrs {
		headerLine := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := w.w.Write([]byte(headerLine))
		if err != nil {
			return err
		}
	}

	// Blank line to end headers
	_, err := w.w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	w.state = stateBody
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != stateBody {
		return 0, fmt.Errorf("WriteBody must be called after WriteHeaders")
	}

	n, err := w.w.Write(p)
	if err != nil {
		return n, err
	}

	w.state = stateDone
	return n, nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h["content-length"] = strconv.Itoa(contentLen)
	h["connection"] = "close"
	h["content-type"] = "text/plain"
	return h
}

// Keep old functions for compatibility (optional)
func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	writer := NewWriter(w)
	return writer.WriteStatusLine(statusCode)
}

func WriteHeaders(w io.Writer, hdrs headers.Headers) error {
	for key, value := range hdrs {
		headerLine := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := w.Write([]byte(headerLine))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	return err
}


func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != stateBody {
		return 0, fmt.Errorf("WriteChunkedBody must be called after WriteHeaders")
	}

	if len(p) == 0 {
		return 0, nil
	}

	// Write chunk size in hex
	chunkSize := fmt.Sprintf("%x\r\n", len(p))
	_, err := w.w.Write([]byte(chunkSize))
	if err != nil {
		return 0, err
	}

	// Write chunk data
	n, err := w.w.Write(p)
	if err != nil {
		return n, err
	}

	// Write trailing CRLF
	_, err = w.w.Write([]byte("\r\n"))
	if err != nil {
		return n, err
	}

	return n, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != stateBody {
		return 0, fmt.Errorf("WriteChunkedBodyDone must be called after WriteHeaders")
	}

	// Write final chunk (0 size)
	finalChunk := "0\r\n\r\n"
	n, err := w.w.Write([]byte(finalChunk))
	if err != nil {
		return n, err
	}

	w.state = stateDone
	return n, nil
}
