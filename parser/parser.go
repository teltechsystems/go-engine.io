package parser

import (
	"encoding/base64"
	"fmt"
	"io"
)

const Protocol = 3

// CodeType is the type of Code
type CodeType byte

const (
	CodeText   CodeType = '0'
	CodeBinary CodeType = 0
)

func (t CodeType) byte() byte {
	return byte(t)
}

func (t CodeType) String() string {
	switch t {
	case CodeText:
		return "text"
	case CodeBinary:
		return "binary"
	}
	return fmt.Sprintf("unknow(0x%x)", byte(t))
}

// PacketType is the type of packet
type PacketType byte

const (
	PacketOpen PacketType = iota
	PacketClose
	PacketPing
	PacketPong
	PacketMessage
	PacketUpgrade
	PacketNoop
	packetMax
)

func byteToPacketType(b byte) (PacketType, error) {
	ret := PacketType(b)
	if ret < packetMax {
		return ret, nil
	}
	return PacketNoop, fmt.Errorf("invalid byte 0x%x", b)
}

// Byte return the byte of type
func (t PacketType) byte() byte {
	return byte(t)
}

// String return string
func (t PacketType) String() string {
	switch t {
	case PacketOpen:
		return "open"
	case PacketClose:
		return "close"
	case PacketPing:
		return "ping"
	case PacketPong:
		return "pong"
	case PacketMessage:
		return "Code"
	case PacketUpgrade:
		return "upgrade"
	case PacketNoop:
		return "noop"
	}
	return fmt.Sprintf("unknow(0x%x)", byte(t))
}

// NewEncoder return the writer which encode type t to writer w.
func NewEncoder(w io.Writer, typ PacketType, code CodeType) (io.Writer, error) {
	_, err := w.Write([]byte{typ.byte() + code.byte()})
	if err != nil {
		return nil, err
	}
	return w, nil
}

func normalEncodeLength(n int) int {
	return 1 + n
}

// NewB64Encoder return the writeCloser which encode type t to writer w, as string. When write binary, it uses base64.
func NewB64Encoder(w io.Writer, typ PacketType) (io.WriteCloser, error) {
	_, err := w.Write([]byte{'b', typ.byte() + '0'})
	if err != nil {
		return nil, err
	}
	base := base64.NewEncoder(base64.StdEncoding, w)
	return base, nil
}

func base64EncodeLength(n int) int {
	return 2 + base64.StdEncoding.EncodedLen(n)
}

// PacketDecoder is the decoder which decode data to packet.
type PacketDecoder struct {
	io.Reader
	typ  PacketType
	code CodeType
}

// NewDecoder return the decoder which decode from reader r.
func NewDecoder(r io.Reader) (*PacketDecoder, error) {
	r, msgType, pkgType, err := readType(r)
	if err != nil {
		return nil, err
	}

	ret := &PacketDecoder{
		Reader: r,
		typ:    pkgType,
		code:   msgType,
	}
	return ret, nil
}

func readType(r io.Reader) (io.Reader, CodeType, PacketType, error) {
	var b [1]byte
	if _, err := r.Read(b[:]); err != nil {
		return nil, CodeText, PacketNoop, err
	}
	msgType := CodeText
	if b[0] == 'b' {
		if _, err := r.Read(b[:]); err != nil {
			return nil, CodeText, PacketNoop, err
		}
		r = base64.NewDecoder(base64.StdEncoding, r)
		msgType = CodeBinary
	}
	if b[0] >= '0' {
		b[0] = b[0] - '0'
	} else {
		msgType = CodeBinary
	}
	pkgType, err := byteToPacketType(b[0])
	if err != nil {
		return nil, CodeText, packetMax, err
	}
	return r, msgType, pkgType, nil
}

// PacketType returns the type of packet.
func (d *PacketDecoder) PacketType() PacketType {
	return d.typ
}

// CodeType returns the type of Code, binary or text.
func (d *PacketDecoder) CodeType() CodeType {
	return d.code
}
