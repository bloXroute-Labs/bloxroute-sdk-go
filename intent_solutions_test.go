package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOnIntentSolutions(t *testing.T) {
	t.Run("ws_gateway_dapp_addr", testOnIntentSolutionsWithDappAddr(wsGatewayUrl))
	t.Run("ws_gateway_sender_addr", testOnIntentSolutionsWithSenderAddr(wsGatewayUrl))
	t.Run("grpc_gateway", testOnIntentSolutionsWithDappAddr(grpcGatewayUrl))
}

func testOnIntentSolutionsWithDappAddr(url testURL) func(t *testing.T) {
	return testOnIntentSolutions(url, true)
}

func testOnIntentSolutionsWithSenderAddr(url testURL) func(t *testing.T) {
	return testOnIntentSolutions(url, false)
}

func testOnIntentSolutions(url testURL, useDApp bool) func(t *testing.T) {
	return func(t *testing.T) {
		config := testConfig(t, url)

		ctx := contextWithSignal(context.Background())

		c, err := NewClient(ctx, config)
		require.NoError(t, err)

		receive := make(chan struct{})
		submitIntentParams := createdSubmitIntentParams(t)

		subscriptionParams := &IntentSolutionsParams{}
		if useDApp {
			subscriptionParams.DappPrivateKey = submitIntentParams.DappPrivateKey
		} else {
			subscriptionParams.DappPrivateKey = submitIntentParams.SenderPrivateKey
		}

		err = c.OnIntentSolutions(ctx, subscriptionParams, func(ctx context.Context, err error, result *OnIntentSolutionsNotification) {
			require.NoError(t, err)
			require.NotNilf(t, result, "result is nil")
			require.NotEmptyf(t, result.IntentID, "intent ID is empty")
			require.NotEmptyf(t, result.IntentSolution, "intent solution is empty")
			close(receive)
		})
		require.NoError(t, err)

		subRep, err := c.SubmitIntent(ctx, submitIntentParams)
		require.NoError(t, err)

		var resp map[string]string
		err = json.Unmarshal(*subRep, &resp)
		require.NoError(t, err)
		require.NotEmpty(t, resp["intent_id"])

		_, err = c.SubmitIntentSolution(ctx, createSubmitIntentSolutionParams(t, resp["intent_id"], []byte("test intent solution")))
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
