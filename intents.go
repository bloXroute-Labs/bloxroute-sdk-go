package bloxroute_sdk_go

import (
	"context"
	"fmt"

	pb "github.com/bloXroute-Labs/gateway/v2/protobuf"
	"github.com/bloXroute-Labs/gateway/v2/types"
	"github.com/ethereum/go-ethereum/crypto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// IntentsParams is the params object for the OnIntents subscription
type IntentsParams struct {
	// SolverPrivateKey is the private key of your solver
	// Required if SolverAddress, Hash, and Signature are not provided
	SolverPrivateKey string

	// SolverAddress is the ETH address of the solver
	// Required if SolverPrivateKey is not provided
	SolverAddress string

	// Hash is the Keccak256Hash of the SolverAddress bytes
	// Required if SolverPrivateKey is not provided
	Hash []byte

	// Signature is the ECDSA signature of the Hash signed by the solver's private key
	// Required if SolverPrivateKey is not provided
	Signature []byte

	// FromTimestamp is an optional timestamp to specify the starting point for receiving intents
	FromTimestamp *timestamppb.Timestamp `json:"fromTimestamp"`
}

// OnIntentsNotification is the notification object for the OnIntents subscription
type OnIntentsNotification struct {
	DappAddress   string `json:"dappAddress"`
	SenderAddress string `json:"senderAddress"`
	IntentID      string `json:"intentID"`
	Intent        []byte `json:"intent"`
	Timestamp     string `json:"timestamp"`
}

// OnIntents subscribes to a stream of all new intents as they are propagated in the BDN.
func (c *Client) OnIntents(ctx context.Context, params *IntentsParams, callbackFunc CallbackFunc[*OnIntentsNotification]) error {

	if c.handler.Type() != handlerSourceTypeGatewayGRPC {
		return fmt.Errorf("OnIntents is only supported for with a GRPC handler")
	}

	if params == nil {
		return fmt.Errorf("params is required")
	}

	var solverAddress string
	var hash []byte
	var signature []byte

	if params.SolverPrivateKey != "" {
		privateKey, err := crypto.HexToECDSA(params.SolverPrivateKey)
		if err != nil {
			return fmt.Errorf("failed to parse solver private key: %v", err)
		}

		publicKey := privateKey.PublicKey
		solverAddress = crypto.PubkeyToAddress(publicKey).String()
		hash = crypto.Keccak256Hash([]byte(solverAddress)).Bytes()
		signature, err = crypto.Sign(hash, privateKey)
		if err != nil {
			return fmt.Errorf("failed to sign solver hash: %v", err)
		}
	} else {
		if params.SolverAddress == "" || len(params.Hash) == 0 || len(params.Signature) == 0 {
			return fmt.Errorf("solver address, hash, and signature are required when solver private key is not provided")
		}
		solverAddress = params.SolverAddress
		hash = params.Hash
		signature = params.Signature
	}

	wrap := func(ctx context.Context, err error, result any) {
		if err != nil {
			callbackFunc(ctx, err, nil)
			return
		}
		callbackFunc(ctx, nil, result.(*OnIntentsNotification))
	}

	req := &pb.IntentsRequest{
		SolverAddress: solverAddress,
		Hash:          hash,
		Signature:     signature,
		FromTimestamp: params.FromTimestamp,
	}
	return c.handler.Subscribe(ctx, types.UserIntentsFeed, req, wrap)
}

func (c *Client) UnsubscribeFromOnIntent() error {
	return c.handler.UnsubscribeRetry(types.UserIntentsFeed)
}
