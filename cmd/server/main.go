package main

import (
	"log"
	"os"

	"github.com/KasumiMercury/tolo-relation-stub/internal/config"
	"github.com/KasumiMercury/tolo-relation-stub/internal/server"
	"github.com/KasumiMercury/tolo-relation-stub/internal/store"
)

func main() {
	cfg, err := config.Load(os.Args[1:])
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	s, err := store.LoadFromFile(cfg.DataPath)
	if err != nil {
		log.Fatalf("load seed: %v", err)
	}

	r := server.NewRouter(s)

	log.Printf("relation-stub listening on %s (data=%s)", cfg.Addr, cfg.DataPath)
	if err := r.Run(cfg.Addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
