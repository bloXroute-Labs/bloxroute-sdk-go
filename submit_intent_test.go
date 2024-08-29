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

func TestSubmitIntent(t *testing.T) {
	t.Run("grpc_gateway", testSubmitIntent(grpcGatewayUrl))
	t.Run("ws_gateway", testSubmitIntent(wsGatewayUrl))
}

func testSubmitIntent(url testURL) func(t *testing.T) {
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
		assert.NotEmpty(t, submitIntentResponse["intent_id"])
	}
}

func createdSubmitIntentParams(t *testing.T) *SubmitIntentParams {
	// get the solver private key from the environment variable
	senderPrivateKey := os.Getenv("SENDER_PRIVATE_KEY")
	if senderPrivateKey == "" {
		// Generate a new private key if the environment variable is not set
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		senderPrivateKey = hex.EncodeToString(crypto.FromECDSA(privateKey))
	}

	// get dApp private key from the environment variable
	dAppPrivateKeyRaw := os.Getenv("DAPP_PRIVATE_KEY")
	if dAppPrivateKeyRaw == "" {
		// Generate a new private key if the environment variable is not set
		dAppPrivateKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		dAppPrivateKeyRaw = hex.EncodeToString(crypto.FromECDSA(dAppPrivateKey))
	}

	return &SubmitIntentParams{
		DappPrivateKey:   dAppPrivateKeyRaw,
		SenderPrivateKey: senderPrivateKey,
		Intent:           []byte("test intent"),
	}
}
