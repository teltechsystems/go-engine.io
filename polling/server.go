package polling

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"math"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/googollee/go-engine.io/parser"
	"github.com/googollee/go-engine.io/transport"
)

var ErrReadTimeout = errors.New("read timeout")

type Polling struct {
	sendChan      chan bool
	encoder       *parser.PayloadEncoder
	decoderChan   chan *parser.PacketDecoder
	readDeadline  time.Time
	writeDeadline time.Time
	getLocker     sync.Mutex
	postLocker    sync.Mutex
	isClosed      int32
}

func NewServer(w http.ResponseWriter, r *http.Request) (transport.Server, error) {
	newEncoder := parser.NewBinaryPayloadEncoder
	if r.URL.Query()["b64"] != nil {
		newEncoder = parser.NewTextPayloadEncoder
	}
	ret := &Polling{
		sendChan: makeSendChan(),
		encoder:  newEncoder(),
	}
	return ret, nil
}

func (p *Polling) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if p.IsClosed() {
		http.Error(w, "closed", http.StatusForbidden)
		return
	}
	switch r.Method {
	case "GET":
		p.get(w, r)
	case "POST":
		p.post(w, r)
	}
}

func (p *Polling) Close() error {
	atomic.StoreInt32(&p.isClosed, 1)
	close(p.sendChan)
	close(p.decoderChan)
	return nil
}

func (p *Polling) IsClosed() bool {
	return atomic.LoadInt32(&p.isClosed) != 0
}

func (p *Polling) NextWriter(msg parser.MessageType, pkg parser.PacketType) (io.WriteCloser, error) {
	if p.IsClosed() {
		return nil, io.EOF
	}

	ret, err := p.encoder.Next(pkg, msg)

	if err != nil {
		return nil, err
	}
	return newWriter(ret, p), nil
}

func (p *Polling) NextReader() (*parser.PacketDecoder, error) {
	if p.IsClosed() {
		return nil, io.EOF
	}

	timeout := time.Duration(math.MaxInt64)
	if !p.readDeadline.IsZero() {
		timeout = p.readDeadline.Sub(time.Now())
	}

	select {
	case d := <-p.decoderChan:
		return d, nil
	case <-time.After(timeout):
	}
	return nil, ErrReadTimeout
}

func (p *Polling) SetReadDeadline(t time.Time) error {
	p.readDeadline = t
	return nil
}

func (p *Polling) SetWriteDeadline(t time.Time) error {
	p.writeDeadline = t
	return nil
}

func (p *Polling) get(w http.ResponseWriter, r *http.Request) {
	p.getLocker.Lock()
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

	if j := r.URL.Query().Get("j"); j != "" {
		// JSONP Polling
		w.Header().Set("Content-Type", "text/javascript; charset=UTF-8")
		w.Write([]byte("___eio[" + j + "](\""))
		tmp := bytes.Buffer{}
		p.encoder.EncodeTo(&tmp)
		template.JSEscape(w, tmp.Bytes())
		w.Write([]byte("\");"))
	} else {
		// XHR Polling
		if p.encoder.IsText() {
			w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		} else {
			w.Header().Set("Content-Type", "application/octet-stream")
		}
		p.encoder.EncodeTo(w)
	}
}

func (p *Polling) post(w http.ResponseWriter, r *http.Request) {
	p.postLocker.Lock()
	defer p.postLocker.Unlock()

	w.Header().Set("Content-Type", "text/html")

	var decoder *parser.PayloadDecoder
	if j := r.URL.Query().Get("j"); j != "" {
		// JSONP Polling
		d := r.FormValue("d")
		decoder = parser.NewPayloadDecoder(bytes.NewBufferString(d))
	} else {
		// XHR Polling
		decoder = parser.NewPayloadDecoder(r.Body)
	}
	for {
		d, err := decoder.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		r := newReader(d.ReadCloser)
		d.ReadCloser = r
		p.decoderChan <- d
		r.wait()
	}
	w.Write([]byte("ok"))
}
