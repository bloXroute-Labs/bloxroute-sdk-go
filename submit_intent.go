package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
	"fmt"

	pb "github.com/bloXroute-Labs/gateway/v2/protobuf"
)

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

// SubmitIntentReply is the reply from the SubmitIntent method
type SubmitIntentReply struct {
	// IntentID is the UUID of the intent
	IntentID string `json:"intentId"`

	// FirstSeen is the timestamp when intent was first seen in BDN (for now empty)
	FirstSeen string `json:"first_seen"`
}

// SubmitIntent submits an intent to the BDN
func (c *Client) SubmitIntent(ctx context.Context, params *SubmitIntentParams) (*SubmitIntentReply, error) {
	if c.handler.Type() != handlerSourceTypeGatewayGRPC {
		return nil, fmt.Errorf("submit intent is only supported with the gateway GRPC handler")
	}

	if params == nil {
		return nil, fmt.Errorf("params is required")
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

	req := &pb.SubmitIntentRequest{
		DappAddress:   params.DappAddress,
		SenderAddress: params.SenderAddress,
		Intent:        params.Intent,
		Hash:          params.Hash,
		Signature:     params.Signature,
	}

	res, err := c.handler.Request(ctx, RPCSubmitIntent, req)
	if err != nil {
		return nil, fmt.Errorf("failed to submit intent: %v", err)
	}

	var reply SubmitIntentReply
	err = json.Unmarshal(*res, &reply)
	if err != nil {
		return nil, fmt.Errorf("failed to parse submit intent response: %v", err)
	}

	return &reply, nil
}
