package polling

import "github.com/googollee/go-engine.io/parser"

type reader struct {
	*parser.PacketDecoder
	closed chan struct{}
}

func newReader(d *parser.PacketDecoder) *reader {
	return &reader{
		PacketDecoder: d,
		closed:        make(chan struct{}),
	}
}

func (r *reader) Close() error {
	r.closed <- struct{}{}

	return nil
}

func (r *reader) wait() chan struct{} {
	return r.closed
}
