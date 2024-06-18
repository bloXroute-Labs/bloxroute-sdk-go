package bloxroute_sdk_go

import (
	"context"
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

		c, err := NewClient(context.Background(), config)
		require.NoError(t, err)

		_, err = submitIntentSolutionTest(context.Background(), t, c)
		require.NoError(t, err)
		require.NoError(t, c.Close())
	}
}

func submitIntentSolutionTest(ctx context.Context, t *testing.T, c *Client) (*json.RawMessage, error) {

	// Get the solver private key from the environment variable
	solverPrivateKey := os.Getenv("SOLVER_PRIVATE_KEY")
	require.NotEmpty(t, solverPrivateKey)

	privateKey, err := crypto.HexToECDSA(solverPrivateKey)
	if err != nil {
		return nil, err
	}

	publicKey := privateKey.PublicKey
	solverAddress := crypto.PubkeyToAddress(publicKey).Hex()

	// intent solution
	intentID := "test-intent-solution-id"
	intentSolution := []byte("test intent solution")

	hash := crypto.Keccak256Hash(intentSolution).Bytes()
	signature, err := crypto.Sign(hash, privateKey)
	if err != nil {
		return nil, err
	}

	params := &SubmitIntentSolutionParams{
		SolverAddress:  solverAddress,
		IntentID:       intentID,
		IntentSolution: intentSolution,
		Hash:           hash,
		Signature:      signature,
	}

	return c.SubmitIntentSolution(ctx, params)
}
