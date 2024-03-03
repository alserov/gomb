package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/alserov/gomb/server/internal/config"
	"github.com/gorilla/websocket"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"
)

func MustStart(cfg *config.Config) {
	defer func() {
		err := recover()
		if err != nil {
			log.Fatalf("panic recovery: %v", err)
		}
	}()

	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	l.Info("starting server", slog.String("addr", cfg.Addr))

	s := &http.Server{
		Addr: cfg.Addr,
		Handler: &handler{
			log:    l,
			topics: make(map[string]Buffer),
		},
	}

	chStop := make(chan os.Signal)
	signal.Notify(chStop, syscall.SIGINT, syscall.SIGTERM)

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

var (
	upgrader = websocket.Upgrader{
		WriteBufferSize: 1024,
		ReadBufferSize:  1024,
	}
)

type handler struct {
	topics map[string]Buffer
	log    *slog.Logger
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		u, _ := url.Parse(r.RequestURI)

		topic := path.Base(u.RequestURI())
		method := strings.Split(u.RequestURI(), "/")[1]

		switch method {
		case "produce":
			h.produce(w, r, topic)
		case "consume":
			h.consume(w, r, topic)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *handler) produce(w http.ResponseWriter, r *http.Request, topic string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, ok := h.topics[topic]; !ok {
		h.topics[topic] = make(Buffer, 128)
	}

	go func() {
		defer conn.Close()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				conn.WriteMessage(websocket.CloseMessage, []byte(fmt.Sprintf("failed to read message: %s", err.Error())))
				return
			}

			h.topics[topic] <- msg
		}
	}()
}

func (h *handler) consume(w http.ResponseWriter, r *http.Request, topic string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, ok := h.topics[topic]; !ok {
		fmt.Println(topic)
		conn.Close()
		return
	}

	go func() {
		defer conn.Close()
		for msg := range h.topics[topic] {
			if err = conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
				return
			}
		}
	}()
}
