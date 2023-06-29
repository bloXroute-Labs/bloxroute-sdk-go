package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/bloXroute-Labs/gateway/v2/jsonrpc"
	"github.com/bloXroute-Labs/gateway/v2/types"
)

// Project is the name of a defi project for use with a tenderly feed
type Project string

// Project enumeration
const (
	UniswapV2   Project = "uniswapV2"
	UniswapV3   Project = "uniswapV3"
	Pancakeswap Project = "pancakeswap"
	Biswap      Project = "biswap"
	Sushiswap   Project = "sushiswap"
	Apeswap     Project = "apeswap"
	Quickswap   Project = "quickswap"
)

// NewTxParams is the parameters for subscribing to new transactions
type NewTxParams struct {
	// Include is the list of fields to include in the response.
	// The values of these fields depend on the feed type.
	// Optional (defaults to ["tx_hash"])
	Include []string `json:"include"`

	// Filters is SQL-like syntax string for logical operations.
	// Optional
	Filters string `json:"filters,omitempty"`

	// BlockchainNetwork is the blockchain network to subscribe to.
	// Optional (defaults to "Mainnet")
	BlockchainNetwork string `json:"blockchain_network,omitempty"`

	// Project is the name of a DeFi project for use with a tenderly feed
	// Optional
	Project Project `json:"project,omitempty"`
}

// OnNewTx subscribes to new transactions feed
func (c *Client) OnNewTx(ctx context.Context, params *NewTxParams, cb CallbackFunc) error {
	if params == nil {
		params = &NewTxParams{}
	}

	// add at least tx_hash to the include list
	if len(params.Include) == 0 {
		params.Include = []string{"tx_hash"}
	}

	raw, err := json.Marshal([]interface{}{types.NewTxsFeed, params})
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}

	subReq := &jsonrpc2.Request{
		ID:     randomID(),
		Method: string(jsonrpc.RPCSubscribe),
		Params: (*json.RawMessage)(&raw),
	}
	return c.subscribe(ctx, types.NewTxsFeed, subReq, cb)
}

// UnsubscribeFromNewTxs unsubscribes from new transactions feed
func (c *Client) UnsubscribeFromNewTxs() error {
	return c.unsubscribeRetry(types.NewTxsFeed)
}
