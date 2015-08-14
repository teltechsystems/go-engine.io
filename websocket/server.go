package websocket

import (
	"net/http"

	"github.com/googollee/go-engine.io/transport"
	"github.com/gorilla/websocket"
)

type Server struct {
	websocketBase
}

func NewServer(upgrader *websocket.Upgrader, w http.ResponseWriter, r *http.Request) (transport.Server, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	ret := &Server{
		websocketBase: websocketBase{
			conn: conn,
		},
	}

	return ret, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}
