package polling

import (
	"bytes"
	"errors"
	"io"
	"math"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/googollee/go-engine.io/parser"
	"github.com/googollee/go-engine.io/transport"
)

var ErrTimeout = errors.New("timeout")

type server struct {
	sendChan chan bool
	readChan chan *reader
	data     []parser.Packet

	readDeadline  time.Time
	writeDeadline time.Time
	getLocker     tryLocker
	postLocker    tryLocker
	isClosed      int32
	waiter        sync.WaitGroup
}

func NewServer(w http.ResponseWriter, r *http.Request) (transport.Server, error) {
	ret := &server{
		sendChan: make(chan bool, 1),
		readChan: make(chan *reader),
	}
	return ret, nil
}

func (p *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if p.IsClosed() {
		http.Error(w, "closed", http.StatusForbidden)
		return
	}
	p.waiter.Add(1)
	defer p.waiter.Done()

	switch r.Method {
	case "GET":
		p.get(w, r)
	case "POST":
		p.post(w, r)
	}
}

func (p *server) Close() error {
	if !atomic.CompareAndSwapInt32(&p.isClosed, 0, 1) {
		return nil
	}
	p.waiter.Wait()
	close(p.sendChan)
	close(p.readChan)
	return nil
}

func (p *server) IsClosed() bool {
	return atomic.LoadInt32(&p.isClosed) != 0
}

func (p *server) NextWriter(code parser.CodeType, typ parser.PacketType) (io.WriteCloser, error) {
	if p.IsClosed() {
		return nil, io.EOF
	}

	return newWriter(p, code, typ), nil
}

func (p *server) NextReader() (parser.CodeType, parser.PacketType, io.ReadCloser, error) {
	if p.IsClosed() {
		return 0, 0, nil, io.EOF
	}

	timeout := time.Duration(math.MaxInt64)
	if !p.readDeadline.IsZero() {
		timeout = p.readDeadline.Sub(time.Now())
	}

	select {
	case d := <-p.readChan:
		return d.CodeType(), d.PacketType(), d, nil
	case <-time.After(timeout):
	}
	return 0, 0, nil, ErrTimeout
}

func (p *server) SetReadDeadline(t time.Time) error {
	p.readDeadline = t
	return nil
}

func (p *server) SetWriteDeadline(t time.Time) error {
	p.writeDeadline = t
	return nil
}

func (p *server) get(w http.ResponseWriter, r *http.Request) {
	if !p.getLocker.TryLock() {
		http.Error(w, "interleave GET", http.StatusBadRequest)
	}
	defer p.getLocker.Unlock()

	timeout := time.Duration(math.MaxInt64)
	if !p.writeDeadline.IsZero() {
		timeout = p.writeDeadline.Sub(time.Now())
	}

	select {
	case <-p.sendChan:
	case <-time.After(timeout):
		http.Error(w, "timeout", http.StatusRequestTimeout)
		return
	}

	encode := parser.EncodeToBinaryPayload
	if r.URL.Query()["b64"] != nil {
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		encode = parser.EncodeToTextPayload
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	if j := r.URL.Query().Get("j"); j != "" {
		// JSONP Polling
		w.Header().Set("Content-Type", "text/javascript; charset=UTF-8")
		w.Write([]byte("___eio[" + j + "](\""))
		encode(newJSWriter(w), p.data)
		w.Write([]byte("\");"))
	} else {
		// XHR Polling
		encode(w, p.data)
	}
}

func (p *server) post(w http.ResponseWriter, r *http.Request) {
	if !p.postLocker.TryLock() {
		http.Error(w, "interleave POST", http.StatusBadRequest)
	}
	defer p.postLocker.Unlock()

	var decoder *parser.PayloadDecoder
	if j := r.URL.Query().Get("j"); j != "" {
		// JSONP Polling
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		d := r.FormValue("d")
		decoder = parser.NewPayloadDecoder(bytes.NewBufferString(d))
	} else {
		// XHR Polling
		decoder = parser.NewPayloadDecoder(r.Body)
	}
	deadline := p.readDeadline
	for {
		d, err := decoder.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		r := newReader(d)
		timeout := time.Duration(math.MaxInt64)
		if !deadline.IsZero() {
			timeout = deadline.Sub(time.Now())
		}
		select {
		case p.readChan <- r:
		case <-time.After(timeout):
			http.Error(w, "timeout", http.StatusRequestTimeout)
			return
		}
		timeout = time.Duration(math.MaxInt64)
		if !deadline.IsZero() {
			timeout = deadline.Sub(time.Now())
		}
		select {
		case <-r.wait():
		case <-time.After(timeout):
			http.Error(w, "timeout", http.StatusRequestTimeout)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("ok"))
}
