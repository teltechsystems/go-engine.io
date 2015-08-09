package websocket

import (
	"net/http"

	"github.com/googollee/go-engine.io/transport"
	"github.com/gorilla/websocket"
)

func NewCreater(upgrader *websocket.Upgrader) transport.Creater {
	newServer := func(w http.ResponseWriter, r *http.Request, callback transport.Callback) (transport.Server, error) {
		return NewServer(upgrader, w, r, callback)
	}
	return transport.Creater{
		Name:   "websocket",
		Server: newServer,
		Client: NewClient,
	}
}
