package server

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/bntrtm/http-from-tcp/internal/request"
	"github.com/bntrtm/http-from-tcp/internal/response"
)

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

	r, err := request.RequestFromReader(conn)
	if err != nil {
		handlerErr := &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    err.Error(),
		}
		handlerErr.Write(conn)
		return
	}

	buf := bytes.NewBuffer([]byte{})

	handlerErr := s.handler(buf, r)
	if handlerErr != nil {
		handlerErr.Write(conn)
		return
	}

	_ = response.WriteStatusLine(conn, response.StatusOK)
	headers := response.GetDefaultHeaders(len(buf.Bytes()))
	_ = response.WriteHeaders(conn, headers)
	_, _ = conn.Write(buf.Bytes())
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
