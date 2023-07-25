package bloxroute_sdk_go

import (
	"context"

	"github.com/bloXroute-Labs/gateway/v2/types"
)

// OnNewBlockParams is the params object for the OnNewBlock subscription
type BdnBlockParams struct {
	// Include is the list of fields to include in the response.
	// The values of these fields depend on the feed type.
	// Optional (defaults to ["hash"])
	Include []string `json:"include"`
}

// OnNewBlock subscribes to a stream of all new blocks as they are propagated in the BDN.
func (c *Client) OnBdnBlock(ctx context.Context, params *BdnBlockParams, callbackFunc CallbackFunc[*OnBdnBlockNotification]) error {
	if params == nil {
		params = &BdnBlockParams{}
	}

	if len(params.Include) == 0 {
		params.Include = []string{"hash"}
	}

	wrap := func(ctx context.Context, err error, result any) {
		callbackFunc(ctx, err, result.(*OnBdnBlockNotification))
	}

	return c.handler.Subscribe(ctx, types.BDNBlocksFeed, params, wrap)
}

func (c *Client) UnsubscribeFromBdnBlock() error {
	return c.handler.UnsubscribeRetry(types.BDNBlocksFeed)
}
