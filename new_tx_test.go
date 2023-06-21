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

func TestOnNewTx(t *testing.T) {
	t.Run("cloud api", testOnNewTx(t, "CLOUD_API_URL"))
	t.Run("gateway", testOnNewTx(t, "GATEWAY_URL"))
}

func testOnNewTx(t *testing.T, url string) func(t *testing.T) {
	return func(t *testing.T) {
		config := &Config{
			AuthHeader: os.Getenv("AUTH_HEADER"),
		}
		switch url {
		case "CLOUD_API_URL":
			config.CloudAPIURL = os.Getenv("CLOUD_API_URL")
		case "GATEWAY_URL":
			config.GatewayURL = os.Getenv("GATEWAY_URL")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		defer cancel()

		c, err := NewClient(config)
		require.NoError(t, err)

		started := make(chan struct{})
		wg := &sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer wg.Done()
			close(started)
			err := c.Run(ctx)
			require.NoError(t, err)
		}()

		<-started

		receive := make(chan struct{})

		err = c.OnNewTx(ctx, nil, func(ctx context.Context, result *json.RawMessage) {
			close(receive)
		})
		require.NoError(t, err)

		// wait for the first new tx
		select {
		case <-receive:
		case <-time.After(10 * time.Second):
			require.Fail(t, "timeout waiting for new tx")
		}

		err = c.UnsubscribeFromNewTxs()
		require.NoError(t, err)

		require.NoError(t, c.Close())
		wg.Wait()
	}
}
