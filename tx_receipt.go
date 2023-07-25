package bloxroute_sdk_go

import (
	"context"

	"github.com/bloXroute-Labs/gateway/v2/types"
)

// TODO: add inline warning about gateway needing consensus layer, executing client, & ws connection

// TxReceiptParams is the parameters for subscribing to tx receipts
type TxReceiptParams struct {
	// Include is the list of fields to include in the response.
	// The values of these fields depend on the feed type.
	// Optional (defaults to ['block_hash'])
	Include []string `json:"include"`
}

// OnTxReceipt subscribes to  all transaction receipts in each newly mined block.
func (c *Client) OnTxReceipt(ctx context.Context, params *TxReceiptParams, callbackFunc CallbackFunc[*OnTxReceiptNotification]) error {
	if params == nil {
		params = &TxReceiptParams{}
	}

	// add at least block_hash to the include list
	if len(params.Include) == 0 {
		params.Include = []string{"block_hash"}
	}

	wrap := func(ctx context.Context, err error, result any) {
		if err != nil {
			callbackFunc(ctx, err, nil)
			return
		}
		callbackFunc(ctx, err, result.(*OnTxReceiptNotification))
	}

	return c.handler.Subscribe(ctx, types.TxReceiptsFeed, params, wrap)
}

// UnsubscribeFromTxReceipts unsubscribes from the tx receipts feed.
func (c *Client) UnsubscribeFromTxReceipts() error {
	return c.handler.UnsubscribeRetry(types.TxReceiptsFeed)
}
