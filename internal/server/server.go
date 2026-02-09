package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener: listener,
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

	// Write status line
	err := response.WriteStatusLine(conn, response.StatusOK)
	if err != nil {
		fmt.Println("Error writing status line:", err)
		return
	}

	// Get default headers (empty body = 0 length)
	headers := response.GetDefaultHeaders(0)

	// Write headers
	err = response.WriteHeaders(conn, headers)
	if err != nil {
		fmt.Println("Error writing headers:", err)
		return
	}

	// No body to write (Content-Length: 0)
}
