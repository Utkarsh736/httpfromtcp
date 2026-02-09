package server

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync/atomic"

	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
	handler  Handler
}

// Handler function type
type Handler func(req *request.Request, w io.Writer) *HandlerError

// HandlerError represents an error with status code
type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener: listener,
		handler:  handler,
	}

	// Start listening in background
	go s.listen()

	return s, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// If server is closed, ignore errors
			if s.closed.Load() {
				return
			}
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Handle each connection in a new goroutine
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	// Parse request
	req, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Println("Error parsing request:", err)
		writeHandlerError(conn, &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    "Bad Request\n",
		})
		return
	}

	// Create buffer for handler to write to
	buf := &bytes.Buffer{}

	// Call handler
	handlerErr := s.handler(req, buf)

	if handlerErr != nil {
		// Handler returned error
		writeHandlerError(conn, handlerErr)
		return
	}

	// Success - write response
	bodyBytes := buf.Bytes()

	// Write status line
	err = response.WriteStatusLine(conn, response.StatusOK)
	if err != nil {
		fmt.Println("Error writing status line:", err)
		return
	}

	// Get default headers with correct content length
	headers := response.GetDefaultHeaders(len(bodyBytes))

	// Write headers
	err = response.WriteHeaders(conn, headers)
	if err != nil {
		fmt.Println("Error writing headers:", err)
		return
	}

	// Write body
	_, err = conn.Write(bodyBytes)
	if err != nil {
		fmt.Println("Error writing body:", err)
		return
	}
}

func writeHandlerError(w io.Writer, handlerErr *HandlerError) {
	// Write status line
	err := response.WriteStatusLine(w, handlerErr.StatusCode)
	if err != nil {
		fmt.Println("Error writing error status line:", err)
		return
	}

	// Get headers with error message length
	headers := response.GetDefaultHeaders(len(handlerErr.Message))

	// Write headers
	err = response.WriteHeaders(w, headers)
	if err != nil {
		fmt.Println("Error writing error headers:", err)
		return
	}

	// Write error message body
	_, err = w.Write([]byte(handlerErr.Message))
	if err != nil {
		fmt.Println("Error writing error body:", err)
		return
	}
}
