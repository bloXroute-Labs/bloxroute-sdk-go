package ws

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/fasthttp/websocket"
)

const (
	messageSizeLimit = 15 * 1024 * 1024
	msgCHanSize      = 1000
)

type connection struct {
	remoteAddress string
	closed        chan struct{}
	msgsJSON      chan interface{}

	fastHTTPConn *websocket.Conn
}

// Dial dials websocket connection which is safe to use in goroutines
func Dial(ctx context.Context, url string, headers http.Header, opts *DialOptions) (Conn, error) {
	c := &connection{
		remoteAddress: url,
		closed:        make(chan struct{}),
		msgsJSON:      make(chan interface{}, msgCHanSize),
	}

	if opts == nil {
		opts = &DialOptions{}
	}

	if opts.HandshakeTimeout == 0 {
		opts.HandshakeTimeout = DefaultHandshakeTimeout
	}

	if opts.TLSClientConfig == nil {
		opts.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: opts.HandshakeTimeout,
		TLSClientConfig:  opts.TLSClientConfig,
	}

	var err error
	c.fastHTTPConn, _, err = dialer.DialContext(ctx, url, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to dial websocket connection: %w", err)
	}

	c.fastHTTPConn.SetReadLimit(messageSizeLimit)

	go c.write()

	return c, nil
}

// ReadMessage reads text message from websocket connection
func (c *connection) ReadMessage(ctx context.Context) (data []byte, err error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.closed:
		return nil, ErrAlreadyClosed
	default:
	}

	_, data, err = c.fastHTTPConn.ReadMessage()

	if err != nil && IsWSClosedError(err) {
		return nil, c.Close()
	}

	return data, err
}

// WriteJSON writes JSON message to websocket connection
func (c *connection) WriteJSON(ctx context.Context, v interface{}) error {
	select {
	case <-c.closed:
		return ErrAlreadyClosed
	case <-ctx.Done():
		return nil
	case c.msgsJSON <- v:
	default:
		// in case the channel is full, we close the connection
		return c.Close()
	}

	return nil
}

func (c *connection) write() {
	for {
		select {
		case <-c.closed:
			return
		case msg := <-c.msgsJSON:
			err := c.writeTimeoutJSON(msg)
			if err != nil && IsWSClosedError(err) {
				return
			}
		}
	}
}

// Close closes websocket connection
// It is safe to call Close multiple times
func (c *connection) Close() error {
	select {
	case <-c.closed:
		return nil
	default:
		close(c.closed)
	}

	err := c.fastHTTPConn.Close()
	if err != nil {
		if IsWSClosedError(err) {
			return nil
		}

		return fmt.Errorf("failed to close websocket connection: %w", err)
	}

	return nil
}

// IsWSClosedError checks if error is websocket close error
func IsWSClosedError(err error) bool {
	if err == nil {
		return false
	}

	var websocketCloseErr *websocket.CloseError
	if errors.As(err, &websocketCloseErr) ||
		errors.Is(err, ErrAlreadyClosed) ||
		errors.Is(err, io.EOF) ||
		strings.Contains(err.Error(), "already wrote close") ||
		strings.Contains(err.Error(), "EOF") ||
		strings.Contains(err.Error(), "broken pipe") ||
		strings.Contains(err.Error(), "connection reset by peer") {
		return true
	}

	return false
}

// writeTimeoutJSON writes JSON message to websocket connection with timeout
// should not be used for concurrent writes
func (c *connection) writeTimeoutJSON(msg interface{}) error {
	return c.fastHTTPConn.WriteJSON(msg)
}
