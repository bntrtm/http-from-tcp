package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/bntrtm/http-from-tcp/internal/request"
	"github.com/bntrtm/http-from-tcp/internal/response"
)

type Handler func(w *response.Writer, r *request.Request)

type serverState int

const (
	StatusListening serverState = iota
	StatusClosed
)

// Server is an HTTP 1.1 server
type Server struct {
	listener net.Listener
	isClosed atomic.Bool
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	var isServerClosed atomic.Bool
	isServerClosed.Store(false)

	s := &Server{
		listener: listener,
		handler:  handler,
	}

	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	s.isClosed.Store(true)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	w := response.NewWriter(conn)

	r, err := request.RequestFromReader(conn)
	if err != nil {
		_ = w.WriteStatusLine(response.StatusBadRequest)
		body := fmt.Appendf([]byte{}, "Error parsing request: %v", err)
		_ = w.WriteHeaders(response.GetDefaultHeaders(len(body)))
		_, _ = w.WriteBody(body)
		return
	}

	s.handler(w, r)
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.isClosed.Load() {
				return
			}
			log.Fatalf("Could not establish HTTP connection: %s\n", err)
			return
		}
		fmt.Println("Connection accepted:", conn.LocalAddr())

		go s.handle(conn)
	}
}
