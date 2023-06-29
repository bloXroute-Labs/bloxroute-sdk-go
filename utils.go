package bloxroute_sdk_go

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/bloXroute-Labs/bloxroute-sdk-go/utils"
	"github.com/cenkalti/backoff/v4"
	"github.com/fasthttp/websocket"
	"github.com/sourcegraph/jsonrpc2"
)

func reconnect(ctx context.Context, dialer *websocket.Dialer, url, authHeader string) (*websocket.Conn, error) {
	backOff := backoff.NewExponentialBackOff()
	backOff.MaxElapsedTime = reconnectTimeout
	backOff.InitialInterval = reconnectInitialInterval

	var conn *websocket.Conn
	fn := func() (err error) {
		conn, _, err = dialer.DialContext(ctx, url, http.Header{authHeaderKey: []string{authHeader}})
		if err != nil {
			return fmt.Errorf("failed to reconnect to cloud API after %s: %w", reconnectTimeout, err)
		}
		return nil
	}

	if err := backoff.Retry(fn, backOff); err != nil {
		return nil, err
	}
	return conn, nil
}

// isWSClosedError checks if error is websocket close error
func isWSClosedError(err error) bool {
	if err != nil {
		var websocketCloseErr *websocket.CloseError
		asWS := errors.As(err, &websocketCloseErr)
		isEOF := errors.Is(err, io.EOF)
		if asWS || isEOF {
			return true
		}

		if utils.Contains([]string{
			"connection reset by peer",
			"already wrote close",
			"broken pipe",
			"EOF",
		}, err.Error()) {
			return true
		}
	}
	return false
}

func randomID() jsonrpc2.ID {
	return jsonrpc2.ID{
		Str:      strconv.FormatUint(rand.New(rand.NewSource(time.Now().UnixNano())).Uint64(), 10),
		IsString: true,
	}
}
