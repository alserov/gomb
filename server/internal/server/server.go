package server

import (
	"context"
	"errors"
	"github.com/alserov/gomb/server/internal/config"
	"github.com/alserov/gomb/server/internal/healthchecker"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func MustStart(cfg *config.Config) {
	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	defer func() {
		err := recover()
		if err != nil {
			l.Error("panic recovery", slog.Any("panic", err))
		}
	}()

	l.Info("starting server", slog.String("addr", cfg.Addr))
	if len(cfg.PredefinedTopics) > 0 {
		l.Info("predefined topics", slog.Int("amount", len(cfg.PredefinedTopics)))
	}

	topics := make(map[string]Buffer)
	for topic, bufferSize := range cfg.PredefinedTopics {
		topics[topic] = make(Buffer, bufferSize)
	}

	s := &http.Server{
		Addr: cfg.Addr,
		Handler: &handler{
			hub: hub{
				t:   topics,
				c:   make(map[string][]Consumer),
				log: l,
			},
			log: l,
		},
	}

	chStop := make(chan os.Signal)
	signal.Notify(chStop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		healthchecker.PrintStat(l)
	}()

	go func() {
		l.Info("server is ready")
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic("failed to serve: " + err.Error())
		}
	}()

	<-chStop
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	err := s.Shutdown(ctx)
	if err != nil {
		panic("failed to shutdown server: " + err.Error())
	}
}
