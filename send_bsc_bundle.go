package bloxroute_sdk_go

import (
	"context"
	"encoding/json"

	"github.com/bloXroute-Labs/gateway/v2/jsonrpc"
)

// SendBscBundleParams is the parameters for sending a bundle of transactions
type SendBscBundleParams struct {
	// [Optional, default: all]
	// A dictionary of MEV builders that should receive the bundle. For each MEV builder, a signature (which can be an empty string) is required.
	MevBuilders map[string]string `json:"mev_builders,omitempty"`

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

	// [Optional] A unique identifier of the bundle.
	UUID string `json:"uuid"`

	// [Optional, default: False]
	// A boolean indicating whether it is okay to mix the bundle with other bundles and transactions
	AvoidMixedBundles bool `json:"avoid_mixed_bundles,omitempty"`

	OriginalSenderAccountID string `json:"original_sender_account_id"`

	// Optional, default: False] A boolean flag indicating if the bundle should be sent just to a
	// single block builder who is participating in the priority fee refund program.
	PriorityFeeRefund bool `json:"priority_fee_refund"`

	// [Optional] A string representing the wallet address to receive refund when priority_fee_refund
	// flag is enabled. Users who do not want to specify the refund_recipient parameter must contact
	// bloXroute to enable their refund address.
	IncomingRefundRecipient string `json:"refund_recipient"`

	// [Optional, default: 1] An integer that specifies the number of subsequent blocks that the bundle is valid for.
	// The maximum value allowed for this parameter is 20. For example, when block_number parameter is 1000,
	// and blocks_count is 3, then the current bundle would be processed with block numbers 1000, 1001, 1002.
	BlocksCount int `json:"blocks_count,omitempty"`

	// [Optional] A list of transaction hashes within the bundle that can be removed from the bundle if it's
	// deemed useful (but not revert). For example, when transaction is invalid. Default is empty list:
	// the whole bundle would be excluded if any transaction fails.
	DroppingTxHashes []string `json:"dropping_tx_hashes,omitempty"`

	// From protocol version 53
	EndOfBlock bool `json:"end_of_block"`
}

type sendBscBundleParams struct {
	SendBscBundleParams
	BlockchainNetwork string `json:"blockchain_network"`
}

// SendBscBundle submits a BSC bundle to the Cloud-API, which validates and forwards the bundle to
// MEV Relays directly connected to BSC validators participating in our MEV solution program.
func (c *Client) SendBscBundle(ctx context.Context, params *SendBscBundleParams) (*json.RawMessage, error) {
	sendBscBundleParams := &sendBscBundleParams{
		SendBscBundleParams: *params,
		BlockchainNetwork:   c.blockchainNetwork,
	}

	return c.handler.Request(ctx, jsonrpc.RPCBundleSubmission, sendBscBundleParams)
}
