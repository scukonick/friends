package main

import (
	"context"
	"log"

	"github.com/scukonick/friends/internal/server"
)

func main() {
	srv := server.NewServer()
	err := srv.ListenAndServe(context.Background(), "127.0.0.1:9090")
	if err != nil {
		log.Fatalln("failed to start server: ", err)
	}

}
