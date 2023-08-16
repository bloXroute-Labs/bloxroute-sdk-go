package bloxroute_sdk_go

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOnNewBlock(t *testing.T) {
	t.Run("ws_cloud_api", testOnNewBlock(wsCloudApiUrl))
	t.Run("ws_gateway", testOnNewBlock(wsGatewayUrl))
	t.Run("grpc_gateway", testOnNewBlock(grpcGatewayUrl))
}

func testOnNewBlock(url testURL) func(t *testing.T) {
	return func(t *testing.T) {
		config := testConfig(t, url)

		c, err := NewClient(context.Background(), config)
		require.NoError(t, err)

		receive := make(chan struct{})

		err = c.OnNewBlock(context.Background(), &NewBlockParams{Include: []string{}}, func(ctx context.Context, err error, result *OnBdnBlockNotification) {
			require.NoError(t, err)
			require.NotNilf(t, result, "result is nil")
			require.NotEmptyf(t, result.Hash, "hash is empty")
			require.NotEmptyf(t, result.Header, "header is empty")

			close(receive)
		})
		require.NoError(t, err)

		select {
		case <-receive:
		case <-time.After(time.Minute):
			require.Fail(t, "timeout waiting for block")
		}

		err = c.UnsubscribeFromOnNewBlock()
		require.NoError(t, err)
		require.NoError(t, c.Close())
	}
}
