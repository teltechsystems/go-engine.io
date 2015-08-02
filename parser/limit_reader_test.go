package parser

import (
	"bytes"
	"io"
	"testing"

	"github.com/googollee/go-assert"
)

func TestLimitReader(t *testing.T) {
	b := bytes.NewBufferString("1234567890")
	r := newLimitReader(b, 5)
	p := make([]byte, 1024)

	n, err := r.Read(p)
	assert.MustEqual(t, err, nil)
	assert.Equal(t, string(p[:n]), "12345")

	n, err = r.Read(p)
	assert.MustEqual(t, err, io.EOF)

	err = r.Close()
	assert.MustEqual(t, err, nil)
	assert.Equal(t, b.String(), "67890")
}

func TestRemainLimitReader(t *testing.T) {
	b := bytes.NewBufferString("1234567890")
	r := newLimitReader(b, 5)
	p := make([]byte, 3)

	n, err := r.Read(p)
	assert.MustEqual(t, err, nil)
	assert.Equal(t, string(p[:n]), "123")

	err = r.Close()
	assert.MustEqual(t, err, nil)
	assert.Equal(t, b.String(), "67890")
}
