package server

import (
	"context"
	"log"
	"net"
)

// Server serves tcp and udp connections
type Server struct {
}

// NewServer returns instance of Server
func NewServer() *Server {
	return &Server{}
}

func handlerFunc(ctx context.Context, conn net.Conn) {
	_, err := conn.Write([]byte("please stay home\n"))
	if err != nil {
		log.Printf("write failed: %v\n", err)
	}

	err = conn.Close()
	if err != nil {
		log.Printf("failed to close conn: %v\n", err)
	}
}

func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
		// handle error
	}
	log.Println("ready to accept connections")
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
			// handle error
		}
		go handlerFunc(ctx, conn)
	}
}
