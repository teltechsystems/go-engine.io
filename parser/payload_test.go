package parser

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"github.com/googollee/go-assert"
)

type Test struct {
	name    string
	packets []Packet
	output  string
}

func doTest(t *testing.T, f func(io.Writer, []Packet) error, test Test) {
	buf := bytes.NewBuffer(nil)

	err := f(buf, test.packets)
	assert.MustEqual(t, err, nil)

	assert.Equal(t, buf.String(), test.output)

	decoder := NewPayloadDecoder(buf)

	for i := 0; ; i++ {
		d, err := decoder.Next()
		if err == io.EOF {
			break
		}
		assert.MustEqual(t, err, nil)
		assert.Equal(t, d.PacketType(), test.packets[i].Type)
		assert.Equal(t, d.CodeType(), test.packets[i].Code)

		if l := len(test.packets[i].Data); l > 0 {
			b, err := ioutil.ReadAll(d)
			assert.MustEqual(t, err, nil)
			assert.Equal(t, b, test.packets[i].Data)
			b, err = ioutil.ReadAll(d)
			assert.MustEqual(t, err, nil)
			assert.Equal(t, len(b), 0)
		}
		assert.MustEqual(t, err, nil)
	}
	assert.Equal(t, buf.Len(), 0)
}

func TestTextPayload(t *testing.T) {
	var tests = []Test{
		{"all in one", []Packet{
			Packet{CodeText, PacketOpen, nil},
			Packet{CodeText, PacketMessage, []byte("测试")},
			Packet{CodeBinary, PacketMessage, []byte("测试")},
		}, "\x31\x3a\x30\x37\x3a\x34\xe6\xb5\x8b\xe8\xaf\x95\x31\x30\x3a\x62\x34\x35\x72\x57\x4c\x36\x4b\x2b\x56"},
	}
	for _, test := range tests {
		encoder := EncodeToTextPayload
		doTest(t, encoder, test)
	}
}

func TestBinaryPayload(t *testing.T) {
	var tests = []Test{
		{"all in one", []Packet{
			Packet{CodeText, PacketOpen, nil},
			Packet{CodeText, PacketMessage, []byte("测试")},
			Packet{CodeBinary, PacketMessage, []byte("测试")},
		}, "\x00\x01\xff\x30\x00\x07\xff\x34\xe6\xb5\x8b\xe8\xaf\x95\x01\x07\xff\x04\xe6\xb5\x8b\xe8\xaf\x95"},
	}
	for _, test := range tests {
		encoder := EncodeToBinaryPayload
		doTest(t, encoder, test)
	}
}
