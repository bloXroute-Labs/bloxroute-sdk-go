package bloxroute_sdk_go

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSendBscBundle(t *testing.T) {
	t.Run("ws_cloud_api", testSendBscBundle(wsCloudApiUrl))
	time.Sleep(5 * time.Second) // give the ws conn time to close
	t.Run("ws_gateway", testSendBscBundle(wsGatewayUrl))
}

func testSendBscBundle(url testURL) func(t *testing.T) {
	return func(t *testing.T) {
		config := testConfig(t, url)

		txBytes := os.Getenv("TX_BYTES")
		require.NotEmpty(t, txBytes)

		c, err := NewClient(context.Background(), config)
		require.NoError(t, err)

		sendBscBundleParams := &SendBscBundleParams{
			Transactions: []string{txBytes},
			BlockNumber:  "0x0",
		}

		_, err = c.SendBscBundle(context.Background(), sendBscBundleParams)
		require.NoError(t, err)

		require.NoError(t, c.Close())
	}
}
