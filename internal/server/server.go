package server

import (
	"fmt"
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

// Handler now takes response.Writer instead of io.Writer
type Handler func(req *request.Request, w *response.Writer) error

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener: listener,
		handler:  handler,
	}

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
			if s.closed.Load() {
				return
			}
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	// Parse request
	req, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Println("Error parsing request:", err)
		return
	}

	// Create response writer
	w := response.NewWriter(conn)

	// Call handler
	err = s.handler(req, w)
	if err != nil {
		fmt.Println("Handler error:", err)
		return
	}
}
