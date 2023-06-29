package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/bloXroute-Labs/gateway/v2/jsonrpc"
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
func (c *Client) OnTxReceipt(ctx context.Context, params *TxReceiptParams, cb CallbackFunc) error {
	if params == nil {
		params = &TxReceiptParams{}
	}

	// add at least block_hash to the include list
	if len(params.Include) == 0 {
		params.Include = []string{"block_hash"}
	}

	raw, err := json.Marshal([]interface{}{types.TxReceiptsFeed, params})
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
func (c *Client) UnsubscribeFromTxReceipts() error {
	return c.unsubscribeRetry(types.TxReceiptsFeed)
}
