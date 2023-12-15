package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"wb-l0/internal/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	srv, err := server.Boot(ctx)
	if err != nil {
		log.Println(err)
		return
	}

	// Wait for interruption signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-c

	log.Println("shutting down!")
	cancel()
	srv.Shutdown(ctx)

	log.Println("shutdown complete")
}
