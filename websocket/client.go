package websocket

import (
	"net/http"

	"github.com/googollee/go-engine.io/transport"
	"github.com/gorilla/websocket"
)

type client struct {
	websocketBase
	resp *http.Response
}

func NewClient(r *http.Request) (transport.Client, error) {
	dialer := websocket.DefaultDialer

	conn, resp, err := dialer.Dial(r.URL.String(), r.Header)
	if err != nil {
		return nil, err
	}

	return &client{
		websocketBase: websocketBase{
			conn: conn,
		},
		resp: resp,
	}, nil
}

func (c *client) Response() *http.Response {
	return c.resp
}
