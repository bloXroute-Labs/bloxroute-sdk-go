package bloxroute_sdk_go

import (
	"context"
	"encoding/json"

	"github.com/bloXroute-Labs/gateway/v2/jsonrpc"
)

// SendEthBundleParams is the parameters for sending an MEV bundle
type SendEthBundleParams struct {

	// The hex-encoded bytes of the transactions (without 0x prefix)
	Transactions []string `json:"transaction"`

	// Block number of a future block to include this bundle in, in hex value.
	// For traders who would like more than one block to be targeted,
	// please send multiple requests targeting each specific block.
	BlockNumber string `json:"block_number"`

	// [Optional] The minimum timestamp that the bundle is valid on,
	// an integer in unix epoch format. Default value is None.
	MinTimestamp uint `json:"min_timestamp,omitempty"`

	// [Optional] The minimum timestamp that the bundle is valid on,
	// an integer in unix epoch format. Default value is None.
	MaxTimestamp uint `json:"max_timestamp,omitempty"`

	// [Optional] A list of transaction hashes within the bundle that
	// are allowed to revert. Default is empty list: the whole bundle
	// would be excluded if any transaction reverts.
	RevertingHashes []string `json:"reverting_hashes,omitempty"`

	// [Optional] A unique identifier of the bundle. Used for replacement & cancellation
	// Visit https://docs.bloxroute.com/apis/mev-solution/bundle-submission for info
	Uuid string `json:"uuid,omitempty"`

	// [Optional, default: True] A boolean flag indicating if the MEV
	// bundle executes frontrunning strategy.
	Frontrunning bool `json:"frontrunning,omitempty"`

	// [Optional, default: False] A boolean flag indicating if the bundle
	// should be enrolled in the BackRunMe service.
	BackRunMe bool `json:"enable_backrunme,omitempty"`

	// [Optional] When BackRunMe is enabled, the transaction's from address
	// collects backrun reward by default but can be overwritten with this parameter.
	BackRunMeRewardAddress string `json:"backrunme_reward_address,omitempty"`

	// [Optional, default: bloxroute builder and flashbots builder]
	// A dictionary of MEV builders that should receive the bundle.
	MevBuilders map[string]string `json:"mev_builders,omitempty"`
}

// SendEthBundle submits a bundle to the Cloud-API or Gateway, which validates and forwards the bundle to MEV relays.
// Please contact bloXroute support if you have questions regarding the parameters.
func (c *Client) SendEthBundle(ctx context.Context, params *SendEthBundleParams) (*json.RawMessage, error) {
	return c.handler.Request(ctx, jsonrpc.RPCBundleSubmission, params)
}
