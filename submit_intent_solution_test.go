package bloxroute_sdk_go

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestSubmitIntentSolution(t *testing.T) {
	t.Run("grpc_gateway", testSubmitIntentSolution(grpcGatewayUrl))
	t.Run("ws_gateway", testSubmitIntentSolution(wsGatewayUrl))
}

func testSubmitIntentSolution(url testURL) func(t *testing.T) {
	return func(t *testing.T) {
		config := testConfig(t, url)

		ctx := context.Background()

		c, err := NewClient(ctx, config)
		require.NoError(t, err)

		resp, err := c.SubmitIntent(ctx, createdSubmitIntentParams(t))
		require.NoError(t, err)
		require.NoError(t, c.Close())

		var submitIntentResponse map[string]string
		err = json.Unmarshal(*resp, &submitIntentResponse)
		require.NoError(t, err)
		require.NotEmpty(t, submitIntentResponse["intent_id"])

		_, err = c.SubmitIntentSolution(ctx, createSubmitIntentSolutionParams(t, submitIntentResponse["intent_id"], []byte("test intent solution")))
		require.NoError(t, err)
		require.NoError(t, c.Close())
	}
}

func createSubmitIntentSolutionParams(t *testing.T, intentID string, intentSolution []byte) *SubmitIntentSolutionParams {
	solverPrivateKey := os.Getenv("SOLVER_PRIVATE_KEY")
	if solverPrivateKey == "" {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		solverPrivateKey = hex.EncodeToString(crypto.FromECDSA(privateKey))
	}

	return &SubmitIntentSolutionParams{
		SolverPrivateKey: solverPrivateKey,
		IntentID:         intentID,
		IntentSolution:   intentSolution,
	}
}
