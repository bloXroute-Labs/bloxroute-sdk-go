package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestSubmitIntent(t *testing.T) {
	t.Run("grpc_gateway", testSubmitIntent(grpcGatewayUrl))
	t.Run("ws_gateway", testSubmitIntent(wsGatewayUrl))
}

func testSubmitIntent(url testURL) func(t *testing.T) {
	return func(t *testing.T) {
		config := testConfig(t, url)

		c, err := NewClient(context.Background(), config)
		require.NoError(t, err)

		_, err = submitTestIntent(context.Background(), t, c)
		require.NoError(t, err)
		require.NoError(t, c.Close())
	}
}

func submitTestIntent(ctx context.Context, t *testing.T, c *Client) (*json.RawMessage, error) {

	// Get the solver private key from the environment variable
	solverPrivateKey := os.Getenv("SOLVER_PRIVATE_KEY")
	require.NotEmpty(t, solverPrivateKey)

	// Convert the solver private key from hex to an ECDSA private key
	solverPrivKey, err := crypto.HexToECDSA(solverPrivateKey)
	if err != nil {
		return nil, err
	}

	// Derive the solver address from the solver private key
	solverAddress := crypto.PubkeyToAddress(solverPrivKey.PublicKey)

	intent := []byte("test intent")

	// Hash the intent payload
	hash := crypto.Keccak256Hash(intent).Bytes()

	// Sign the hash using the solver's private key
	signature, err := crypto.Sign(hash, solverPrivKey)

	if err != nil {
		return nil, err
	}

	// Submit the intent
	submitIntentParams := &SubmitIntentParams{
		DappAddress:   solverAddress.Hex(),
		SenderAddress: solverAddress.Hex(),
		Intent:        intent,
		Hash:          hash,
		Signature:     signature,
	}

	return c.SubmitIntent(ctx, submitIntentParams)
}
