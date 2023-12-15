package server

import (
	"context"
	"fmt"
	"log"

	"wb-l0/internal/config/env"
)

func Boot(ctx context.Context) (*Server, error) {
	cfg := env.ReadConfig()
	server, err := NewServerFromConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("error initializing server: %s", err)
	}

	server.Run(ctx)
	log.Println("completed initialization")
	return server, nil
}
