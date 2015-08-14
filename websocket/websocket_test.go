package websocket

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/googollee/go-assert"
	"github.com/gorilla/websocket"

	"github.com/googollee/go-engine.io/parser"
)

func TestWebsocket(t *testing.T) {
	creater := NewCreater(&websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	})
	sync := make(chan int)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s, err := creater.Server(w, r)
		assert.MustEqual(t, err, nil)
		defer s.Close()

		{
			req, err := http.NewRequest("GET", "/", nil)
			assert.MustEqual(t, err, nil)
			recoder := httptest.NewRecorder()
			s.ServeHTTP(recoder, req)
			assert.MustEqual(t, recoder.Code, http.StatusBadRequest)
		}

		{
			w, err := s.NextWriter(parser.CodeText, parser.PacketOpen)
			assert.MustEqual(t, err, nil)
			err = w.Close()
			assert.MustEqual(t, err, nil)
		}

		<-sync
		sync <- 1

		{
			code, typ, r, err := s.NextReader()
			assert.MustEqual(t, err, nil)
			assert.MustEqual(t, code, parser.CodeBinary)
			assert.MustEqual(t, typ, parser.PacketMessage)
			b, err := ioutil.ReadAll(r)
			assert.MustEqual(t, err, nil)
			assert.Equal(t, string(b), "测试Binary")
			err = r.Close()
			assert.MustEqual(t, err, nil)
		}

		<-sync
		sync <- 1

		{
			code, typ, r, err := s.NextReader()
			assert.MustEqual(t, err, nil)
			assert.MustEqual(t, code, parser.CodeText)
			assert.MustEqual(t, typ, parser.PacketMessage)
			b, err := ioutil.ReadAll(r)
			assert.MustEqual(t, err, nil)
			assert.Equal(t, string(b), "测试Text")
			err = r.Close()
			assert.MustEqual(t, err, nil)
		}

		<-sync
		sync <- 1

		{
			w, err := s.NextWriter(parser.CodeText, parser.PacketMessage)
			assert.MustEqual(t, err, nil)
			_, err = w.Write([]byte("日本語Text"))
			assert.MustEqual(t, err, nil)
			err = w.Close()
			assert.MustEqual(t, err, nil)
		}

		<-sync
		sync <- 1

		{
			w, err := s.NextWriter(parser.CodeBinary, parser.PacketMessage)
			assert.MustEqual(t, err, nil)
			_, err = w.Write([]byte("日本語Binary"))
			assert.MustEqual(t, err, nil)
			err = w.Close()
			assert.MustEqual(t, err, nil)
		}

		<-sync
		sync <- 1
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	u.Scheme = "ws"
	req, err := http.NewRequest("GET", u.String(), nil)
	assert.MustEqual(t, err, nil)

	c, _ := NewClient(req)
	defer c.Close()

	{
		code, typ, r, err := c.NextReader()
		assert.MustEqual(t, err, nil)
		assert.Equal(t, typ, parser.PacketOpen)
		assert.Equal(t, code, parser.CodeText)
		b, err := ioutil.ReadAll(r)
		assert.MustEqual(t, err, nil)
		assert.Equal(t, len(b), 0)
		err = r.Close()
		assert.MustEqual(t, err, nil)
	}

	sync <- 1
	<-sync

	{
		w, err := c.NextWriter(parser.CodeBinary, parser.PacketMessage)
		assert.MustEqual(t, err, nil)
		_, err = w.Write([]byte("测试Binary"))
		assert.MustEqual(t, err, nil)
		err = w.Close()
		assert.MustEqual(t, err, nil)
	}

	sync <- 1
	<-sync

	{
		w, err := c.NextWriter(parser.CodeText, parser.PacketMessage)
		assert.MustEqual(t, err, nil)
		_, err = w.Write([]byte("测试Text"))
		assert.MustEqual(t, err, nil)
		err = w.Close()
		assert.MustEqual(t, err, nil)
	}

	sync <- 1
	<-sync

	{
		code, typ, r, _ := c.NextReader()
		assert.Equal(t, typ, parser.PacketMessage)
		assert.Equal(t, code, parser.CodeText)
		b, err := ioutil.ReadAll(r)
		assert.MustEqual(t, err, nil)
		assert.Equal(t, string(b), "日本語Text")
		err = r.Close()
		assert.MustEqual(t, err, nil)
	}

	sync <- 1
	<-sync

	{
		code, typ, r, _ := c.NextReader()
		assert.Equal(t, typ, parser.PacketMessage)
		assert.Equal(t, code, parser.CodeBinary)
		b, err := ioutil.ReadAll(r)
		assert.MustEqual(t, err, nil)
		assert.Equal(t, string(b), "日本語Binary")
		err = r.Close()
		assert.MustEqual(t, err, nil)
	}

	sync <- 1
	<-sync
}
