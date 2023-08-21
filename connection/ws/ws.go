package ws

import (
	"context"
	"crypto/tls"
	"errors"
	"time"
)

const DefaultHandshakeTimeout = 15 * time.Second

// ErrAlreadyClosed is returned when trying to read/write from/to closed connection
var ErrAlreadyClosed = errors.New("websocket connection is closed")

// Conn provides interface for websocket connection
type Conn interface {
	ReadMessage(context.Context) ([]byte, error)
	WriteJSON(context.Context, interface{}) error
	Close() error
}

// DialOptions represents Dial's options.
type DialOptions struct {
	HandshakeTimeout time.Duration
	TLSClientConfig  *tls.Config
}
