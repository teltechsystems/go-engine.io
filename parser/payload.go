package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
)

// Packet is information of one packet
type Packet struct {
	Code CodeType
	Type PacketType
	Data []byte
}

type payloadEncoder interface {
	Encode(Packet) error
}

// EncodeToBinaryPayload encodes packets to w, using binary code.
func EncodeToBinaryPayload(w io.Writer, packets []Packet) error {
	encoder := newBinaryPayloadEncoder(w)
	return encodeToPayload(encoder, packets)
}

// EncodeToTextPayload encodes packets to w, using text code.
func EncodeToTextPayload(w io.Writer, packets []Packet) error {
	encoder := newTextPayloadEncoder(w)
	return encodeToPayload(encoder, packets)
}

func encodeToPayload(encoder payloadEncoder, packets []Packet) error {
	for i := range packets {
		if err := encoder.Encode(packets[i]); err != nil {
			return err
		}
	}
	return nil
}

type textPayloadEncoder struct {
	w io.Writer
}

func newTextPayloadEncoder(w io.Writer) payloadEncoder {
	return &textPayloadEncoder{
		w: w,
	}
}

func (e *textPayloadEncoder) Encode(p Packet) error {
	length := normalEncodeLength
	if p.Code == CodeBinary {
		length = base64EncodeLength
	}
	prefix := fmt.Sprintf("%d:", length(len(p.Data)))
	if _, err := io.WriteString(e.w, prefix); err != nil {
		return err
	}
	var encoder io.Writer
	var err error
	if p.Code == CodeBinary {
		encoder, err = NewB64Encoder(e.w, p.Type)
	} else {
		encoder, err = NewEncoder(e.w, p.Type, p.Code)
	}
	if err != nil {
		return err
	}
	if _, err := encoder.Write(p.Data); err != nil {
		return err
	}
	if c, ok := encoder.(io.Closer); ok {
		if err := c.Close(); err != nil {
			return err
		}
	}
	return nil
}

type binaryPayloadEncoder struct {
	w io.Writer
}

func newBinaryPayloadEncoder(w io.Writer) payloadEncoder {
	return &binaryPayloadEncoder{
		w: w,
	}
}

func (e *binaryPayloadEncoder) Encode(p Packet) error {
	l := normalEncodeLength(len(p.Data))
	prefix := "0"
	if p.Code == CodeBinary {
		prefix = "1"
	}
	header := []byte(fmt.Sprintf("%s%d", prefix, l))
	for i, b := range header {
		header[i] = b - '0'
	}
	if _, err := e.w.Write(header); err != nil {
		return err
	}
	if _, err := e.w.Write([]byte{0xff}); err != nil {
		return err
	}
	encoder, err := NewEncoder(e.w, p.Type, p.Code)
	if err != nil {
		return err
	}
	if _, err := encoder.Write(p.Data); err != nil {
		return err
	}
	return nil
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
