package server

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"sync"

	"github.com/scukonick/friends/internal/dispatcher"
)

// TCPServer serves tcp connections
type TCPServer struct {
	disp dispatcher.Dispatcher
}

// NewTCPServer returns instance of TCPServer
func NewTCPServer(disp dispatcher.Dispatcher) *TCPServer {
	return &TCPServer{
		disp: disp,
	}
}

func (s *TCPServer) handlerFunc(ctx context.Context, wg *sync.WaitGroup, conn net.Conn) {
	defer wg.Done()
	defer safeClose(conn)

	req := &dispatcher.Request{}
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	err := decoder.Decode(req)
	if err != nil {
		log.Printf("ERR: failed to decode input")
		return
	}

	readCtx, cancel := context.WithCancel(ctx)
	msgs, err := s.disp.Connect(readCtx, req)
	if err != nil {
		log.Printf("failed to connect: %v", err)
		return
	}

	innerWG := &sync.WaitGroup{}
	innerWG.Add(1)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		for m := range msgs {
			err := encoder.Encode(m)
			if err != nil {
				log.Printf("failed to encode msg: %v", err)
			}
		}
	}(innerWG)

	for {
		buf := make([]byte, 1)
		_, err := conn.Read(buf)
		if err != nil {
			// connection failed
			cancel()
			break
		}
	}

	// connection closed, sending our farewells
	err = s.disp.Disconnect(ctx, req)
	if err != nil {
		log.Printf("ERR: failed to disconnect")
	}

	wg.Wait()
}

func (s *TCPServer) ListenAndServe(ctx context.Context, addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	log.Println("ready to accept tcp connections")
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			conn, err := ln.Accept()
			if err != nil {
				break
			}

			wg.Add(1)
			go s.handlerFunc(ctx, wg, conn)
		}
	}()

	<-ctx.Done()
	err = ln.Close()
	if err != nil {
		log.Printf("ERR: failed to stop listening: %v", err)
	}

	wg.Wait()
	log.Println("exiting")

	return nil
}

func safeClose(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Printf("failed to close connection: %v", err)
	}
}
