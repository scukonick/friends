package server

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"sync"
	"time"

	"github.com/scukonick/friends/internal/dispatcher"
)

type udpConns map[string]*udpConn

type sendFunc func(data []byte) error

type finishFunc func()

type udpConn struct {
	in      chan []byte
	send    sendFunc
	release finishFunc
}

type UDPServer struct {
	disp dispatcher.Dispatcher

	conns udpConns
	lock  sync.RWMutex
}

func NewUDPServer(disp dispatcher.Dispatcher) *UDPServer {
	return &UDPServer{
		disp:  disp,
		conns: make(udpConns),
		lock:  sync.RWMutex{},
	}
}

func (s *UDPServer) handlerFunc(ctx context.Context, wg *sync.WaitGroup, conn *udpConn) {
	defer wg.Done()
	defer conn.release()

	firstMsg := <-conn.in

	req := &dispatcher.Request{}
	err := json.Unmarshal(firstMsg, req)
	if err != nil {
		log.Printf("failed to unmarshal json: %v", err)
		return
	}

	ctx, cancel := context.WithCancel(ctx)

	msgs, err := s.disp.Connect(ctx, req)
	if err != nil {
		log.Printf("failed to connect: %v", err)
		return
	}

	innerWG := &sync.WaitGroup{}

	innerWG.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		for m := range msgs {
			data, err := json.Marshal(m)
			if err != nil {
				log.Printf("failed to encode msg: %v", err)
			}

			err = conn.send(data)
			if err != nil {
				log.Printf("failed to send msg: %v", err)
			}
		}
	}(innerWG)

	innerWG.Add(1)
	go func(wg *sync.WaitGroup) {
		// waiting for keepalive messages
		defer wg.Done()

		for {
			select {
			case <-time.After(15 * time.Second):
				cancel()
			case <-ctx.Done():
				return
			case <-conn.in:
				continue
			}
		}
	}(innerWG)

	innerWG.Wait()

	// connection closed, sending our farewells
	err = s.disp.Disconnect(ctx, req)
	if err != nil {
		log.Printf("ERR: failed to disconnect: %v", err)
	}
}

func (s *UDPServer) ListenAndServe(ctx context.Context, addr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return err
	}
	ln, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		return err
	}

	log.Println("ready to accept udp connections")
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		buffer := make([]byte, 4096)

		for {
			n, addr, err := ln.ReadFromUDP(buffer)
			if err != nil {
				break
			}

			if addr == nil {
				// no connection
				// preventing 100% CPU usage
				time.Sleep(10 * time.Millisecond)
				continue
			}

			// checking if existing connections
			s.lock.RLock()
			addrStr := addr.String()
			c, ok := s.conns[addrStr]
			s.lock.RUnlock()
			if !ok {
				// storing new connection
				c = &udpConn{
					in: make(chan []byte),
					send: func(data []byte) error {
						_, err := ln.WriteTo(data, addr)
						return err
					},
					release: func() {
						s.lock.Lock()
						defer s.lock.Unlock()
						delete(s.conns, addr.String())
					},
				}
				s.lock.Lock()
				s.conns[addr.String()] = c
				s.lock.Unlock()

				wg.Add(1)
				go s.handlerFunc(ctx, wg, c)
			}

			msg := make([]byte, n)
			copy(msg, buffer[:n])
			c.in <- msg

			wg.Add(1)
			go func(msg []byte) {
				defer wg.Done()

				select {
				case <-ctx.Done():
				case c.in <- msg:
				}
			}(msg)
		}
	}()

	<-ctx.Done()
	err = ln.Close()
	if err != nil {
		log.Printf("ERR: failed to stop listening: %v", err)
	}

	wg.Wait()
	log.Println("exiting udp server")

	return nil
}
