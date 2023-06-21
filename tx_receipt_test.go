package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOnTxReceipt(t *testing.T) {
	t.Run("gateway", testOnTxReceipt(t, "GATEWAY_URL"))
}

func testOnTxReceipt(t *testing.T, url string) func(t *testing.T) {
	return func(t *testing.T) {
		config := &Config{
			AuthHeader: os.Getenv("AUTH_HEADER"),
			GatewayURL: os.Getenv("GATEWAY_URL"),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		defer cancel()

		c, err := NewClient(config)
		require.NoError(t, err)

		wg := &sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer wg.Done()
			err := c.Run(ctx)
			require.NoError(t, err)
		}()

		receive := make(chan struct{})

		err = c.OnTxReceipt(ctx, nil, func(ctx context.Context, result *json.RawMessage) {
			close(receive)
		})
		require.NoError(t, err)

		// wait for the first tx receipt
		select {
		case <-receive:
		case <-time.After(10 * time.Second):
			require.Fail(t, "timeout waiting for tx receipt")
		}

		err = c.UnsubscribeFromNewTxs()
		require.NoError(t, err)

		require.NoError(t, c.Close())
		wg.Wait()
	}
}
