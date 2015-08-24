package polling

import (
	"bytes"
	"io"
	"text/template"

	"github.com/googollee/go-engine.io/parser"
)

type jsWriter struct {
	errWriter
}

func newJSWriter(w io.Writer) *jsWriter {
	return &jsWriter{
		errWriter: errWriter{
			w: w,
		},
	}
}

func (w *jsWriter) Write(p []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	template.JSEscape(&w.errWriter, p)
	if w.err != nil {
		return 0, w.err
	}
	return len(p), nil
}

func (w *jsWriter) Error() error {
	return w.err
}

type errWriter struct {
	w   io.Writer
	err error
}

func (w *errWriter) Write(p []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	n, err := w.w.Write(p)
	if err != nil {
		w.err = err
	}
	return n, err
}

type writer struct {
	server *server
	code   parser.CodeType
	typ    parser.PacketType
	buf    *bytes.Buffer
}

func newWriter(server *server, code parser.CodeType, typ parser.PacketType) *writer {
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
