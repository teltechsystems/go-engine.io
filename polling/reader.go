package polling

import "io"

type reader struct {
	io.ReadCloser
	closed chan struct{}
}

func newReader(r io.ReadCloser) *reader {
	return &reader{
		ReadCloser: r,
		closed:     make(chan struct{}),
	}
}

func (r *reader) Close() error {
	defer func() {
		r.closed <- struct{}{}
	}()
	return r.ReadCloser.Close()
}

func (r *reader) wait() {
	<-r.closed
}
