package polling

import "io"

func makeSendChan() chan bool {
	return make(chan bool, 1)
}

type writer struct {
	io.WriteCloser
	server *Polling
}

func newWriter(w io.WriteCloser, server *Polling) *writer {
	return &writer{
		WriteCloser: w,
		server:      server,
	}
}

func (w *writer) Close() error {
	select {
	case w.server.sendChan <- true:
	default:
	}
	return w.WriteCloser.Close()
}
