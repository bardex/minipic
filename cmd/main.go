package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bardex/minipic/internal/app"
	"github.com/bardex/minipic/internal/httpserver"
	"github.com/bardex/minipic/internal/httpserver/middleware"
)

var (
	release   = "UNKNOWN"
	buildDate = "UNKNOWN"
	gitHash   = "UNKNOWN"
)

func main() {
	configPath := flag.String("config", "/etc/minipic/config.toml", "Path to configuration file")
	flag.Parse()

	fmt.Printf("Minipic ver: %s %s %s\n", release, buildDate, gitHash)

	cfg, err := NewConfig(*configPath)
	if err != nil {
		log.Fatalf("Fail loading configuration:%s", err)
	}

	h := httpserver.NewHandler(
		app.NewImageDownloader(),
		app.Resizer{},
	)

	if cfg.Cache.Limit > 0 {
		h = middleware.NewCache(app.NewLruCache(cfg.Cache.Directory, cfg.Cache.Limit), h)
	}

	server := httpserver.NewServer(cfg.Server.Listen, h)

	go func() {
		log.Printf("listening on %s...\n", cfg.Server.Listen)
		if err := server.Start(); err != nil {
			log.Fatalln(err)
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-done

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Println("shutdown server...")
	if err := server.Stop(ctx); err != nil {
		log.Println("failed to stop http server: " + err.Error())
	}
}
