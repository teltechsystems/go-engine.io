package websocket

import (
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/googollee/go-engine.io/parser"
	"github.com/googollee/go-engine.io/transport"
	"github.com/gorilla/websocket"
)

type Server struct {
	conn *websocket.Conn
}

func NewServer(upgrader *websocket.Upgrader, w http.ResponseWriter, r *http.Request) (transport.Server, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	ret := &Server{
		conn: conn,
	}

	return ret, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func (s *Server) NextWriter(msg parser.MessageType, pkg parser.PacketType) (io.WriteCloser, error) {
	wsType := websocket.TextMessage
	if msg == parser.MessageBinary {
		wsType = websocket.BinaryMessage
	}

	w, err := s.conn.NextWriter(wsType)
	if err != nil {
		return nil, err
	}
	ret, err := parser.NewEncoder(w, pkg, msg)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (s Server) NextReader() (*parser.PacketDecoder, error) {
	for {
		t, r, err := s.conn.NextReader()
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

func (s *Server) Close() error {
	return s.conn.Close()
}

func (s *Server) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

func (s *Server) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

func (s *Server) SetReadDeadline(t time.Time) error {
	return s.conn.SetReadDeadline(t)
}

func (s *Server) SetWriteDeadline(t time.Time) error {
	return s.conn.SetWriteDeadline(t)
}
