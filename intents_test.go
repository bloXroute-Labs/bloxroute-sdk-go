package bloxroute_sdk_go

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestOnIntents(t *testing.T) {
	t.Run("grpc_gateway", testOnIntents(grpcGatewayUrl))
}

func testOnIntents(url testURL) func(t *testing.T) {
	return func(t *testing.T) {
		config := testConfig(t, url)

		// Get the solver private key from the environment variable
		solverPrivateKey := os.Getenv("SOLVER_PRIVATE_KEY")
		require.NotEmpty(t, solverPrivateKey)

		c, err := NewClient(context.Background(), config)
		require.NoError(t, err)

		receive := make(chan struct{})

		// Get the current timestamp
		currentTimestamp := time.Now()

		params := &IntentsParams{
			SolverPrivateKey: solverPrivateKey,
			FromTimestamp:    timestamppb.New(currentTimestamp),
		}

		err = c.OnIntents(context.Background(), params, func(ctx context.Context, err error, result *OnIntentsNotification) {
			require.NoError(t, err)
			require.NotNilf(t, result, "result is nil")
			require.NotEmptyf(t, result.DappAddress, "dapp address is empty")
			require.NotEmptyf(t, result.SenderAddress, "sender address is empty")
			require.NotEmptyf(t, result.IntentID, "intent ID is empty")
			require.NotEmptyf(t, result.Intent, "intent is empty")
			require.NotEmptyf(t, result.Timestamp, "timestamp is empty")

			close(receive)
		})
		require.NoError(t, err)

		subRep, err := submitTestIntent(context.Background(), t, c)
		require.NoError(t, err)
		require.NotEmpty(t, subRep.IntentID)

		select {
		case <-receive:
		case <-time.After(time.Minute):
			require.Fail(t, "timeout waiting for intent")
		}

		err = c.UnsubscribeFromOnIntent()
		require.NoError(t, err)
		require.NoError(t, c.Close())
	}
}
