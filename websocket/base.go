package websocket

import (
	"io"
	"io/ioutil"
	"time"

	"github.com/googollee/go-engine.io/parser"
	"github.com/gorilla/websocket"
)

type websocketBase struct {
	conn *websocket.Conn
}

type writeCloser struct {
	io.Writer
	io.Closer
}

func (b websocketBase) NextWriter(code parser.CodeType, typ parser.PacketType) (io.WriteCloser, error) {
	wsType := websocket.TextMessage
	if code == parser.CodeBinary {
		wsType = websocket.BinaryMessage
	}

	w, err := b.conn.NextWriter(wsType)
	if err != nil {
		return nil, err
	}
	ret, err := parser.NewEncoder(w, typ, code)
	if err != nil {
		return nil, err
	}
	return writeCloser{
		Writer: ret,
		Closer: w,
	}, nil
}

func (b websocketBase) NextReader() (parser.CodeType, parser.PacketType, io.ReadCloser, error) {
	for {
		t, r, err := b.conn.NextReader()
		if err != nil {
			return 0, 0, nil, err
		}

		switch t {
		case websocket.TextMessage:
			fallthrough
		case websocket.BinaryMessage:
			decoder, err := parser.NewDecoder(r)
			if err != nil {
				return 0, 0, nil, err
			}
			return decoder.CodeType(), decoder.PacketType(), ioutil.NopCloser(decoder), nil
		}
	}
}

func (b websocketBase) Close() error {
	return b.conn.Close()
}

func (b websocketBase) SetReadDeadline(t time.Time) error {
	return b.conn.SetReadDeadline(t)
}

func (b websocketBase) SetWriteDeadline(t time.Time) error {
	return b.conn.SetWriteDeadline(t)
}
