package bloxroute_sdk_go

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubmitQuote(t *testing.T) {
	t.Run("grpc_gateway", testSubmitQuote(grpcGatewayUrl))
	t.Run("ws_gateway", testSubmitQuote(wsGatewayUrl))
}

func testSubmitQuote(url testURL) func(t *testing.T) {
	return func(t *testing.T) {
		config := testConfig(t, url)

		ctx := context.Background()
		c, err := NewClient(ctx, config)
		require.NoError(t, err)

		resp, err := c.SubmitQuote(ctx, createdSubmitQuoteParams(t))
		require.NoError(t, err)
		require.NoError(t, c.Close())

		var submitQuoteResponse map[string]string
		err = json.Unmarshal(*resp, &submitQuoteResponse)
		require.NoError(t, err)
		assert.NotEmpty(t, submitQuoteResponse["quote_id"])
	}
}

func createdSubmitQuoteParams(t *testing.T) *SubmitQuoteParams {
	solverPrivateKey := os.Getenv("SOLVER_PRIVATE_KEY")
	if solverPrivateKey == "" {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		solverPrivateKey = hex.EncodeToString(crypto.FromECDSA(privateKey))
	}

	dAppPrivateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	return &SubmitQuoteParams{
		DappAddress:      crypto.PubkeyToAddress(dAppPrivateKey.PublicKey).Hex(),
		SolverPrivateKey: solverPrivateKey,
		Quote:            []byte("0x1234567890"),
	}
}
