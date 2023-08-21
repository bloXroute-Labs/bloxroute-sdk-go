package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
)

// GetBscBundlePrice gets the BSC bundle price that corresponds to your subscription tier.
// The response has keys '1', '2', and 'higher', corresponding to the number of transactions in the bundle.
func (c *Client) GetBscBundlePrice(ctx context.Context) (*json.RawMessage, error) {
	return c.handler.Request(ctx, "bsc_get_bundle_price", nil)
}
