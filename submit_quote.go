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
	ErrSubmitQuoteGatewayOnly    = errors.New("SubmitQuote is only supported on the gateway GRPC and WS handlers")
	ErrQuoteRequired             = errors.New("quote is required")
	ErrSubmitQuoteParamsRequired = fmt.Errorf("sender address, hash, and signature are required when sender private key is not provided")
)

// SubmitQuoteParams is the parameters for submitting a Quote
type SubmitQuoteParams struct {
	// DappAddress is the ETH address of the DApp that should receive this Quote
	DappAddress string

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

	// Quote is the Quote payload
	Quote []byte
}

// SubmitQuote submits a Quote to the BDN
func (c *Client) SubmitQuote(ctx context.Context, params *SubmitQuoteParams) (*json.RawMessage, error) {
	if c.handler.Type() != handlerSourceTypeGatewayGRPC && c.handler.Type() != handlerSourceTypeGatewayWS {
		return nil, ErrSubmitQuoteGatewayOnly
	}

	if params == nil {
		return nil, ErrNilParams
	}

	if params.Quote == nil {
		return nil, ErrQuoteRequired
	}

	if params.SolverPrivateKey != "" {
		solverPrivateKey, err := crypto.HexToECDSA(params.SolverPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse solver private key: %v", err)
		}
		params.SolverAddress = crypto.PubkeyToAddress(solverPrivateKey.PublicKey).Hex()
		params.Hash = crypto.Keccak256Hash(params.Quote).Bytes()
		params.Signature, err = crypto.Sign(params.Hash, solverPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to sign intentSolutionHash: %v", err)
		}
	} else if params.SolverAddress == "" || len(params.Hash) == 0 || len(params.Signature) == 0 {
		return nil, ErrSubmitQuoteParamsRequired
	}

	var req interface{}

	if c.handler.Type() == handlerSourceTypeGatewayGRPC {
		req = &pb.SubmitQuoteRequest{
			DappAddress:   params.DappAddress,
			SolverAddress: params.SolverAddress,
			Quote:         params.Quote,
			Hash:          params.Hash,
			Signature:     params.Signature,
		}
	} else {
		req = &jsonrpc.RPCSubmitQuotePayload{
			DappAddress:   params.DappAddress,
			SolverAddress: params.SolverAddress,
			Quote:         params.Quote,
			Hash:          params.Hash,
			Signature:     params.Signature,
		}
	}

	return c.handler.Request(ctx, jsonrpc.RPCSubmitQuote, req)
}
