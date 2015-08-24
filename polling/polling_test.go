package polling

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/googollee/go-assert"
	"github.com/googollee/go-engine.io/parser"
)

func TestPolling(t *testing.T) {
	svr, err := NewServer(nil, nil)
	assert.MustEqual(t, err, nil)
	httpSvr := httptest.NewServer(svr)
	defer httpSvr.Close()
	defer svr.Close()

	req, err := http.NewRequest("GET", httpSvr.URL, nil)
	assert.MustEqual(t, err, nil)
	clt, err := NewClient(req)
	assert.MustEqual(t, err, nil)
	defer clt.Close()

	w, err := svr.NextWriter(parser.CodeBinary, parser.PacketOpen)
	assert.MustEqual(t, err, nil)
	err = w.Close()
	assert.MustEqual(t, err, nil)

	c, ty, r, err := clt.NextReader()
	assert.MustEqual(t, err, nil)
	assert.Equal(t, c, parser.CodeBinary)
	assert.Equal(t, ty, parser.PacketOpen)
	b, err := ioutil.ReadAll(r)
	assert.MustEqual(t, err, nil)
	assert.Equal(t, len(b), 0)
	err = r.Close()
	assert.MustEqual(t, err, nil)

	w, err = clt.NextWriter(parser.CodeText, parser.PacketMessage)
	assert.MustEqual(t, err, nil)
	_, err = w.Write([]byte("测试"))
	assert.MustEqual(t, err, nil)
	err = w.Close()
	assert.MustEqual(t, err, nil)

	c, ty, r, err = svr.NextReader()
	assert.MustEqual(t, err, nil)
	assert.Equal(t, c, parser.CodeText)
	assert.Equal(t, ty, parser.PacketMessage)
	b, err = ioutil.ReadAll(r)
	assert.MustEqual(t, err, nil)
	assert.Equal(t, string(b), "测试")
	err = r.Close()
	assert.MustEqual(t, err, nil)
}
