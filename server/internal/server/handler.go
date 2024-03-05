package server

import (
	"github.com/gorilla/websocket"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strings"
)

var (
	upgrader = websocket.Upgrader{
		WriteBufferSize: 1024,
		ReadBufferSize:  1024,
	}
)

type handler struct {
	hub hub
	log *slog.Logger
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

	producerID := r.Header.Get("X-Request-ID")

	h.log.Info("producer connected", slog.String("id", producerID))

	if _, ok := h.hub.t[topic]; !ok {
		h.hub.t[topic] = make(Buffer, 32)
	}

	go h.hub.produce(conn, topic, producerID)
}

func (h *handler) consume(w http.ResponseWriter, r *http.Request, topic string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	consumerID := r.Header.Get("X-Request-ID")

	h.log.Info("consumer connected", slog.String("id", consumerID))

	if _, ok := h.hub.t[topic]; !ok {
		conn.Close()
		return
	}

	if _, ok := h.hub.c[topic]; !ok {
		go h.hub.consume(topic)
	}

	h.hub.addConsumer(conn, topic, consumerID)
}
