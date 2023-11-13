package bloxroute_sdk_go

import (
	"context"

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
	// Optional (defaults to ["tx_hash"] for WS and ["raw_tx"] for GRPC)
	Include []string `json:"include"`

	// Duplicates indicates whether to include transactions already published in the feed
	// Optional (defaults to false)
	Duplicates bool `json:"duplicates"`

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
func (c *Client) OnNewTx(ctx context.Context, params *NewTxParams, callbackFunc CallbackFunc[*NewTxNotification]) error {
	if params == nil {
		params = &NewTxParams{}
	}

	// add at least tx_hash to the include list
	if len(params.Include) == 0 {
		if c.handler.Type() != handlerSourceTypeGatewayGRPC {
			params.Include = []string{"tx_hash"}
		} else {
			params.Include = []string{"raw_tx"}
		}
	}

	wrap := func(ctx context.Context, err error, result any) {
		if err != nil {
			callbackFunc(ctx, err, nil)
			return
		}
		callbackFunc(ctx, err, result.(*NewTxNotification))
	}

	return c.handler.Subscribe(ctx, types.NewTxsFeed, params, wrap)
}

// UnsubscribeFromNewTxs unsubscribes from new transactions feed
func (c *Client) UnsubscribeFromNewTxs() error {
	return c.handler.UnsubscribeRetry(types.NewTxsFeed)
}
