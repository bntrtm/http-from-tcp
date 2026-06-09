package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/bntrtm/http-from-tcp/internal/response"
)

type serverState int

const (
	StatusListening serverState = iota
	StatusClosed
)

type Server struct {
	listener net.Listener
	isClosed atomic.Bool
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	var isServerClosed atomic.Bool
	isServerClosed.Store(false)

	s := &Server{
		listener: listener,
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
	err := response.WriteStatusLine(conn, 200)
	if err != nil {
		log.Println(err)
		return
	}
	err = response.WriteHeaders(conn, response.GetDefaultHeaders(0))
	if err != nil {
		log.Println(err)
	}
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
