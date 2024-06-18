package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	pb "github.com/bloXroute-Labs/gateway/v2/protobuf"
)

var ErrSubmitIntentGatewayOnly = errors.New("SubmitIntent is only supported on the gateway GRPC and WS handlers")

// SubmitIntentParams is the parameters for submitting an intent
type SubmitIntentParams struct {
	// DappAddress is the ETH address of the DApp that should receive solution for this intent
	DappAddress string

	// SenderAddress is the ETH address of the intent sender
	SenderAddress string

	// Intent is the intent payload
	Intent []byte

	// Hash is the Keccak256Hash of the intent payload
	Hash []byte

	// Signature is the ECDSA signature of the hash
	Signature []byte
}

// SubmitIntent submits an intent to the BDN
func (c *Client) SubmitIntent(ctx context.Context, params *SubmitIntentParams) (*json.RawMessage, error) {
	if c.handler.Type() != handlerSourceTypeGatewayGRPC && c.handler.Type() != handlerSourceTypeGatewayWS {
		return nil, ErrSubmitIntentGatewayOnly
	}

	if params == nil {
		return nil, ErrNilParams
	}

	if params.DappAddress == "" {
		return nil, fmt.Errorf("dapp address is required")
	}

	if params.SenderAddress == "" {
		return nil, fmt.Errorf("sender address is required")
	}

	if len(params.Intent) == 0 {
		return nil, fmt.Errorf("intent is required")
	}

	if len(params.Hash) == 0 {
		return nil, fmt.Errorf("hash is required")
	}

	if len(params.Signature) == 0 {
		return nil, fmt.Errorf("signature is required")
	}

	var req interface{}

	if c.handler.Type() == handlerSourceTypeGatewayGRPC {
		req = &pb.SubmitIntentRequest{
			DappAddress:   params.DappAddress,
			SenderAddress: params.SenderAddress,
			Intent:        params.Intent,
			Hash:          params.Hash,
			Signature:     params.Signature,
		}
	} else {
		req = &RPCSubmitIntentPayload{
			DappAddress:   params.DappAddress,
			SenderAddress: params.SenderAddress,
			Intent:        params.Intent,
			Hash:          params.Hash,
			Signature:     params.Signature,
		}
	}

	return c.handler.Request(ctx, RPCSubmitIntent, req)
}
