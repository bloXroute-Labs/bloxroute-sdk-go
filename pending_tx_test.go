package bloxroute_sdk_go

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOnPendingTx(t *testing.T) {
	t.Run("ws_cloud_api", testOnPendingTx(wsCloudApiUrl))
	t.Run("ws_gateway", testOnPendingTx(wsGatewayUrl))
	t.Run("grpc_gateway", testOnPendingTx(grpcGatewayUrl))
}

func testOnPendingTx(url testURL) func(t *testing.T) {
	return func(t *testing.T) {
		config := testConfig(t, url)

		c, err := NewClient(context.Background(), config)
		require.NoError(t, err)

		receive := make(chan struct{})

		err = c.OnPendingTx(context.Background(), &PendingTxParams{Include: []string{"raw_tx"}}, func(ctx context.Context, err error, result *NewTxNotification) {
			select {
			case <-receive:
				return
			default:
			}

			require.NoError(t, err)
			require.NotNilf(t, result, "result is nil")
			require.NotEmptyf(t, result.RawTx, "raw tx is empty")

			close(receive)
		})
		require.NoError(t, err)

		// wait for the first new tx
		select {
		case <-receive:
		case <-time.After(10 * time.Second):
			require.Fail(t, "timeout waiting for new tx")
		}

		err = c.UnsubscribeFromPendingTxs()
		require.NoError(t, err)
		require.NoError(t, c.Close())
	}
}
