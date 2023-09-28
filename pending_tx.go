package bloxroute_sdk_go

import (
	"context"

	"github.com/bloXroute-Labs/gateway/v2/types"
)

// PendingTxParams is the parameters for subscribing to new transactions
type PendingTxParams struct {
	// Include is the list of fields to include in the response.
	// The values of these fields depend on the feed type.
	// Optional (defaults to ["tx_hash"])
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

// OnPendingTx subscribes to types.PendingTxsFeed feed
func (c *Client) OnPendingTx(ctx context.Context, params *PendingTxParams, callbackFunc CallbackFunc[*NewTxNotification]) error {
	if params == nil {
		params = &PendingTxParams{}
	}

	// add at least tx_hash to the include list
	if len(params.Include) == 0 {
		params.Include = []string{"tx_hash"}
	}

	wrap := func(ctx context.Context, err error, result any) {
		if err != nil {
			callbackFunc(ctx, err, nil)
			return
		}
		callbackFunc(ctx, err, result.(*NewTxNotification))
	}

	return c.handler.Subscribe(ctx, types.PendingTxsFeed, params, wrap)
}

// UnsubscribeFromPendingTxs unsubscribes from types.PendingTxsFeed feed
func (c *Client) UnsubscribeFromPendingTxs() error {
	return c.handler.UnsubscribeRetry(types.PendingTxsFeed)
}
