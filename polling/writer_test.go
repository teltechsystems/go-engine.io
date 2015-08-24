package polling

import (
	"bytes"
	"io"
	"testing"

	"github.com/googollee/go-assert"
	"github.com/googollee/go-engine.io/parser"
)

func TestWriter(t *testing.T) {
	svr, err := NewServer(nil, nil)
	assert.MustEqual(t, err, nil)
	s, ok := svr.(*server)
	assert.MustEqual(t, ok, true)

	assert.Equal(t, len(s.data), 0)
	select {
	case <-s.sendChan:
		assert.MustEqual(t, "not here", "")
	default:
	}

	w := newWriter(s, parser.CodeText, parser.PacketMessage)
	_, err = w.Write([]byte("test"))
	assert.MustEqual(t, err, nil)
	err = w.Close()
	assert.MustEqual(t, err, nil)

	w = newWriter(s, parser.CodeBinary, parser.PacketOpen)
	err = w.Close()
	assert.MustEqual(t, err, nil)

	select {
	case <-s.sendChan:
	default:
		assert.MustEqual(t, "not here", "")
	}
	assert.Equal(t, len(s.data), 2)

	assert.Equal(t, s.data[0].Code, parser.CodeText)
	assert.Equal(t, s.data[0].Type, parser.PacketMessage)
	assert.Equal(t, s.data[0].Data, []byte("test"))

	assert.Equal(t, s.data[1].Code, parser.CodeBinary)
	assert.Equal(t, s.data[1].Type, parser.PacketOpen)
	assert.Equal(t, len(s.data[1].Data), 0)
}

type failWriter struct{}

func (w failWriter) Write(p []byte) (int, error) {
	return 0, io.EOF
}

func TestJSWriter(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	w := newJSWriter(buf)
	n, err := w.Write([]byte("abc><"))
	assert.MustEqual(t, err, nil)
	assert.MustEqual(t, w.Error(), nil)
	assert.Equal(t, n, 5)
	assert.Equal(t, buf.String(), `abc\x3E\x3C`)

	w = newJSWriter(failWriter{})
	n, err = w.Write([]byte("abc><"))
	assert.MustEqual(t, err, io.EOF)
	assert.MustEqual(t, w.Error(), io.EOF)
	n, err = w.Write([]byte("abc><"))
	assert.MustEqual(t, err, io.EOF)
	assert.MustEqual(t, w.Error(), io.EOF)
}
