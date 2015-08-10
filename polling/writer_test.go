package polling

import (
	"bytes"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWriter(t *testing.T) {
	p := &Polling{
		sendChan: makeSendChan(),
	}
	sendChan := p.sendChan

	Convey("Wait close", t, func() {
		w := newFakeWriteCloser()

		select {
		case <-sendChan:
			t.Fatal("should not run here")
		default:
		}

		writer := newWriter(w, p)
		err := writer.Close()
		So(err, ShouldBeNil)

		select {
		case <-sendChan:
		default:
			panic("should not run here")
		}

		select {
		case <-sendChan:
			panic("should not run here")
		default:
		}
	})

	Convey("Many writer with close", t, func() {
		for i := 0; i < 10; i++ {
			w := newFakeWriteCloser()
			writer := newWriter(w, p)
			err := writer.Close()
			So(err, ShouldBeNil)
		}

		select {
		case <-sendChan:
		default:
			panic("should not run here")
		}

		select {
		case <-sendChan:
			panic("should not run here")
		default:
		}
	})

	Convey("Close with not normal", t, func() {
		p := &Polling{
			state:    stateClosing,
			sendChan: makeSendChan(),
		}

		w := newFakeWriteCloser()
		writer := newWriter(w, p)
		err := writer.Close()
		So(err, ShouldNotBeNil)
	})
}

type fakeWriteCloser struct {
	*bytes.Buffer
}

func newFakeWriteCloser() *fakeWriteCloser {
	return &fakeWriteCloser{
		Buffer: bytes.NewBuffer(nil),
	}
}

func (f *fakeWriteCloser) Close() error {
	return nil
}
