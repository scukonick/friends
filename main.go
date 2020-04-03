package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/scukonick/friends/internal/bus"
	"github.com/scukonick/friends/internal/dispatcher"

	"github.com/scukonick/friends/internal/server"
)

func main() {
	localBus := bus.NewLocalBus()
	disp := dispatcher.NewBaseDispatcher(localBus)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		tcpSrv := server.NewTCPServer(disp)
		err := tcpSrv.ListenAndServe(ctx, "127.0.0.1:9090")
		if err != nil {
			log.Fatalln("failed to start tcp server: ", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		tcpSrv := server.NewUDPServer(disp)
		err := tcpSrv.ListenAndServe(ctx, "127.0.0.1:9090")
		if err != nil {
			log.Fatalln("failed to start udp server: ", err)
		}
	}()

	<-exitHandler()
	log.Println("received exit signal")
	cancel()
	wg.Wait()
	log.Println("exiting")
}

func exitHandler() <-chan os.Signal {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	return c
}
