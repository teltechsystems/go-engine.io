package websocket

import (
	"io"
	"io/ioutil"
	"net/http"

	"github.com/googollee/go-engine.io/parser"
	"github.com/googollee/go-engine.io/transport"
	"github.com/gorilla/websocket"
)

type client struct {
	conn *websocket.Conn
	resp *http.Response
}

func NewClient(r *http.Request) (transport.Client, error) {
	dialer := websocket.DefaultDialer

	conn, resp, err := dialer.Dial(r.URL.String(), r.Header)
	if err != nil {
		return nil, err
	}

	return &client{
		conn: conn,
		resp: resp,
	}, nil
}

func (c *client) Response() *http.Response {
	return c.resp
}

func (c *client) NextReader() (*parser.PacketDecoder, error) {
	for {
		t, r, err := c.conn.NextReader()
		if err != nil {
			return nil, err
		}
		switch t {
		case websocket.TextMessage:
			fallthrough
		case websocket.BinaryMessage:
			return parser.NewDecoder(ioutil.NopCloser(r))
		}
	}
}

func (c *client) NextWriter(msg parser.MessageType, pkg parser.PacketType) (io.WriteCloser, error) {
	wsType := websocket.TextMessage
	if msg == parser.MessageBinary {
		wsType = websocket.BinaryMessage
	}

	w, err := c.conn.NextWriter(wsType)
	if err != nil {
		return nil, err
	}
	ret, err := parser.NewEncoder(w, pkg, msg)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *client) Close() error {
	return c.conn.Close()
}
