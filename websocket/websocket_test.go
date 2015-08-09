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
			w, err := s.NextWriter(parser.MessageText, parser.PacketOpen)
			assert.MustEqual(t, err, nil)
			err = w.Close()
			assert.MustEqual(t, err, nil)
		}

		<-sync
		sync <- 1

		{
			r, err := s.NextReader()
			assert.MustEqual(t, err, nil)
			assert.MustEqual(t, r.MessageType(), parser.MessageBinary)
			assert.MustEqual(t, r.PacketType(), parser.PacketMessage)
			b, err := ioutil.ReadAll(r)
			assert.MustEqual(t, err, nil)
			assert.Equal(t, string(b), "测试Binary")
			err = r.Close()
			assert.MustEqual(t, err, nil)
		}

		<-sync
		sync <- 1

		{
			r, err := s.NextReader()
			assert.MustEqual(t, err, nil)
			assert.MustEqual(t, r.MessageType(), parser.MessageText)
			assert.MustEqual(t, r.PacketType(), parser.PacketMessage)
			b, err := ioutil.ReadAll(r)
			assert.MustEqual(t, err, nil)
			assert.Equal(t, string(b), "测试Text")
			err = r.Close()
			assert.MustEqual(t, err, nil)
		}

		<-sync
		sync <- 1

		{
			w, err := s.NextWriter(parser.MessageText, parser.PacketMessage)
			assert.MustEqual(t, err, nil)
			_, err = w.Write([]byte("日本語Text"))
			assert.MustEqual(t, err, nil)
			err = w.Close()
			assert.MustEqual(t, err, nil)
		}

		<-sync
		sync <- 1

		{
			w, err := s.NextWriter(parser.MessageBinary, parser.PacketMessage)
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
		decoder, _ := c.NextReader()
		assert.Equal(t, decoder.PacketType(), parser.PacketOpen)
		assert.Equal(t, decoder.MessageType(), parser.MessageText)
		decoder.Close()
	}

	sync <- 1
	<-sync

	{
		w, err := c.NextWriter(parser.MessageBinary, parser.PacketMessage)
		assert.MustEqual(t, err, nil)
		_, err = w.Write([]byte("测试Binary"))
		assert.MustEqual(t, err, nil)
		err = w.Close()
		assert.MustEqual(t, err, nil)
	}

	sync <- 1
	<-sync

	{
		w, err := c.NextWriter(parser.MessageText, parser.PacketMessage)
		assert.MustEqual(t, err, nil)
		_, err = w.Write([]byte("测试Text"))
		assert.MustEqual(t, err, nil)
		err = w.Close()
		assert.MustEqual(t, err, nil)
	}

	sync <- 1
	<-sync

	{
		decoder, _ := c.NextReader()
		assert.Equal(t, decoder.PacketType(), parser.PacketMessage)
		assert.Equal(t, decoder.MessageType(), parser.MessageText)
		r, err := ioutil.ReadAll(decoder)
		decoder.Close()
		assert.MustEqual(t, err, nil)
		assert.Equal(t, string(r), "日本語Text")
	}

	sync <- 1
	<-sync

	{
		decoder, _ := c.NextReader()
		assert.Equal(t, decoder.PacketType(), parser.PacketMessage)
		assert.Equal(t, decoder.MessageType(), parser.MessageBinary)
		r, err := ioutil.ReadAll(decoder)
		decoder.Close()
		assert.MustEqual(t, err, nil)
		assert.Equal(t, string(r), "日本語Binary")
	}

	sync <- 1
	<-sync
}
