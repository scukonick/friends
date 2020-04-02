package main

import (
	"context"
	"log"

	"github.com/scukonick/friends/internal/bus"
	"github.com/scukonick/friends/internal/dispatcher"

	"github.com/scukonick/friends/internal/server"
)

func main() {
	localBus := bus.NewLocalBus()
	disp := dispatcher.NewBaseDispatcher(localBus)

	srv := server.NewServer(disp)
	err := srv.ListenAndServe(context.Background(), "127.0.0.1:9090")
	if err != nil {
		log.Fatalln("failed to start server: ", err)
	}

}
