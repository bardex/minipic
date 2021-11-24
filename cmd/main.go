package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/bardex/minipic/internal"
	"github.com/bardex/minipic/pkg/respcache"
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

	h := internal.NewHandler(
		internal.NewImageDownloader(),
		internal.ResizerByImaging{},
	)

	if cfg.Cache.Limit > 0 {
		h = respcache.NewCacheMiddleware(respcache.NewLruCache(cfg.Cache.Directory, cfg.Cache.Limit), h)
	}

	server := internal.NewServer(cfg.Server.Listen, h)
	fmt.Printf("listening on %s\n", cfg.Server.Listen)

	if err := server.Start(); err != nil {
		log.Fatalln(err)
	}
}
