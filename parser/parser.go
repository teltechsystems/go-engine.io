package parser

import (
	"encoding/base64"
	"fmt"
	"io"
)

const Protocol = 3

// MessageType is the type of message
type MessageType byte

const (
	MessageText   MessageType = '0'
	MessageBinary MessageType = 0
)

func (t MessageType) byte() byte {
	return byte(t)
}

func (t MessageType) String() string {
	switch t {
	case MessageText:
		return "text"
	case MessageBinary:
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
		return "message"
	case PacketUpgrade:
		return "upgrade"
	case PacketNoop:
		return "noop"
	}
	return fmt.Sprintf("unknow(0x%x)", byte(t))
}

// PacketEncoder is the encoder which encode the packet.
type PacketEncoder struct {
	io.WriteCloser
}

// NewEncoder return the encoder which encode type t to writer w.
func NewEncoder(w io.WriteCloser, pkg PacketType, msg MessageType) (*PacketEncoder, error) {
	t := pkg.byte() + msg.byte()
	if _, err := w.Write([]byte{t}); err != nil {
		return nil, err
	}
	return &PacketEncoder{
		WriteCloser: w,
	}, nil
}

// NewB64Encoder return the encoder which encode type t to writer w, as string. When write binary, it uses base64.
func NewB64Encoder(w io.WriteCloser, pkg PacketType) (*PacketEncoder, error) {
	_, err := w.Write([]byte{'b', pkg.byte() + '0'})
	if err != nil {
		return nil, err
	}
	base := base64.NewEncoder(base64.StdEncoding, w)
	return &PacketEncoder{
		WriteCloser: base,
	}, nil
}

// PacketDecoder is the decoder which decode data to packet.
type PacketDecoder struct {
	io.ReadCloser
	pkg PacketType
	msg MessageType
}

// NewDecoder return the decoder which decode from reader r.
func NewDecoder(r io.ReadCloser) (*PacketDecoder, error) {
	rc, msgType, pkgType, err := readType(r)
	if err != nil {
		return nil, err
	}

	ret := &PacketDecoder{
		ReadCloser: rc,
		pkg:        pkgType,
		msg:        msgType,
	}
	return ret, nil
}

type base64Decoder struct {
	io.Reader
	io.Closer
}

func readType(r io.ReadCloser) (io.ReadCloser, MessageType, PacketType, error) {
	b := []byte{0xff}
	if _, err := r.Read(b); err != nil {
		return nil, MessageText, PacketNoop, err
	}
	msgType := MessageText
	if b[0] == 'b' {
		if _, err := r.Read(b); err != nil {
			return nil, MessageText, PacketNoop, err
		}
		r = base64Decoder{
			Reader: base64.NewDecoder(base64.StdEncoding, r),
			Closer: r,
		}
		msgType = MessageBinary
	}
	if b[0] >= '0' {
		b[0] = b[0] - '0'
	} else {
		msgType = MessageBinary
	}
	pkgType, err := byteToPacketType(b[0])
	if err != nil {
		return nil, MessageText, packetMax, err
	}
	return r, msgType, pkgType, nil
}

// PacketType returns the type of packet.
func (d *PacketDecoder) PacketType() PacketType {
	return d.pkg
}

// MessageType returns the type of message, binary or text.
func (d *PacketDecoder) MessageType() MessageType {
	return d.msg
}
