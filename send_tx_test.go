package bloxroute_sdk_go

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSendTx(t *testing.T) {
	t.Run("ws_cloud_api", testSendTx(wsCloudApiUrl))
	t.Run("ws_gateway", testSendTx(wsGatewayUrl))
	t.Run("grpc_gateway", testSendTx(grpcGatewayUrl))
}

func testSendTx(url testURL) func(t *testing.T) {
	return func(t *testing.T) {
		config := testConfig(t, url)

		// get tx bytes from os env and error if not found
		txBytes := os.Getenv("TX_BYTES")
		require.NotEmpty(t, txBytes)

		c, err := NewClient(context.Background(), config)
		require.NoError(t, err)

		sendTxParams := &SendTxParams{
			Transaction: txBytes,
		}

		_, err = c.SendTx(context.Background(), sendTxParams)

		require.NoError(t, err)
		require.NoError(t, c.Close())
	}
}
