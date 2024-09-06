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
	ErrSubmitIntentGatewayOnly    = errors.New("SubmitIntent is only supported on the gateway GRPC and WS handlers")
	ErrIntentRequired             = errors.New("intent is required")
	ErrDappAddressRequired        = errors.New("dApp address is required")
	ErrSubmitIntentParamsRequired = fmt.Errorf("sender address, hash, and signature are required when sender private key is not provided")
)

// SubmitIntentParams is the parameters for submitting an intent
type SubmitIntentParams struct {
	// DappPrivateKey is the private key of your DApp used to prove the ownership of the DApp address
	// Required if DappAddress is not provided
	DappPrivateKey string

	// DappAddress is the ETH address of the DApp that should receive solution for this intent
	DappAddress string

	// SenderPrivateKey is the private key of the intent sender
	// Required if SenderAddress, Hash, and Signature are not provided
	SenderPrivateKey string

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

	if params.Intent == nil {
		return nil, ErrIntentRequired
	}

	if params.DappPrivateKey != "" {
		privateKey, err := crypto.HexToECDSA(params.DappPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse dapp private key: %v", err)
		}
		dappAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
		params.DappAddress = dappAddress
	} else if params.DappAddress == "" {
		return nil, ErrDappAddressRequired
	}

	if params.SenderPrivateKey != "" {
		privateKey, err := crypto.HexToECDSA(params.SenderPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse sender private key: %v", err)
		}
		params.SenderAddress = crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
		params.Hash = crypto.Keccak256Hash(params.Intent).Bytes()
		params.Signature, err = crypto.Sign(params.Hash, privateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to sign sender hash: %v", err)
		}
	} else if params.SenderAddress == "" || len(params.Hash) == 0 || len(params.Signature) == 0 {
		return nil, ErrSubmitIntentParamsRequired
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

	return c.handler.Request(ctx, jsonrpc.RPCSubmitIntent, req)
}
