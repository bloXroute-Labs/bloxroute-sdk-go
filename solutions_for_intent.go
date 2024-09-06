package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/bloXroute-Labs/gateway/v2/jsonrpc"
	"github.com/ethereum/go-ethereum/crypto"
)

var ErrGetSolutionsForIntentsGatewayOnly = errors.New("GetSolutionsForIntents is only supported on the gateway WS handler")

// GetSolutionsForIntentParams is the params object for the GetSolutionsForIntent subscription
type GetSolutionsForIntentParams struct {
	// DAppOrSenderPrivateKey is the private key of your DApp used to prove the ownership of the DApp address
	// Required if DappAddress, Hash, and Signature are not provided
	DAppOrSenderPrivateKey string

	// IntentID is the ID of the intent for which solutions are requested
	IntentID string

	// DappAddress is the ETH address of the DApp that should receive solutions
	// Required if DAppOrSenderPrivateKey is not provided
	DAppOrSenderAddress string

	// Hash is the Keccak256Hash of the DappAddress bytes
	// Required if DAppOrSenderPrivateKey is not provided
	Hash []byte

	// Signature is the ECDSA signature of the Hash signed by the DApp private key
	// Required if DAppOrSenderPrivateKey is not provided
	Signature []byte
}

// OnSolutionsForIntentNotification is the notification object for the GetSolutionsForIntent request
type OnSolutionsForIntentNotification struct {
	ID            string    // UUIDv4
	SolverAddress string    // ETH Address
	IntentID      string    // UUIDv4
	Solution      []byte    // Variable length
	Hash          []byte    // Keccak256
	Signature     []byte    // ECDSA Signature
	Timestamp     time.Time // Short timestamp
	DappAddress   string    // ETH Address
}

// GetSolutionsForIntent submits a request to get solutions for a specific intent
func (c *Client) GetSolutionsForIntent(ctx context.Context, params *GetSolutionsForIntentParams) (*json.RawMessage, error) {
	if c.handler.Type() != handlerSourceTypeGatewayWS {
		return nil, ErrGetSolutionsForIntentsGatewayOnly
	}

	if params == nil {
		return nil, ErrNilParams
	}

	if params.IntentID == "" {
		return nil, ErrIntentIDRequired
	}

	var dappAddress string
	var hash []byte
	var signature []byte

	if params.DAppOrSenderPrivateKey != "" {
		privateKey, err := crypto.HexToECDSA(params.DAppOrSenderPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse dapp private key: %v", err)
		}

		publicKey := privateKey.PublicKey
		dappAddress = crypto.PubkeyToAddress(publicKey).String()
		hash = crypto.Keccak256Hash([]byte(dappAddress + params.IntentID)).Bytes()
		signature, err = crypto.Sign(hash, privateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to sign dapp hash: %v", err)
		}
	} else {
		if params.DAppOrSenderAddress == "" || len(params.Hash) == 0 || len(params.Signature) == 0 {
			return nil, fmt.Errorf("dapp address, hash, and signature are required when dapp private key is not provided")
		}
		dappAddress = params.DAppOrSenderAddress
		hash = params.Hash
		signature = params.Signature
	}

	req := map[string]interface{}{
		"dapp_or_sender_address": dappAddress,
		"hash":                   hash,
		"signature":              signature,
		"intent_id":              params.IntentID,
	}

	return c.handler.Request(ctx, jsonrpc.RPCGetIntentSolutions, req)
}
