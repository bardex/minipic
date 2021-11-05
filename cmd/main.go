package main

import (
	"flag"
	"log"

	"github.com/bardex/minipic/pkg/resp_cache"

	"github.com/bardex/minipic/internal"
)

func main() {
	configPath := flag.String("config", "/etc/minipic/config.toml", "Path to configuration file")
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	cfg, err := NewConfig(*configPath)
	if err != nil {
		log.Fatalf("Fail loading configuration:%s", err)
	}

	h := internal.NewHandler(
		internal.NewImageDownloader(),
		internal.ResizerByImaging{},
	)

	if cfg.Cache.Limit > 0 {
		h = resp_cache.NewCacheMiddleware(resp_cache.NewLruCache(cfg.Cache.Directory, cfg.Cache.Limit), h)
	}

	server := internal.NewServer(cfg.Server.Listen, h)
	log.Println("listening " + cfg.Server.Listen)

	if err := server.Start(); err != nil {
		log.Fatalln(err)
	}
}
