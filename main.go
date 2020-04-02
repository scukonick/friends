package main

import (
	"context"
	"log"
	"sync"

	"github.com/scukonick/friends/internal/bus"
	"github.com/scukonick/friends/internal/dispatcher"

	"github.com/scukonick/friends/internal/server"
)

func main() {
	localBus := bus.NewLocalBus()
	disp := dispatcher.NewBaseDispatcher(localBus)

	ctx := context.Background()

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

	wg.Wait()
}
