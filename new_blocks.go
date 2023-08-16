package bloxroute_sdk_go

import (
	"context"

	"github.com/bloXroute-Labs/gateway/v2/types"
)

// NewBlockParams is the params object for the OnNewBlock subscription
type NewBlockParams struct {
	// Include is the list of fields to include in the response.
	// The values of these fields depend on the feed type.
	// Optional (defaults to ["hash", "header"])
	Include []string `json:"include"`
}

// OnNewBlock subscribes to a stream of all new blocks as they are propagated in the BDN.
func (c *Client) OnNewBlock(ctx context.Context, params *NewBlockParams, callbackFunc CallbackFunc[*OnBdnBlockNotification]) error {
	if params == nil {
		params = &NewBlockParams{}
	}

	if len(params.Include) == 0 {
		params.Include = []string{"hash", "header"}
	}

	wrap := func(ctx context.Context, err error, result any) {
		callbackFunc(ctx, err, result.(*OnBdnBlockNotification))
	}

	return c.handler.Subscribe(ctx, types.NewBlocksFeed, params, wrap)
}

func (c *Client) UnsubscribeFromOnNewBlock() error {
	return c.handler.UnsubscribeRetry(types.NewBlocksFeed)
}
