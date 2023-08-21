package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bloXroute-Labs/gateway/v2/jsonrpc"
)

// SendPrivateTxParams are the parameters for sending private transactions with bloxroute
type SendPrivateTxParams struct {
	// The hex-encoded bytes of the transaction (without 0x prefix)
	Transaction string `json:"transaction"`

	// Optional, number of seconds to wait for the transaction to be included in a block
	// before it is sent to the network publically. Default is 0, which means it will
	// never be sent to the network publically.
	Timeout uint `json:"timeout,omitempty"`

	// A boolean flag indicating if the MEV bundle executes frontrunning strategy
	Frontrunning bool `json:"frontrunning,omitempty"`

	// An optional dictionary of MEV builders that should receive the private
	// transaction. The defaults are bloxroute and flashbots builders
	MevBuilders map[string]string `json:"mev_builders,omitempty"`
}

// SendPrivateTx sends a single transaction faster than the p2p network using the BDN
func (c *Client) SendPrivateTx(ctx context.Context, params *SendPrivateTxParams) (*json.RawMessage, error) {
	if c.blockchainNetwork != "Mainnet" {
		return nil, fmt.Errorf("SendPrivateTx is only supported on Mainnet")
	}

	// error if the user isn't using the cloud API
	if c.handler.Type() != handlerSourceTypeCloudAPIWS {
		return nil, fmt.Errorf("SendPrivateTx is only supported on the cloud API")
	}

	return c.handler.Request(ctx, jsonrpc.RPCPrivateTx, params)
}
