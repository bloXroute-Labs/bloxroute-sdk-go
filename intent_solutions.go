package bloxroute_sdk_go

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/bloXroute-Labs/gateway/v2/protobuf"
	"github.com/bloXroute-Labs/gateway/v2/types"
	"github.com/ethereum/go-ethereum/crypto"
)

var ErrIntentsSolutionsGatewayOnly = errors.New("OnIntentSolutions is only supported on the gateway GRPC and WS handlers")

// IntentSolutionsParams is the params object for the OnIntentSolutions subscription
type IntentSolutionsParams struct {
	// DappPrivateKey is the private key of your DApp used to prove the ownership of the DApp address
	// Required if DappAddress, Hash, and Signature are not provided
	DappPrivateKey string

	// DappAddress is the ETH address of the DApp that should receive solutions
	// Required if DappPrivateKey is not provided
	DappAddress string

	// Hash is the Keccak256Hash of the DappAddress bytes
	// Required if DappPrivateKey is not provided
	Hash []byte

	// Signature is the ECDSA signature of the Hash signed by the DApp private key
	// Required if DappPrivateKey is not provided
	Signature []byte
}

// OnIntentSolutionsNotification is the notification object for the OnIntentSolutions subscription
type OnIntentSolutionsNotification struct {
	IntentID       string `json:"intentID"`
	IntentSolution []byte `json:"intentSolution"`
}

// OnIntentSolutions subscribes to a stream of new solutions that match the dappAddress of the subscription as they are propagated in the BDN.
func (c *Client) OnIntentSolutions(ctx context.Context, params *IntentSolutionsParams, callbackFunc CallbackFunc[*OnIntentSolutionsNotification]) error {
	if c.handler.Type() != handlerSourceTypeGatewayGRPC && c.handler.Type() != handlerSourceTypeGatewayWS {
		return ErrIntentsSolutionsGatewayOnly
	}

	if params == nil {
		return ErrNilParams
	}

	var dappAddress string
	var hash []byte
	var signature []byte

	if params.DappPrivateKey != "" {
		privateKey, err := crypto.HexToECDSA(params.DappPrivateKey)
		if err != nil {
			return fmt.Errorf("failed to parse dapp private key: %v", err)
		}

		publicKey := privateKey.PublicKey
		dappAddress = crypto.PubkeyToAddress(publicKey).String()
		hash = crypto.Keccak256Hash([]byte(dappAddress)).Bytes()
		signature, err = crypto.Sign(hash, privateKey)
		if err != nil {
			return fmt.Errorf("failed to sign dapp hash: %v", err)
		}
	} else {
		if params.DappAddress == "" || len(params.Hash) == 0 || len(params.Signature) == 0 {
			return fmt.Errorf("dapp address, hash, and signature are required when dapp private key is not provided")
		}
		dappAddress = params.DappAddress
		hash = params.Hash
		signature = params.Signature
	}

	wrap := func(ctx context.Context, err error, result any) {
		if err != nil {
			callbackFunc(ctx, err, nil)
			return
		}
		msg := result.(*OnIntentSolutionsNotification)
		callbackFunc(ctx, nil, msg)
	}

	var req interface{}

	if c.handler.Type() == handlerSourceTypeGatewayGRPC {
		req = &pb.IntentSolutionsRequest{
			DappAddress: dappAddress,
			Hash:        hash,
			Signature:   signature,
		}
	} else {
		req = map[string]interface{}{
			"dapp_address": dappAddress,
			"hash":         hash,
			"signature":    signature,
		}
	}

	return c.handler.Subscribe(ctx, types.UserIntentSolutionsFeed, req, wrap)
}

func (c *Client) UnsubscribeFromOnIntentSolutions() error {
	return c.handler.UnsubscribeRetry(types.UserIntentSolutionsFeed)
}
