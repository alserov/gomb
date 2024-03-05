package server

import (
	"github.com/gorilla/websocket"
	"log/slog"
	"sync"
)

type hub struct {
	log *slog.Logger

	c map[string][]Consumer
	t map[string]Buffer

	mu sync.Mutex
}

func (h *hub) produce(conn *websocket.Conn, topic string, producerID string) {
	defer h.log.Info("producer disconnected", slog.String("id", producerID))

	for {
		t, msg, err := conn.ReadMessage()
		if err != nil || t == websocket.CloseMessage {
			return
		}

		h.t[topic] <- msg
	}
}

func (h *hub) consume(topic string) {
	for msg := range h.t[topic] {
		for _, cons := range h.c[topic] {
			if err := cons.conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
				return
			}
		}
	}
}

func (h *hub) addConsumer(conn *websocket.Conn, topic, consumerID string) {
	h.c[topic] = append(h.c[topic], Consumer{conn: conn, id: consumerID})
	go func() {
		for {
			t, _, err := conn.ReadMessage()
			if err != nil || t == websocket.CloseMessage {
				h.removeConsumer(topic, consumerID)
				if len(h.c[topic]) == 0 {
					delete(h.c, topic)
				}
				return
			}
		}
	}()
}

func (h *hub) removeConsumer(topic, consID string) {
	for i, c := range h.c[topic] {
		if c.id == consID {
			c.conn.Close()

			h.log.Info("consumer disconnected", slog.String("id", c.id))

			h.mu.Lock()
			h.c[topic] = append(h.c[topic][:i], h.c[topic][i+1:]...)
			h.mu.Unlock()

			return
		}
	}
}
