package bloxroute_sdk_go

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSendPrivateTx(t *testing.T) {
	t.Run("ws_cloud_api", testSendPrivateTx(wsCloudApiUrl))
}

func testSendPrivateTx(url testURL) func(t *testing.T) {
	return func(t *testing.T) {
		config := testConfig(t, url)

		txBytes := os.Getenv("TX_BYTES")
		require.NotEmpty(t, txBytes)

		c, err := NewClient(context.Background(), config)
		require.NoError(t, err)

		sendTxParams := &SendPrivateTxParams{
			Transaction: txBytes,
		}

		_, err = c.SendPrivateTx(context.Background(), sendTxParams)
		require.NoError(t, err)
		require.NoError(t, c.Close())
	}
}
