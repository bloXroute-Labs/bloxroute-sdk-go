package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bloXroute-Labs/gateway/v2/jsonrpc"
	"github.com/bloXroute-Labs/gateway/v2/types"
	"github.com/sourcegraph/jsonrpc2"
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
func (c *Client) OnBlock(ctx context.Context, params *OnBlockParams, callbackFunc CallbackFunc) error {
	if params == nil {
		return fmt.Errorf("params is nil or empty")
	}
	if len(params.CallParams) == 0 {
		return fmt.Errorf("at least one call_params is required")
	}

	raw, err := json.Marshal([]interface{}{types.OnBlockFeed, params})
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}

	subRequest := &jsonrpc2.Request{
		ID:     randomID(),
		Method: string(jsonrpc.RPCSubscribe),
		Params: (*json.RawMessage)(&raw),
	}

	return c.subscribe(ctx, types.OnBlockFeed, subRequest, callbackFunc)
}

func (c *Client) UnsubscribeFromEthOnBlock() error {
	return c.unsubscribeRetry(types.OnBlockFeed)
}
