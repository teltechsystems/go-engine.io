package polling

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/googollee/go-engine.io/parser"
	"github.com/googollee/go-engine.io/transport"
)

type client struct {
	url           url.URL
	req           http.Request
	resp          *http.Response
	seq           uint
	data          []parser.Packet
	decoder       *parser.PayloadDecoder
	readDeadline  time.Time
	writeDeadline time.Time
	getResp       *http.Response
	posting       tryLocker
	postError     error
	postErrLocker sync.Mutex
	isClose       int32
}

func NewClient(r *http.Request) (transport.Client, error) {
	ret := &client{
		req: *r,
		url: *r.URL,
		seq: 0,
	}
	return ret, nil
}

func (c *client) Response() *http.Response {
	return c.resp
}

func (c *client) SetReadDeadline(t time.Time) error {
	c.readDeadline = t
	return nil
}

func (c *client) SetWriteDeadline(t time.Time) error {
	c.writeDeadline = t
	return nil
}

func (c *client) NextReader() (parser.CodeType, parser.PacketType, io.ReadCloser, error) {
	if c.isClosed() {
		return 0, 0, nil, io.EOF
	}
	if c.decoder != nil {
		ret, err := c.decoder.Next()
		if err == nil {
			return ret.CodeType(), ret.PacketType(), ioutil.NopCloser(ret), nil
		}
		if err != io.EOF {
			return 0, 0, nil, err
		}
		c.getResp.Body.Close()
		c.decoder = nil
	}
	req := c.getReq()
	req.Method = "GET"
	var err error
	client := *http.DefaultClient
	if !c.readDeadline.IsZero() {
		now := time.Now()
		if c.readDeadline.Before(now) {
			return 0, 0, nil, ErrTimeout
		}
		client.Timeout = now.Sub(c.readDeadline)
	}
	c.getResp, err = client.Do(req)
	if err != nil {
		return 0, 0, nil, err
	}
	if c.resp == nil {
		c.resp = c.getResp
	}
	c.decoder = parser.NewPayloadDecoder(c.getResp.Body)
	ret, err := c.decoder.Next()
	if err != nil {
		return 0, 0, nil, err
	}
	return ret.CodeType(), ret.PacketType(), ioutil.NopCloser(ret), err
}

func (c *client) NextWriter(code parser.CodeType, typ parser.PacketType) (io.WriteCloser, error) {
	if c.isClosed() {
		return nil, io.EOF
	}

	if !c.writeDeadline.IsZero() {
		now := time.Now()
		if c.writeDeadline.Before(now) {
			return nil, ErrTimeout
		}
	}
	if err := c.getPostError(); err != nil {
		return nil, err
	}

	return newClientWriter(c, code, typ), nil
}

func (c *client) Close() error {
	atomic.StoreInt32(&c.isClose, 1)
	return nil
}

func (c *client) isClosed() bool {
	return atomic.LoadInt32(&c.isClose) != 0
}

func (c *client) getReq() *http.Request {
	req := c.req
	url := c.url
	req.URL = &url
	query := req.URL.Query()
	query.Set("t", fmt.Sprintf("%d-%d", time.Now().Unix()*1000, c.seq))
	c.seq++
	req.URL.RawQuery = query.Encode()
	return &req
}

func (c *client) getPostError() error {
	c.postErrLocker.Lock()
	defer c.postErrLocker.Unlock()

	return c.postError
}

func (c *client) doPost() error {
	if c.isClosed() {
		return io.EOF
	}
	if !c.posting.TryLock() {
		return nil
	}

	req := c.getReq()
	req.Method = "POST"
	buf := bytes.NewBuffer(nil)
	data := c.data
	c.data = nil
	if err := parser.EncodeToBinaryPayload(buf, data); err != nil {
		c.posting.Unlock()
		return err
	}
	req.Body = ioutil.NopCloser(buf)
	client := *http.DefaultClient
	if !c.writeDeadline.IsZero() {
		now := time.Now()
		if c.writeDeadline.Before(now) {
			c.posting.Unlock()
			return ErrTimeout
		}
		client.Timeout = c.writeDeadline.Sub(now)
	}
	go func() {
		defer c.posting.Unlock()
		resp, err := client.Do(req)
		if err != nil {
			c.postErrLocker.Lock()
			c.postError = err
			c.postErrLocker.Unlock()
			return
		}
		if resp.StatusCode != http.StatusOK {
			c.postErrLocker.Lock()
			c.postError = errors.New(resp.Status)
			c.postErrLocker.Unlock()
		}
	}()
	return nil
}

type clientWriter struct {
	code   parser.CodeType
	typ    parser.PacketType
	buf    *bytes.Buffer
	client *client
}

func newClientWriter(c *client, code parser.CodeType, typ parser.PacketType) io.WriteCloser {
	return &clientWriter{
		code:   code,
		typ:    typ,
		buf:    bytes.NewBuffer(nil),
		client: c,
	}
}

func (w *clientWriter) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

func (w *clientWriter) Close() error {
	w.client.data = append(w.client.data, parser.Packet{
		Code: w.code,
		Type: w.typ,
		Data: w.buf.Bytes(),
	})
	return w.client.doPost()
}
