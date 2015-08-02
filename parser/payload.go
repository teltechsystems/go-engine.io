package parser

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
)

var ErrInvalidMessageType = errors.New("invalid message type")

// PayloadEncoder is the encoder to encode packets as payload.
type PayloadEncoder struct {
	buffers [][]byte
	isText  bool
}

// NewTextPayloadEncoder returns the encoder which encode as string.
func NewTextPayloadEncoder() *PayloadEncoder {
	return &PayloadEncoder{
		isText: true,
	}
}

// NewBinaryPayloadEncoder returns the encoder which encode as binary.
func NewBinaryPayloadEncoder() *PayloadEncoder {
	return &PayloadEncoder{
		isText: false,
	}
}

type encoder struct {
	*PacketEncoder
	buf          *buffer
	binaryPrefix string
	payload      *PayloadEncoder
}

func (e encoder) Close() error {
	if err := e.PacketEncoder.Close(); err != nil {
		return err
	}
	var buffer []byte
	if e.payload.isText {
		buffer = []byte(fmt.Sprintf("%d:%s", e.buf.Len(), e.buf.String()))
	} else {
		buffer = []byte(fmt.Sprintf("%s%d", e.binaryPrefix, e.buf.Len()))
		for i, n := 0, len(buffer); i < n; i++ {
			buffer[i] = buffer[i] - '0'
		}
		buffer = append(buffer, 0xff)
		buffer = append(buffer, e.buf.Bytes()...)
	}

	e.payload.buffers = append(e.payload.buffers, buffer)

	return nil
}

func (e *PayloadEncoder) nextText(t PacketType) (io.WriteCloser, error) {
	buf := newBuffer()
	pEncoder, err := NewEncoder(buf, t, MessageText)
	if err != nil {
		return nil, err
	}
	return encoder{
		PacketEncoder: pEncoder,
		buf:           buf,
		binaryPrefix:  "0",
		payload:       e,
	}, nil
}

func (e *PayloadEncoder) nextBinary(t PacketType) (io.WriteCloser, error) {
	buf := newBuffer()
	var pEncoder *PacketEncoder
	var err error
	if e.isText {
		pEncoder, err = NewB64Encoder(buf, t)
	} else {
		pEncoder, err = NewEncoder(buf, t, MessageBinary)
	}
	if err != nil {
		return nil, err
	}
	return encoder{
		PacketEncoder: pEncoder,
		buf:           buf,
		binaryPrefix:  "1",
		payload:       e,
	}, nil
}

// Next returns next writer.
func (e *PayloadEncoder) Next(pkg PacketType, msg MessageType) (io.WriteCloser, error) {
	switch msg {
	case MessageBinary:
		return e.nextBinary(pkg)
	case MessageText:
		return e.nextText(pkg)
	}
	return nil, ErrInvalidMessageType
}

// EncodeTo writes encoded payload to writer w. It will clear the buffer of encoder.
func (e *PayloadEncoder) EncodeTo(w io.Writer) error {
	for _, b := range e.buffers {
		_, err := io.Copy(w, bytes.NewReader(b))
		if err != nil {
			return err
		}
	}
	return nil
}

//IsText returns true if payload encode to text, otherwise returns false.
func (e *PayloadEncoder) IsText() bool {
	return e.isText
}

// PayloadDecoder is the decoder to decode payload.
type PayloadDecoder struct {
	r *bufio.Reader
}

// NewPayloadDecoder returns the payload decoder which read from reader r.
func NewPayloadDecoder(r io.Reader) *PayloadDecoder {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	return &PayloadDecoder{
		r: br,
	}
}

// Next returns the packet decoder. Make sure it will be closed after used.
func (d *PayloadDecoder) Next() (*PacketDecoder, error) {
	firstByte, err := d.r.Peek(1)
	if err != nil {
		return nil, err
	}
	isBinary := firstByte[0] < '0'
	delim := byte(':')
	if isBinary {
		d.r.ReadByte()
		delim = 0xff
	}
	line, err := d.r.ReadBytes(delim)
	if err != nil {
		return nil, err
	}
	l := len(line)
	if l < 1 {
		return nil, fmt.Errorf("invalid input")
	}
	lenByte := line[:l-1]
	if isBinary {
		for i, n := 0, l; i < n; i++ {
			line[i] = line[i] + '0'
		}
	}
	packetLen, err := strconv.ParseInt(string(lenByte), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid input")
	}
	return NewDecoder(newLimitReader(d.r, packetLen))
}

type buffer struct {
	*bytes.Buffer
}

func newBuffer() *buffer {
	return &buffer{
		Buffer: bytes.NewBuffer(nil),
	}
}

func (b *buffer) Close() error {
	return nil
}
