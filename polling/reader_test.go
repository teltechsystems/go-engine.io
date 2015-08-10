package polling

import (
	"io"
	"strings"
	"testing"

	"github.com/googollee/go-assert"
)

type fakeReader struct {
	io.Reader
	closed bool
}

func (f *fakeReader) Close() error {
	f.closed = true
	return nil
}

func TestReader(t *testing.T) {
	r := fakeReader{
		Reader: strings.NewReader("abc"),
	}
	reader := newReader(r)

	var b [10]byte
	n, err := reader.Read()
	assert.MustEqual(t, err, nil)
	assert.Equal(t, n, 3)
	assert.Equal(t, b[:n], "abc")

	assert.MustEqual(t, r.closed, false)
	go reader.Close()
	reader.wait()
	assert.MustEqual(t, r.closed, true)
}
