package websocket

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/googollee/go-assert"
	"github.com/googollee/go-engine.io/transport"
	"github.com/gorilla/websocket"

	"github.com/googollee/go-engine.io/parser"
)

type fakeCallback struct {
	onPacket    chan bool
	messageType parser.MessageType
	packetType  parser.PacketType
	body        []byte
	err         error
	closedCount int
	countLocker sync.Mutex
	closeServer transport.Server
}

func newFakeCallback() *fakeCallback {
	return &fakeCallback{
		onPacket: make(chan bool),
	}
}

func (f *fakeCallback) OnPacket(r *parser.PacketDecoder) {
	f.packetType = r.PacketType()
	f.messageType = r.MessageType()
	f.body, f.err = ioutil.ReadAll(r)
	f.onPacket <- true
}

func (f *fakeCallback) OnClose(s transport.Server) {
	f.countLocker.Lock()
	defer f.countLocker.Unlock()
	f.closedCount++
	f.closeServer = s
}

func (f *fakeCallback) ClosedCount() int {
	f.countLocker.Lock()
	defer f.countLocker.Unlock()
	return f.closedCount
}

func TestWebsocket(t *testing.T) {
	creater := NewCreater(&websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	})
	sync := make(chan int)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f := newFakeCallback()
		s, err := creater.Server(w, r, f)
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
			<-f.onPacket
			assert.MustEqual(t, f.err, nil)
			assert.MustEqual(t, f.messageType, parser.MessageBinary)
			assert.MustEqual(t, f.packetType, parser.PacketMessage)
			assert.Equal(t, string(f.body), "测试Binary")
		}

		<-sync
		sync <- 1

		{
			<-f.onPacket
			assert.MustEqual(t, f.err, nil)
			assert.MustEqual(t, f.messageType, parser.MessageText)
			assert.MustEqual(t, f.packetType, parser.PacketMessage)
			assert.Equal(t, string(f.body), "测试Text")
		}

		<-sync
		sync <- 1

		{
			w, err := s.NextWriter(parser.MessageText, parser.PacketMessage)
			assert.MustEqual(t, err, nil)
			w.Write([]byte("日本語Text"))
			err = w.Close()
			assert.MustEqual(t, err, nil)
		}

		<-sync
		sync <- 1

		{
			w, err := s.NextWriter(parser.MessageBinary, parser.PacketMessage)
			assert.MustEqual(t, err, nil)
			w.Write([]byte("日本語Binary"))
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
		w, _ := c.NextWriter(parser.MessageBinary, parser.PacketMessage)
		w.Write([]byte("测试Binary"))
		w.Close()
	}

	sync <- 1
	<-sync

	{
		w, _ := c.NextWriter(parser.MessageText, parser.PacketMessage)
		w.Write([]byte("测试Text"))
		w.Close()
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
