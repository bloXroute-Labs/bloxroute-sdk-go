package bloxroute_sdk_go

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSendEthBundle(t *testing.T) {
	t.Run("ws_cloud_api", testSendEthBundle(wsCloudApiUrl))
	time.Sleep(5 * time.Second) // give the websocket conn time to close
	t.Run("ws_gateway", testSendEthBundle(wsGatewayUrl))
}

func testSendEthBundle(url testURL) func(t *testing.T) {
	return func(t *testing.T) {
		config := testConfig(t, url)

		txBytes := os.Getenv("TX_BYTES")
		require.NotEmpty(t, txBytes)

		c, err := NewClient(context.Background(), config)
		require.NoError(t, err)

		sendEthBundleParams := &SendEthBundleParams{
			Transactions: []string{txBytes},
			BlockNumber:  "0x0",
		}

		_, err = c.SendEthBundle(context.Background(), sendEthBundleParams)
		require.NoError(t, err)

		require.NoError(t, c.Close())
	}
}
