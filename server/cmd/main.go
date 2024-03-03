package main

import (
	"github.com/alserov/gomb/server/internal/config"
	"github.com/alserov/gomb/server/internal/server"
)

func main() {
	cfg := config.MustLoad()
	server.MustStart(cfg)
}
