package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/bloXroute-Labs/gateway/v2/jsonrpc"
)

// SendTxParams are the parameters for sending transactions faster with
// bloxroute & configuring semi-private transactions
type SendTxParams struct {

	// The hex-encoded bytes of the transaction (without 0x prefix)
	Transaction string `json:"transaction"`

	// BlockchainNetwork is the blockchain network to subscribe to.
	// Optional (defaults to "Mainnet")
	BlockchainNetwork string `json:"blockchain_network,omitempty"`

	// A boolean flag indicating if Tx Nonce Monitoring should be
	// enabled for the transaction.
	NonceMonitoring bool `json:"nonce_monitoring,omitempty"`

	// A boolean flag used to send the transaction only to validators
	// accessible via the BDN. Available only for Ethereum, BSC & Polygon
	ValidatorsOnly bool `json:"validators_only,omitempty"`

	// A boolean flag used to send the transaction only to the validators
	// that are next-in-turn, if they are accessible via the BDN. Available
	// only for BSC & Polygon
	NextValidator bool `json:"next_validator,omitempty"`

	// When using next_validator, fall_back is the duration of time (in ms)
	// that your transaction will be delayed before propagation by the
	// BDN as a normal transaction. Default is 0, which indicates no fallback.
	Fallback uint `json:"fallback,omitempty"`

	// A boolean flag indicating if the transaction should be validated by the
	// connected blockchain node via the Gateway. Default is false.
	NodeValidation bool `json:"node_validation,omitempty"`
}

// SendTx sends a single transaction faster than the p2p network using the BDN
func (c *Client) SendTx(ctx context.Context, params *SendTxParams) (*json.RawMessage, error) {

	// set blockchain network to match the config if not set
	if params.BlockchainNetwork == "" {
		if c.cloudAPIHandler != nil {
			params.BlockchainNetwork = c.cloudAPIHandler.config.BlockchainNetwork
		} else {
			params.BlockchainNetwork = c.gatewayHandler.config.BlockchainNetwork
		}
	}

	// error if the user is using mainnet and next validator
	if params.BlockchainNetwork == "Mainnet" && params.NextValidator {
		return nil, fmt.Errorf("NextValidator is not supported on Ethereum Mainnet")
	}

	raw, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}

	subRequest := &jsonrpc2.Request{
		ID:     randomID(),
		Method: string(jsonrpc.RPCTx),
		Params: (*json.RawMessage)(&raw),
	}

	return c.request(ctx, subRequest)

}
