package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	pb "github.com/bloXroute-Labs/gateway/v2/protobuf"
)

var ErrSubmitIntentSolutionGatewayOnly = errors.New("SubmitIntentSolution is only supported on the gateway GRPC and WS handlers")

// SubmitIntentSolutionParams is the parameters for submitting an intent solution
type SubmitIntentSolutionParams struct {

	// SolverAddress is the address of the solver
	SolverAddress string

	// IntentID is the UUID of the intent
	IntentID string

	// IntentSolution is the solution to the intent
	IntentSolution []byte

	// Hash is the Keccak256Hash of the intent solution
	Hash []byte

	// Signature is the ECDSA signature of the hash
	Signature []byte
}

// SubmitIntentSolution submits an intent solution to the BDN
func (c *Client) SubmitIntentSolution(ctx context.Context, params *SubmitIntentSolutionParams) (*json.RawMessage, error) {
	if c.handler.Type() != handlerSourceTypeGatewayGRPC && c.handler.Type() != handlerSourceTypeGatewayWS {
		return nil, ErrSubmitIntentSolutionGatewayOnly
	}

	if params == nil {
		return nil, ErrNilParams
	}

	if params.SolverAddress == "" {
		return nil, fmt.Errorf("solver address is required")
	}

	if params.IntentID == "" {
		return nil, fmt.Errorf("intent ID is required")
	}

	if len(params.IntentSolution) == 0 {
		return nil, fmt.Errorf("intent solution is required")
	}

	if len(params.Hash) == 0 {
		return nil, fmt.Errorf("hash is required")
	}

	if len(params.Signature) == 0 {
		return nil, fmt.Errorf("signature is required")
	}

	var req interface{}

	if c.handler.Type() == handlerSourceTypeGatewayGRPC {
		req = &pb.SubmitIntentSolutionRequest{
			SolverAddress:  params.SolverAddress,
			IntentId:       params.IntentID,
			IntentSolution: params.IntentSolution,
			Hash:           params.Hash,
			Signature:      params.Signature,
		}
	} else {
		req = &RPCSubmitIntentPayloadSolution{
			SolverAddress:  params.SolverAddress,
			IntentID:       params.IntentID,
			IntentSolution: params.IntentSolution,
			Hash:           params.Hash,
			Signature:      params.Signature,
		}
	}

	return c.handler.Request(ctx, RPCSubmitIntentSolution, req)
}
