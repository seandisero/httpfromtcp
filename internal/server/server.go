package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/seandisero/httpfromtcp/internal/request"
	"github.com/seandisero/httpfromtcp/internal/response"
)

// server is HTTP 1.1 server
type Server struct {
	listener net.Listener
	handler  Handler
	closed   atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	svr := &Server{
		listener: listener,
		handler:  handler,
	}
	go svr.listen()
	return svr, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	if s.listener == nil {
		return nil
	}
	err := s.listener.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("error accepting tcp connection, loop closed: %v", err)
			return
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	responseWriter := response.NewWriter(conn)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		responseWriter.WriteStatusLine(response.StatusBadRequest)
		responseWriter.WriteHeaders(response.GetDefaultHeaders(0))
		return
	}

	s.handler(responseWriter, req)
}
