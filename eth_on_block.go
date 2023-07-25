package bloxroute_sdk_go

import (
	"context"
	"fmt"

	"github.com/bloXroute-Labs/gateway/v2/types"
)

// OnBlockParams is the params object for the eth_onBlock subscription
type OnBlockParams struct {
	// Include is a list of fields to include in the response
	// Optional
	Include []string `json:"include"`

	// CallParams is used to build an RPC call request
	// Required
	CallParams []OnBlockParamsCallParams `json:"call-params"`
}

// OnBlockParamsCallParams represents a value in the CallParams array
type OnBlockParamsCallParams interface {
	isEthOnBlockParamsCallParams()
}

// OnBlockParamsCallParamsCommon is the common fields for all CallParams
type OnBlockParamsCallParamsCommon struct {
	// Method is the RPC method to call
	Method string `json:"method"`

	//
	Tag string `json:"tag,omitempty"`

	// Name is a unique string identifier for call
	Name string `json:"name,omitempty"`
}

func (*OnBlockParamsCallParamsCommon) isEthOnBlockParamsCallParams() {}

// OnBlockParamsEthCall represents params for eth_call
// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_call
type OnBlockParamsEthCall struct {
	OnBlockParamsCallParamsCommon
	From  string `json:"from,omitempty"`
	To    string `json:"to"`
	Gas   string `json:"gas"`
	Value string `json:"value"`
	Data  string `json:"data"`
}

// OnBlockParamsGetBalance represents params for eth_getBalance
// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_getbalance
type OnBlockParamsGetBalance struct {
	OnBlockParamsCallParamsCommon
	Address string `json:"address"`
}

// OnBlockParamsGetTransactionCount represents params for eth_getTransactionCount
// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_gettransactioncount
type OnBlockParamsGetTransactionCount struct {
	OnBlockParamsCallParamsCommon
	Address string `json:"address"`
}

// OnBlockParamsGetCode represents params for eth_getCode
// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_getcode
type OnBlockParamsGetCode struct {
	OnBlockParamsCallParamsCommon
	Address string `json:"address"`
}

// OnBlockParamsGetStorageAt represents params for eth_getStorageAt
// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_getstorageat
type OnBlockParamsGetStorageAt struct {
	OnBlockParamsCallParamsCommon
	Address string `json:"address"`
	Pos     string `json:"pos"`
}

// OnBlockParamsBlockNumber represents params for eth_blockNumber
// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_blocknumber
type OnBlockParamsBlockNumber struct {
	OnBlockParamsCallParamsCommon
}

// OnBlock subscribes to stream of changes in the EVM state when a new block is mined
func (c *Client) OnBlock(ctx context.Context, params *OnBlockParams, callbackFunc CallbackFunc[*OnBlockNotification]) error {
	if params == nil {
		return fmt.Errorf("params is nil or empty")
	}
	if len(params.CallParams) == 0 {
		return fmt.Errorf("at least one call_params is required")
	}

	wrap := func(ctx context.Context, err error, result any) {
		if err != nil {
			callbackFunc(ctx, err, nil)
			return
		}
		callbackFunc(ctx, err, result.(*OnBlockNotification))
	}

	return c.handler.Subscribe(ctx, types.OnBlockFeed, params, wrap)
}

func (c *Client) UnsubscribeFromEthOnBlock() error {
	return c.handler.UnsubscribeRetry(types.OnBlockFeed)
}
