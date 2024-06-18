package bloxroute_sdk_go

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOnIntentSolutions(t *testing.T) {
	t.Run("ws_gateway", testOnIntentSolutions(wsGatewayUrl))
	t.Run("grpc_gateway", testOnIntentSolutions(grpcGatewayUrl))
}

func testOnIntentSolutions(url testURL) func(t *testing.T) {
	return func(t *testing.T) {
		config := testConfig(t, url)

		// Get the DApp private key from the environment variable
		dappPrivateKey := os.Getenv("DAPP_PRIVATE_KEY")
		require.NotEmpty(t, dappPrivateKey)

		c, err := NewClient(context.Background(), config)
		require.NoError(t, err)

		receive := make(chan struct{})

		params := &IntentSolutionsParams{
			DappPrivateKey: dappPrivateKey,
		}

		err = c.OnIntentSolutions(context.Background(), params, func(ctx context.Context, err error, result *OnIntentSolutionsNotification) {
			require.NoError(t, err)
			require.NotNilf(t, result, "result is nil")
			require.NotEmptyf(t, result.IntentID, "intent ID is empty")
			require.NotEmptyf(t, result.IntentSolution, "intent solution is empty")
			close(receive)
		})
		require.NoError(t, err)

		_, err = submitIntentSolutionTest(context.Background(), t, c)
		require.NoError(t, err)

		select {
		case <-receive:
		case <-time.After(time.Minute):
			require.Fail(t, "timeout waiting for intent solution")
		}

		err = c.UnsubscribeFromOnIntentSolutions()
		require.NoError(t, err)
		require.NoError(t, c.Close())
	}
}
