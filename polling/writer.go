package polling

import (
	"bytes"

	"github.com/googollee/go-engine.io/parser"
)

type writer struct {
	server *Polling
	code   parser.CodeType
	typ    parser.PacketType
	buf    *bytes.Buffer
}

func newWriter(server *Polling, code parser.CodeType, typ parser.PacketType) *writer {
	return &writer{
		server: server,
		code:   code,
		typ:    typ,
		buf:    bytes.NewBuffer(nil),
	}
}

func (w *writer) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

func (w *writer) Close() error {
	select {
	case w.server.sendChan <- true:
	default:
	}
	w.server.data = append(w.server.data, w.packet())
	return nil
}

func (w *writer) packet() parser.Packet {
	return parser.Packet{
		Code: w.code,
		Type: w.typ,
		Data: w.buf.Bytes(),
	}
}
