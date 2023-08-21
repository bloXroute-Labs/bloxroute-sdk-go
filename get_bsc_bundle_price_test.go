package bloxroute_sdk_go

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetBscBundlePrice(t *testing.T) {
	t.Run("ws_cloud_api", testGetBscBundlePrice(wsCloudApiUrl))
	time.Sleep(5 * time.Second) // give the ws conn time to close
}

func testGetBscBundlePrice(url testURL) func(t *testing.T) {
	return func(t *testing.T) {
		config := testConfig(t, url)

		c, err := NewClient(context.Background(), config)
		require.NoError(t, err)

		_, err = c.GetBscBundlePrice(context.Background())

		require.NoError(t, err)
		require.NoError(t, c.Close())
	}
}
