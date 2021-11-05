package internal

import (
	"context"
	"net/http"
	"time"
)

type server struct {
	server *http.Server
}

func NewServer(addr string, handler http.Handler) *server {
	return &server{
		server: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
	}
}

func (s *server) Start() error {
	return s.server.ListenAndServe()
}

func (s *server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}
