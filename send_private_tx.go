package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
	"fmt"

	bxgateway "github.com/bloXroute-Labs/gateway/v2"
	"github.com/bloXroute-Labs/gateway/v2/jsonrpc"
)

// SendPrivateTxParams are the parameters for sending private transactions with bloxroute.
type SendPrivateTxParams struct {
	// The hex-encoded bytes of the transaction (without 0x prefix)
	Transaction string `json:"transaction"`

	// Available on ETH Mainnet only. Optional, number of seconds to wait for the
	// transaction to be included in a block before it is sent to the network publically.
	// Default is 0, which means it will never be sent to the network publically.
	Timeout uint `json:"timeout,omitempty"`

	// Available on ETH Mainnet only. A boolean flag indicating if the MEV bundle executes
	// a frontrunning strategy
	Frontrunning bool `json:"frontrunning,omitempty"`

	// Available on ETH Mainnet only. An optional dictionary of MEV builders that should
	// receive the private transaction. The defaults are bloxroute and flashbots builders
	MevBuilders map[string]string `json:"mev_builders,omitempty"`
}

// SendPrivateTx sends private TX.
// When on ETH Mainnet, transactions are sent directly to block builders. When on BSC or
// Polygon, SendPrivateTx provides server side front-running protection based on the
// accessibility of the next validator and are eventually sent as semi-private
// transactions (https://docs.bloxroute.com/apis/frontrunning-protection/bsc_private_tx).
func (c *Client) SendPrivateTx(ctx context.Context, params *SendPrivateTxParams) (*json.RawMessage, error) {
	// error if the user isn't using the cloud API
	if c.handler.Type() != handlerSourceTypeCloudAPIWS {
		return nil, fmt.Errorf("SendPrivateTx is only supported on the cloud API")
	}

	if params == nil {
		return nil, ErrNilParams
	}

	requestType := jsonrpc.RPCPrivateTx

	if c.blockchainNetwork != bxgateway.Mainnet {
		// if any other params are set, error
		if params.MevBuilders != nil || params.Frontrunning || params.Timeout != 0 {
			return nil, fmt.Errorf("only the 'Transaction' field is supported for %s", c.blockchainNetwork)
		}
		if c.blockchainNetwork == "BSC-Mainnet" {
			requestType = RPCBSCPrivateTx
		} else if c.blockchainNetwork == "Polygon-Mainnet" {
			requestType = RPCPolygonPrivateTx
		}
	}

	return c.handler.Request(ctx, requestType, params)
}
