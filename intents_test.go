package bloxroute_sdk_go

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnIntents(t *testing.T) {
	t.Run("ws_gateway", testOnIntents(wsGatewayUrl))
	t.Run("ws_gateway_with_filter", testOnIntentsWithFilter(wsGatewayUrl))
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

		params := &IntentsParams{
			SolverPrivateKey: solverPrivateKey,
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

		subRep, err := c.SubmitIntent(context.Background(), createdSubmitIntentParams(t))
		require.NoError(t, err)

		var resp map[string]string
		err = json.Unmarshal(*subRep, &resp)
		require.NoError(t, err)
		require.NotEmpty(t, resp["intent_id"])

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

func testOnIntentsWithFilter(url testURL) func(t *testing.T) {
	return func(t *testing.T) {
		config := testConfig(t, url)

		c, err := NewClient(context.Background(), config)
		require.NoError(t, err)

		intent1 := createdSubmitIntentParams(t)
		intent2 := createdSubmitIntentParams(t)

		solverPrivateKey := os.Getenv("SOLVER_PRIVATE_KEY")
		if solverPrivateKey == "" {
			privateKey, err := crypto.GenerateKey()
			require.NoError(t, err)
			solverPrivateKey = hex.EncodeToString(crypto.FromECDSA(privateKey))
		}

		privateKey, err := crypto.HexToECDSA(intent2.DappPrivateKey)
		require.NoError(t, err)
		params := &IntentsParams{
			SolverPrivateKey: solverPrivateKey,
			DappAddress:      crypto.PubkeyToAddress(privateKey.PublicKey).String(),
		}

		receive := make(chan struct{})

		err = c.OnIntents(context.Background(), params, func(ctx context.Context, err error, result *OnIntentsNotification) {
			require.NoError(t, err)
			assert.NotNilf(t, result, "result is nil")
			assert.NotEmptyf(t, result.DappAddress, "dapp address is empty")
			assert.NotEmptyf(t, result.SenderAddress, "sender address is empty")
			assert.NotEmptyf(t, result.IntentID, "intent ID is empty")
			assert.NotEmptyf(t, result.Intent, "intent is empty")
			assert.NotEmptyf(t, result.Timestamp, "timestamp is empty")
			assert.NotEqual(t, intent1.DappAddress, result.DappAddress, "should not receive intent for dapp address %s", intent1.DappAddress)
			assert.Equal(t, intent2.DappAddress, result.DappAddress, "should receive intent for dapp address %s", intent2.DappAddress)
			close(receive)
		})
		require.NoError(t, err)

		_, err = c.SubmitIntent(context.Background(), intent1)
		require.NoError(t, err)

		_, err = c.SubmitIntent(context.Background(), intent2)
		require.NoError(t, err)

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
