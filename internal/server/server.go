package server

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"

	"wb-l0/internal/config"
	"wb-l0/pkg/httperrors"

	"wb-l0/internal/order"
	"wb-l0/internal/order/consumer"
	orderHttp "wb-l0/internal/order/delivery/http"
	orderRepository "wb-l0/internal/order/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	orderRepository  order.Repository
	orderConsumer    order.Consumer
	serverConfig     config.Server
	shutdownComplete chan struct{}
}

func NewServer(
	orderRepository order.Repository,
	orderConsumer order.Consumer,
	serverConfig config.Server,
) *Server {
	return &Server{
		orderRepository: orderRepository,
		orderConsumer:   orderConsumer,
		serverConfig:    serverConfig,
	}
}

func NewServerFromConfig(ctx context.Context, cfg *config.Config) (*Server, error) {
	primaryDatabase, err := orderRepository.NewPostgresRepositoryFromConfig(ctx, cfg.Postgres)
	if err != nil {
		return nil, err
	}

	cache := orderRepository.NewInMemoryRepository()
	orderRepo := orderRepository.NewCachedRepository(primaryDatabase, cache)

	orderConsumer, err := consumer.NewConsumer(cfg.Nats, orderRepo)
	if err != nil {
		return nil, err
	}

	return NewServer(orderRepo, orderConsumer, cfg.Server), nil
}

func (s *Server) Run(ctx context.Context) {
	s.startWebServer(ctx)
	s.startNatsConsumer(ctx)
}

func (s *Server) Shutdown(ctx context.Context) {
	select {
	case <-s.shutdownComplete:
		log.Println("http server is down")
	case <-ctx.Done():
		log.Println("context deadline exceeded")
	}
}

func (s *Server) startNatsConsumer(ctx context.Context) {
	go func() {
		err := s.orderConsumer.Subscribe(ctx)
		if err != nil {
			log.Println("error starting nats consumer:", err)
			return
		}

		log.Println("started nats consumer")
	}()
}

func (s *Server) startWebServer(ctx context.Context) {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Route("/orders", func(router chi.Router) {
		handler := orderHttp.NewOrderHandler(s.orderRepository)
		router.Get("/{id}", WrapHandler(handler.GetOrder))
	})

	router.NotFound(ErrorHandler(httperrors.ErrNotFound))
	router.MethodNotAllowed(ErrorHandler(httperrors.ErrMethodNotAllowed))

	server := &http.Server{
		Addr:    s.serverConfig.BindAddress,
		Handler: router,
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}

	s.shutdownComplete = make(chan struct{})

	go func() {
		log.Println("starting http server on", s.serverConfig.BindAddress)
		err := server.ListenAndServe()
		if err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Println("http server error:", err)
			}

			return
		}
	}()

	go func() {
		<-ctx.Done()
		defer close(s.shutdownComplete)

		log.Println("shutting down http server...")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Println("error shutting down http server:", err)
		}
	}()
}
