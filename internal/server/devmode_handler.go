package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/coder/websocket"
)

type devModeHandler struct {
	logger *slog.Logger
}

func (d devModeHandler) handle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			_, _ = w.Write([]byte("could not open websocket"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer conn.Close(websocket.StatusGoingAway, "server closed websocket")

		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		socketCtx := conn.CloseRead(ctx)
		for {
			_ = conn.Ping(socketCtx)
			time.Sleep(2 * time.Second)
		}
	}
}
