package bloxroute_sdk_go

import (
	"context"
	"encoding/json"

	"github.com/bloXroute-Labs/gateway/v2/jsonrpc"
)

// SendBscBundleParams is the parameters for sending a bundle of transactions
type SendBscBundleParams struct {

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

	// [Optional, default: all]
	// A dictionary of MEV builders that should receive the bundle. For each MEV builder, a signature (which can be an empty string) is required.
	MevBuilders map[string]string `json:"mev_builders,omitempty"`

	// [Optional, default: False]
	// A boolean indicating whether it is okay to mix the bundle with other bundles and transactions
	AvoidMixedBundles bool `json:"avoid_mixed_bundles,omitempty"`
}

type sendBscBundleParams struct {
	SendBscBundleParams
	BlockchainNetwork string `json:"blockchain_network"`
}

// SendBscBundle submits a BSC bundle to the Cloud-API, which validates and forwards the bundle to
// MEV Relays directly connected to BSC validators participating in our MEV solution program.
func (c *Client) SendBscBundle(ctx context.Context, params *SendBscBundleParams) (*json.RawMessage, error) {
	sendBscBundleParams := &sendBscBundleParams{
		SendBscBundleParams: SendBscBundleParams{
			Transactions:      params.Transactions,
			BlockNumber:       params.BlockNumber,
			MinTimestamp:      params.MinTimestamp,
			MaxTimestamp:      params.MaxTimestamp,
			RevertingHashes:   params.RevertingHashes,
			MevBuilders:       params.MevBuilders,
			AvoidMixedBundles: params.AvoidMixedBundles,
		},
		BlockchainNetwork: "BSC-Mainnet",
	}

	return c.handler.Request(ctx, jsonrpc.RPCBundleSubmission, sendBscBundleParams)
}
