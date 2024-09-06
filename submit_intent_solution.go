package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/bloXroute-Labs/gateway/v2/jsonrpc"
	pb "github.com/bloXroute-Labs/gateway/v2/protobuf"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	ErrSubmitIntentSolutionGatewayOnly    = errors.New("SubmitIntentSolution is only supported on the gateway GRPC and WS handlers")
	ErrIntentIDRequired                   = errors.New("intent ID is required")
	ErrIntentSolutionRequired             = errors.New("intent solution is required")
	ErrSubmitIntentSolutionParamsRequired = fmt.Errorf("solver address, hash, and signature are required when solver private key is not provided")
)

// SubmitIntentSolutionParams is the parameters for submitting an intent solution
type SubmitIntentSolutionParams struct {

	// SolverPrivateKey is the private key of the solver
	// Required if SolverAddress, Hash, and Signature are not provided
	SolverPrivateKey string

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

	if params.IntentID == "" {
		return nil, ErrIntentIDRequired
	}

	if len(params.IntentSolution) == 0 {
		return nil, ErrIntentSolutionRequired
	}

	if params.SolverPrivateKey != "" {
		solverPrivateKey, err := crypto.HexToECDSA(params.SolverPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse solver private key: %v", err)
		}
		params.SolverAddress = crypto.PubkeyToAddress(solverPrivateKey.PublicKey).Hex()
		params.Hash = crypto.Keccak256Hash(params.IntentSolution).Bytes()
		params.Signature, err = crypto.Sign(params.Hash, solverPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to sign intentSolutionHash: %v", err)
		}
	} else if params.SolverAddress == "" || len(params.Hash) == 0 || len(params.Signature) == 0 {
		return nil, ErrSubmitIntentSolutionParamsRequired
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

	return c.handler.Request(ctx, jsonrpc.RPCSubmitIntentSolution, req)
}
