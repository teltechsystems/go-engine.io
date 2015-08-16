package polling

import (
	"net/http"

	"github.com/googollee/go-engine.io/parser"
	"github.com/googollee/go-engine.io/transport"
)

type client struct {
	req        *http.Request
	response   *http.Response
	data       []parser.Packet
	getLocker  tryLocker
	postLocker tryLocker
}

func NewClient(r *http.Request) (transport.Client, error) {
	return &client{
		req: r,
	}, nil
}
