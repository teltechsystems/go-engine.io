package polling

import (
	"strings"
	"testing"

	"github.com/googollee/go-assert"
	"github.com/googollee/go-engine.io/parser"
)

func TestReader(t *testing.T) {
	buf := strings.NewReader("4\xe6\xb5\x8b\xe8\xaf\x95")
	decoder, err := parser.NewDecoder(buf)
	assert.MustEqual(t, err, nil)

	reader := newReader(decoder)

	var b [10]byte
	n, err := reader.Read(b[:])
	assert.MustEqual(t, err, nil)
	assert.Equal(t, n, 6)
	assert.Equal(t, string(b[:n]), "测试")

	sync := make(chan bool)
	go func() {
		err := reader.Close()
		assert.MustEqual(t, err, nil)
		sync <- true
	}()
	<-reader.wait()
	<-sync
}
