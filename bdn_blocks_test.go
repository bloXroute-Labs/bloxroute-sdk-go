package bloxroute_sdk_go

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOnBdnBlock(t *testing.T) {
	t.Run("ws_cloud_api", testOnBdnBlock(wsCloudApiUrl))
	t.Run("ws_gateway", testOnBdnBlock(wsGatewayUrl))
	t.Run("grpc_gateway", testOnBdnBlock(grpcGatewayUrl))
}

func testOnBdnBlock(url testURL) func(t *testing.T) {
	return func(t *testing.T) {
		config := testConfig(t, url)

		c, err := NewClient(context.Background(), config)
		require.NoError(t, err)

		receive := make(chan struct{})

		err = c.OnBdnBlock(context.Background(), &BdnBlockParams{Include: []string{"hash", "header", "transactions"}}, func(ctx context.Context, err error, result *OnBdnBlockNotification) {
			require.NoError(t, err)
			require.NotNilf(t, result, "result is nil")
			require.NotEmptyf(t, result.Hash, "hash is empty")
			require.NotEmptyf(t, result.Header, "header is empty")
			require.Truef(t, len(result.Transactions) > 0, "transactions are not empty")
			close(receive)
		})
		require.NoError(t, err)

		// wait for the first BDN block
		select {
		case <-receive:
		case <-time.After(time.Minute):
			require.Fail(t, "timeout waiting for BDN block")
		}

		err = c.UnsubscribeFromBdnBlock()
		require.NoError(t, err)
		require.NoError(t, c.Close())
	}
}
