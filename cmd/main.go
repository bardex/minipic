package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/bardex/minipic/internal/app"
	"github.com/bardex/minipic/internal/httpserver"
	"github.com/bardex/minipic/internal/httpserver/middleware"
	"github.com/bardex/minipic/internal/respcache"
)

func main() {
	configPath := flag.String("config", "/etc/minipic/config.toml", "Path to configuration file")
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

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
		h = middleware.NewCacheMiddleware(respcache.NewLruCache(cfg.Cache.Directory, cfg.Cache.Limit), h)
	}

	server := httpserver.NewServer(cfg.Server.Listen, h)
	fmt.Printf("listening on %s\n", cfg.Server.Listen)

	if err := server.Start(); err != nil {
		log.Fatalln(err)
	}
}
