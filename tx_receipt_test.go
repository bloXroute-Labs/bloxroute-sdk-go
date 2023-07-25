package bloxroute_sdk_go

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOnTxReceipt(t *testing.T) {
	t.Run("ws_gateway", testOnTxReceipt(wsGatewayUrl))
}

func testOnTxReceipt(url testURL) func(t *testing.T) {
	return func(t *testing.T) {
		config := testConfig(t, url)

		c, err := NewClient(context.Background(), config)
		require.NoError(t, err)

		receive := make(chan struct{})

		err = c.OnTxReceipt(context.Background(), nil, func(ctx context.Context, err error, result *OnTxReceiptNotification) {
			require.NoError(t, err)
			close(receive)
		})
		require.NoError(t, err)

		// wait for the first tx receipt
		select {
		case <-receive:
		case <-time.After(time.Minute):
			require.Fail(t, "timeout waiting for tx receipt")
		}

		err = c.UnsubscribeFromNewTxs()
		require.NoError(t, err)

		require.NoError(t, c.Close())
	}
}
