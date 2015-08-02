package parser

import (
	"bytes"
	"io"
	"testing"

	"github.com/googollee/go-assert"
)

type packet struct {
	pkg  PacketType
	msg  MessageType
	data []byte
}
type Test struct {
	name    string
	packets []packet
	output  string
}

type iEncoder interface {
	Next(pkg PacketType, msg MessageType) (io.WriteCloser, error)
	EncodeTo(w io.Writer) error
	IsText() bool
}

func doTest(t *testing.T, encoder iEncoder, test Test) {
	buf := bytes.NewBuffer(nil)

	for _, p := range test.packets {
		w, err := encoder.Next(p.pkg, p.msg)
		assert.MustEqual(t, err, nil)
		n, err := w.Write(p.data)
		assert.MustEqual(t, err, nil)
		assert.Equal(t, n, len(p.data))
		err = w.Close()
		assert.MustEqual(t, err, nil)
	}

	err := encoder.EncodeTo(buf)
	assert.MustEqual(t, err, nil)
	assert.Equal(t, buf.String(), test.output)

	decoder := NewPayloadDecoder(buf)

	for i := 0; ; i++ {
		d, err := decoder.Next()
		if err == io.EOF {
			break
		}
		assert.MustEqual(t, err, nil)
		assert.Equal(t, d.PacketType(), test.packets[i].pkg)
		assert.Equal(t, d.MessageType(), test.packets[i].msg)

		if l := len(test.packets[i].data); l > 0 {
			buf := make([]byte, len(test.packets[i].data)+1)
			n, err := d.Read(buf)
			assert.MustEqual(t, err, nil)
			assert.Equal(t, buf[:n], test.packets[i].data)
			_, err = d.Read(buf)
			assert.MustEqual(t, err, io.EOF)
		}
		err = d.Close()
		assert.MustEqual(t, err, nil)
	}
	assert.Equal(t, buf.Len(), 0)
}

func TestStringPayload(t *testing.T) {
	var tests = []Test{
		{"all in one", []packet{
			packet{PacketOpen, MessageText, nil},
			packet{PacketMessage, MessageText, []byte("测试")},
			packet{PacketMessage, MessageBinary, []byte("测试")},
		}, "\x31\x3a\x30\x37\x3a\x34\xe6\xb5\x8b\xe8\xaf\x95\x31\x30\x3a\x62\x34\x35\x72\x57\x4c\x36\x4b\x2b\x56"},
	}
	for _, test := range tests {
		encoder := NewTextPayloadEncoder()
		assert.MustEqual(t, encoder.IsText(), true)
		doTest(t, encoder, test)
	}
}

func TestBinaryPayload(t *testing.T) {
	var tests = []Test{
		{"all in one", []packet{
			packet{PacketOpen, MessageText, nil},
			packet{PacketMessage, MessageText, []byte("测试")},
			packet{PacketMessage, MessageBinary, []byte("测试")},
		}, "\x00\x01\xff\x30\x00\x07\xff\x34\xe6\xb5\x8b\xe8\xaf\x95\x01\x07\xff\x04\xe6\xb5\x8b\xe8\xaf\x95"},
	}
	for _, test := range tests {
		encoder := NewBinaryPayloadEncoder()
		assert.MustEqual(t, encoder.IsText(), false)
		doTest(t, encoder, test)
	}
}
