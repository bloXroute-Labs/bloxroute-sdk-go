package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
	"fmt"

	pb "github.com/bloXroute-Labs/gateway/v2/protobuf"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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

// SubmitIntentSolutionReply is the reply from the SubmitIntentSolution method
type SubmitIntentSolutionReply struct {
	SolutionID string                 `json:"solutionId"`
	FirstSeen  *timestamppb.Timestamp `json:"first_seen"`
}

// SubmitIntentSolution submits an intent solution to the BDN
func (c *Client) SubmitIntentSolution(ctx context.Context, params *SubmitIntentSolutionParams) (*SubmitIntentSolutionReply, error) {
	if c.handler.Type() != handlerSourceTypeGatewayGRPC {
		return nil, fmt.Errorf("submit intent solution is only supported with the gateway GRPC handler")
	}

	if params == nil {
		return nil, fmt.Errorf("params is required")
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

	req := &pb.SubmitIntentSolutionRequest{
		SolverAddress:  params.SolverAddress,
		IntentId:       params.IntentID,
		IntentSolution: params.IntentSolution,
		Hash:           params.Hash,
		Signature:      params.Signature,
	}

	res, err := c.handler.Request(ctx, RPCSubmitIntentSolution, req)
	if err != nil {
		return nil, fmt.Errorf("failed to submit intent solution: %v", err)
	}

	var reply SubmitIntentSolutionReply
	err = json.Unmarshal(*res, &reply)
	if err != nil {
		return nil, fmt.Errorf("failed to parse submit intent solution response: %v", err)
	}

	return &reply, nil
}
