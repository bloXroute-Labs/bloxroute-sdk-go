package bloxroute_sdk_go

import (
	"context"
	_ "embed"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/bloXroute-Labs/bloxroute-sdk-go/connection/ws"
	"github.com/cenkalti/backoff/v4"
	"github.com/sourcegraph/jsonrpc2"
)

//go:embed version.txt
var buildVersion string

func reconnect(ctx context.Context, url string, headers http.Header, opts *ws.DialOptions) (ws.Conn, error) {
	backOff := backoff.NewExponentialBackOff()
	backOff.MaxElapsedTime = reconnectTimeout
	backOff.InitialInterval = reconnectInitialInterval

	var conn ws.Conn

	fn := func() error {
		var err error
		conn, err = ws.Dial(ctx, url, headers, opts)
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
