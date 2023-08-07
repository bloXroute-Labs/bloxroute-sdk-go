package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/bloXroute-Labs/gateway/v2/jsonrpc"
	"github.com/bloXroute-Labs/gateway/v2/types"
)

var (
	ErrCloudAPIOnly = errors.New("OnTxStatus & MonitorTx are only supported on the cloud API")
	ErrNoSubID      = errors.New("failed to find subscription for transaction status feed")
)

// OnTxStatus subscribes to a stream of transaction statuses
func (c *Client) OnTxStatus(ctx context.Context, params OnTxStatusParams) error {

	if c.handler.Type() != handlerSourceTypeCloudAPIWS {
		return ErrCloudAPIOnly
	}

	// return an error if there is no callback
	if params.Callback == nil {
		return fmt.Errorf("callback is required")
	}

	wrap := func(ctx context.Context, err error, result any) {
		if err != nil {
			params.Callback(ctx, err, nil)
			return
		}
		params.Callback(ctx, err, result.(*OnTxStatusNotification))
	}

	_, err := subscribeTransactionStatus(ctx, c.handler, wrap)
	if err != nil {
		return err
	}

	if params.Transactions != nil {
		return c.MonitorTxs(ctx, &MonitorTxsParams{
			Transactions: params.Transactions,
		})
	}

	return nil
}

// MonitorTxs monitors the status of transactions
func (c *Client) MonitorTxs(ctx context.Context, params *MonitorTxsParams) error {
	if c.handler.Type() != handlerSourceTypeCloudAPIWS {
		return ErrCloudAPIOnly
	}

	handler := c.handler.(*wsHandler)

	feed, ok := handler.feeds[types.TransactionStatusFeed]
	if !ok {
		return fmt.Errorf("please subscribe to a transaction status feed with OnTxStatus before calling MonitorTxs")
	}

	subscriptionId := feed.subscriptionID

	monitorTxsParams := &monitorTxsParams{
		Transactions:   params.Transactions,
		SubscriptionID: subscriptionId,
	}

	raw, err := json.Marshal(monitorTxsParams)
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}

	subRequest := &jsonrpc2.Request{
		ID:     randomID(),
		Method: string(jsonrpc.RPCStartMonitoringTx),
		Params: (*json.RawMessage)(&raw),
	}

	_, err = handler.request(ctx, subRequest)
	if err != nil {
		return fmt.Errorf("failed to start monitoring transactions: %w", err)
	}

	return nil
}

func subscribeTransactionStatus(ctx context.Context, h handler, callback CallbackFunc[any]) (string, error) {
	if h.Type() != handlerSourceTypeCloudAPIWS {
		return "", ErrCloudAPIOnly
	}

	hh := h.(*wsHandler)

	raw, err := json.Marshal([]interface{}{types.TransactionStatusFeed, map[string]any{"include": []string{"tx_hash", "status"}}})

	if err != nil {
		return "", fmt.Errorf("failed to marshal params: %w", err)
	}

	req := &jsonrpc2.Request{
		ID:     randomID(),
		Method: string(jsonrpc.RPCSubscribe),
		Params: (*json.RawMessage)(&raw),
	}

	resChan, err := hh.subscribe(ctx, types.TransactionStatusFeed, req)
	if err != nil {
		return "", err
	}

	return hh.waitSubscriptionResponse(ctx, resChan, types.TransactionStatusFeed, req, callback)
}

// OnTxStatusParams allow you to include a parameters and a callback function
type OnTxStatusParams struct {
	// Callback function for status updates
	Callback CallbackFunc[*OnTxStatusNotification]

	// [Optional] Specify the transactions to monitor. If unspecified,
	// use MonitorTxs set & update the monitored transactions.
	Transactions []string `json:"transactions"`
}

// MonitorTxsParams are the parameters for updating the monitored transactions
type MonitorTxsParams struct {
	// Raw transaction bytes without 0x prefix.
	Transactions []string `json:"transactions"`
}

type monitorTxsParams struct {
	Transactions   []string `json:"transactions"`
	SubscriptionID string   `json:"subscription_id"`
}

// StopMonitoringTxParams are the parameters for monitoring transactions
type StopMonitoringTxParams struct {
	// Raw transaction bytes without 0x prefix.
	Transactions []string `json:"transactions"`
	// Raw transaction bytes without 0x prefix.
	TransactionHash []string `json:"transaction_hash"`
}

type stopMonitoringTxParams struct {
	Transactions    []string `json:"transactions"`
	TransactionHash []string `json:"transaction_hash"`
	SubscriptionID  string   `json:"subscription_id"`
}

// StopMonitoringTx stops monitoring the status of transactions specified
func (c *Client) StopMonitoringTx(ctx context.Context, params *StopMonitoringTxParams) error {
	if c.handler.Type() != handlerSourceTypeCloudAPIWS {
		return ErrCloudAPIOnly
	}

	handler := c.handler.(*wsHandler)

	feed, ok := handler.feeds[types.TransactionStatusFeed]
	if !ok {
		return ErrNoSubID
	}

	_, ok = handler.subscriptions[feed.subscriptionID]
	if !ok {
		return ErrNoSubID
	}

	// Included Transactions; seems the cloud API requires it despite the docs
	stopMonitorTxsParams := &stopMonitoringTxParams{
		Transactions:    params.Transactions,
		TransactionHash: params.TransactionHash,
		SubscriptionID:  feed.subscriptionID,
	}

	_, err := handler.Request(ctx, jsonrpc.RPCStopMonitoringTx, stopMonitorTxsParams)
	if err != nil {
		return fmt.Errorf("failed to stop monitoring transactions: %w", err)
	}

	err = handler.UnsubscribeRetry(types.TransactionStatusFeed)
	if err != nil {
		return fmt.Errorf("failed to unsubscribe from transaction status feed: %w", err)
	}

	return nil

}
