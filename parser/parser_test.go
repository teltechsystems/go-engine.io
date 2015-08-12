package parser

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/googollee/go-assert"
)

func TestPacketType(t *testing.T) {
	type Test struct {
		byt byte
		typ PacketType
		ok  bool
	}

	tests := []Test{
		{0, PacketOpen, true},
		{1, PacketClose, true},
		{2, PacketPing, true},
		{3, PacketPong, true},
		{4, PacketMessage, true},
		{5, PacketUpgrade, true},
		{6, PacketNoop, true},
		{7, PacketNoop, false},
	}
	for _, test := range tests {
		typ, err := byteToPacketType(test.byt)
		assert.MustEqual(t, err == nil, test.ok, "test: %v", test.byt)
		if err != nil {
			continue
		}
		assert.Equal(t, typ, test.typ)
		assert.Equal(t, test.typ.byte(), test.byt)
	}
}

func TestParser(t *testing.T) {
	type Test struct {
		name   string
		pkg    PacketType
		msg    CodeType
		data   []byte
		output string
		ok     bool
	}

	tests := []Test{
		{"without data", PacketOpen, CodeText, nil, "0", true},
		{"with data", PacketMessage, CodeText, []byte("测试"), "4\xe6\xb5\x8b\xe8\xaf\x95", true},
		{"without data", PacketOpen, CodeBinary, nil, "\x00", true},
		{"with data", PacketMessage, CodeBinary, []byte("测试"), "\x04\xe6\xb5\x8b\xe8\xaf\x95", true},
	}
	for _, test := range tests {
		buf := newBuffer()

		// Create encoder
		encoder, err := NewEncoder(buf, test.pkg, test.msg)
		assert.MustEqual(t, err == nil, test.ok, "test: %s", test.name)
		if err != nil {
			continue
		}
		var _ io.Writer = encoder

		// Encode
		n, err := encoder.Write(test.data)
		assert.MustEqual(t, err, nil, "test: %s", test.name)
		assert.Equal(t, n, len(test.data), "test: %s", test.name)

		// Create decoder
		decoder, err := NewDecoder(buf)
		assert.MustEqual(t, err == nil, test.ok, "test: %s", test.name)
		if err != nil {
			continue
		}
		var _ io.Reader = decoder

		// Decode
		assert.MustEqual(t, decoder.PacketType(), test.pkg, "test: %s", test.name)
		assert.MustEqual(t, decoder.CodeType(), test.msg, "test: %s", test.name)
		decoded := make([]byte, len(test.data)+1)
		if len(test.data) > 0 {
			n, err := decoder.Read(decoded)
			assert.MustEqual(t, err, nil, "test: %s", test.name)
			assert.Equal(t, n, len(test.data), "test: %s", test.name)
		}

		// EOF
		_, err = decoder.Read(decoded[:])
		assert.MustEqual(t, err, io.EOF, "test: %s", test.name)
	}
}

func TestBase64Parser(t *testing.T) {
	type Test struct {
		name   string
		pkg    PacketType
		data   []byte
		output string
		ok     bool
	}

	tests := []Test{
		{"without data", PacketOpen, nil, "b0", true},
		{"with data", PacketMessage, []byte("测试"), "b45rWL6K+V", true},
		{"with text data", PacketMessage, []byte("test"), "b4dGVzdA==", true},
	}
	for _, test := range tests {
		buf := newBuffer()

		// Create encoder
		encoder, err := NewB64Encoder(buf, test.pkg)
		assert.MustEqual(t, err == nil, test.ok, "test: %s", test.name)
		if err != nil {
			continue
		}
		var _ io.Writer = encoder

		// Encode
		n, err := encoder.Write(test.data)
		assert.MustEqual(t, err, nil, "test: %s", test.name)
		assert.Equal(t, n, len(test.data), "test: %s", test.name)

		// Close
		err = encoder.Close()
		assert.MustEqual(t, err, nil, "test: %s", test.name)

		// Create decoder
		decoder, err := NewDecoder(buf)
		assert.MustEqual(t, err == nil, test.ok, "test: %s", test.name)
		if err != nil {
			continue
		}
		var _ io.Reader = decoder

		// Decode
		assert.MustEqual(t, decoder.PacketType(), test.pkg, "test: %s", test.name)
		assert.MustEqual(t, decoder.CodeType(), CodeBinary, "test: %s", test.name)
		b, err := ioutil.ReadAll(decoder)
		assert.MustEqual(t, err, nil, "test: %s", test.name)
		assert.Equal(t, len(b), len(test.data), "test: %s", test.name)
		if len(test.data) > 0 {
			assert.Equal(t, b, test.data, "test: %s", test.name)
		}
	}
}
