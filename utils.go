package bloxroute_sdk_go

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/fasthttp/websocket"
	"github.com/sourcegraph/jsonrpc2"
)

func reconnect(ctx context.Context, dialer *websocket.Dialer, url, authHeader string) (*websocket.Conn, error) {
	backOff := backoff.NewExponentialBackOff()
	backOff.MaxElapsedTime = reconnectTimeout
	backOff.InitialInterval = reconnectInitialInterval

	var conn *websocket.Conn

	fn := func() error {
		var err error
		conn, _, err = dialer.DialContext(ctx, url, http.Header{authHeaderKey: []string{authHeader}})
		if err != nil {
			return fmt.Errorf("failed to reconnect to cloud API after %s: %w", reconnectTimeout, err)
		}

		return nil
	}

	err := backoff.Retry(fn, backOff)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func randomID() jsonrpc2.ID {
	return jsonrpc2.ID{Str: strconv.FormatUint(rand.New(rand.NewSource(time.Now().UnixNano())).Uint64(), 10), IsString: true}
}

// isWSClosedError checks if error is websocket close error
func isWSClosedError(err error) bool {
	if err == nil {
		return false
	}

	var websocketCloseErr *websocket.CloseError
	if errors.As(err, &websocketCloseErr) ||
		errors.Is(err, io.EOF) ||
		strings.Contains(err.Error(), "already wrote close") ||
		strings.Contains(err.Error(), "EOF") ||
		strings.Contains(err.Error(), "broken pipe") ||
		strings.Contains(err.Error(), "connection reset by peer") {
		return true
	}

	return false
}
