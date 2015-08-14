package transport

import (
	"io"
	"net/http"
	"time"

	"github.com/googollee/go-engine.io/parser"
)

type Creater struct {
	Name   string
	Server func(http.ResponseWriter, *http.Request) (Server, error)
	Client func(*http.Request) (Client, error)
}

// Conn is a transport connection.
type Conn interface {

	// NextReader returns packet reader. This function call should be synced.
	NextReader() (parser.CodeType, parser.PacketType, io.ReadCloser, error)

	// NextWriter returns packet writer. This function call should be synced.
	NextWriter(parser.CodeType, parser.PacketType) (io.WriteCloser, error)

	// Close closes the transport.
	Close() error

	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

// Server is a transport layer in server.
type Server interface {
	Conn
	// ServeHTTP handles the http request. It will call conn.onPacket when receive packet.
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// Client is a transport layer in client to connect server.
type Client interface {
	Conn
	// Response returns the response of last http request.
	Response() *http.Response
}
