package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/valyala/fastjson"

	"github.com/bloXroute-Labs/gateway/v2/jsonrpc"
	"github.com/bloXroute-Labs/gateway/v2/types"

	"github.com/bloXroute-Labs/bloxroute-sdk-go/connection/ws"
)

const (
	blockchainHeaderKey = "X-BloXroute-Blockchain"
	sdkVersionHeaderKey = "X-BloXroute-SDK-Version"
	languageHeaderKey   = "X-BloXroute-Code-Language"
	authHeaderKey       = "Authorization"

	requestWaitTimeout = time.Minute

	unsubscribeTimeout         = time.Second * 10
	unsubscribeInitialInterval = time.Millisecond * 100

	reconnectTimeout         = time.Second * 60
	reconnectInitialInterval = time.Millisecond * 100

	closeTimeout = time.Millisecond * 200
)

var (
	ErrNotConnected = errors.New("WS connection not established")
	ErrNoResponse   = errors.New("no response")
)

type wsHandler struct {
	hst             handlerSourceType
	config          *Config
	conn            ws.Conn
	feeds           map[types.FeedType]feed
	subscriptions   map[string]wsSubscription
	pendingResponse map[jsonrpc2.ID]chan requestResponse
	lock            *sync.Mutex
	stop            chan struct{}
	wg              *sync.WaitGroup
	readErr         chan error
}

// requestResponse represents a response to either a normal request or
// a subscription request. If it is a request, Result will be set, otherwise
// if it is a subscription request, ID will be the subscription ID
type requestResponse struct {
	ID     string
	Result []byte
	Error  *RPCError
}

// subscription represents a subscription to a feed
type wsSubscription struct {
	callback CallbackFunc[any]
	feed     types.FeedType
	subReq   *jsonrpc2.Request
}

type feed struct {
	subscriptionID string
}

// Type returns the WS handler type
func (h *wsHandler) Type() handlerSourceType {
	return h.hst
}

// Subscribe subscribes to a feed
func (h *wsHandler) Subscribe(ctx context.Context, f types.FeedType, params any, callback CallbackFunc[any]) error {
	raw, err := json.Marshal([]interface{}{f, params})
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}
	req := &jsonrpc2.Request{
		ID:     randomID(),
		Method: string(jsonrpc.RPCSubscribe),
		Params: (*json.RawMessage)(&raw),
	}

	resChan, err := h.subscribe(ctx, f, req)
	if err != nil {
		return err
	}

	_, err = h.waitSubscriptionResponse(ctx, resChan, f, req, callback)
	return err
}

// Request sends a request via WS
func (h *wsHandler) Request(ctx context.Context, method jsonrpc.RPCRequestType, params any) (*json.RawMessage, error) {
	raw, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}

	req := &jsonrpc2.Request{
		ID:     randomID(),
		Method: string(method),
		Params: (*json.RawMessage)(&raw),
	}

	resChan, err := h.request(ctx, req)
	if err != nil {
		return nil, err
	}

	return h.waitRequestResponse(ctx, resChan, req)
}

// Close stops the read loop, unsubscribes from all feeds and closes the connection
func (h *wsHandler) Close() error {
	close(h.stop)

	// unsubscribe from all feeds and close the connection
	unsubscribeRequest := &jsonrpc2.Request{
		ID:     randomID(),
		Method: string(jsonrpc.RPCUnsubscribe),
	}

	ctx, cancel := context.WithTimeout(context.Background(), unsubscribeTimeout)
	defer cancel()

	err := errors.Join(h.conn.WriteJSON(ctx, unsubscribeRequest), h.conn.Close())

	// this is workaround for the fact that the Read blocks until there is a message
	// and some WS implementations don't support ctx cancellation
	c := make(chan struct{})
	go func() {
		defer close(c)
		h.wg.Wait()
	}()
	select {
	case <-c:
	case <-time.After(closeTimeout):
	}

	return err
}

// UnsubscribeRetry unsubscribes from a feed with retries
func (h *wsHandler) UnsubscribeRetry(f types.FeedType) error {
	backOff := backoff.NewExponentialBackOff()
	backOff.MaxElapsedTime = unsubscribeTimeout
	backOff.InitialInterval = unsubscribeInitialInterval

	fn := func() error {
		return h.unsubscribe(f)
	}

	return backoff.Retry(fn, backOff)
}

// read starts reading messages via WS from gateway or cloud API and handles them.
// read blocks until the context is canceled, an error occurs or the client is closed.
func (h *wsHandler) read(ctx context.Context) {
	defer h.wg.Done()

	for {
		select {
		case <-h.stop:
			return
		case <-ctx.Done():
			return
		default:
			message, err := h.conn.ReadMessage(ctx)
			if err != nil {
				select {
				case <-h.stop:
					// client is closed, so just return
					return
				default:
				}

				if !ws.IsWSClosedError(err) || !*h.config.Reconnect {
					h.config.Logger.Errorf("failed to read message from WS: %s", err)
					return
				}

				// connection is closed, try to reconnect
				err := h.reconnect(ctx)
				if err != nil {
					h.config.Logger.Errorf("failed to reconnect to WS: %s", err)
					continue
				}

				// try to re-subscribe to all feeds
				go h.resubscribeAll(ctx)

				continue
			}

			// double-check if message is nil
			if message == nil {
				continue
			}

			err = h.handleMessage(ctx, message)
			if err != nil {
				h.config.Logger.Errorf("failed to handle message: %s", err)
			}
		}
	}
}

func (h *wsHandler) reconnect(ctx context.Context) error {
	var urlString string
	if h.config.WSCloudAPIURL != "" {
		urlString = h.config.WSCloudAPIURL
	} else {
		urlString = h.config.WSGatewayURL
	}

	headers := http.Header{
		authHeaderKey:       []string{h.config.AuthHeader},
		blockchainHeaderKey: []string{h.config.BlockchainNetwork},
		sdkVersionHeaderKey: []string{buildVersion},
		languageHeaderKey:   []string{runtime.Version()},
	}

	conn, err := h.config.WSConnectFunc(ctx, urlString, headers, h.config.WSDialOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to WS: %w", err)
	}

	h.conn = conn

	return nil
}

func (h *wsHandler) handleMessage(ctx context.Context, message []byte) error {
	var p fastjson.Parser
	v, err := p.ParseBytes(message)
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	// check the subscription ID exists
	id := jsonrpc2.ID{Str: string(v.GetStringBytes("id")), IsString: true}

	h.lock.Lock()
	resChan, ok := h.pendingResponse[id]
	h.lock.Unlock()
	if ok {
		return h.handlePendingResponse(id, resChan, v)
	}

	method := v.GetStringBytes("method")
	if string(method) != "subscribe" {
		return nil
	}

	h.lock.Lock()
	subscription, ok := h.subscriptions[string(v.GetStringBytes("params", "subscription"))]
	h.lock.Unlock()
	if !ok {
		// subscription not found
		return nil
	}

	var res any

	switch subscription.feed {
	case types.OnBlockFeed:
		res = &OnBlockNotification{
			Name:        string(v.GetStringBytes("params", "result", "name")),
			Response:    string(v.GetStringBytes("params", "result", "response")),
			BlockHeight: string(v.GetStringBytes("params", "result", "block_height")),
			Tag:         string(v.GetStringBytes("params", "result", "tag")),
		}
	case types.BDNBlocksFeed:
		res = &OnBdnBlockNotification{}
		err = json.Unmarshal(v.GetObject("params", "result").MarshalTo(nil), &res)
		if err != nil {
			err = fmt.Errorf("failed to unmarshal bdn block notification: %w", err)
		}
	case types.NewBlocksFeed:
		res = &OnBdnBlockNotification{}
		err = json.Unmarshal(v.GetObject("params", "result").MarshalTo(nil), &res)
		if err != nil {
			err = fmt.Errorf("failed to unmarshal new block notification: %w", err)
		}
	case types.NewTxsFeed, types.PendingTxsFeed:
		res = &NewTxNotification{}
		err = json.Unmarshal(v.GetObject("params", "result").MarshalTo(nil), &res)
		if err != nil {
			err = fmt.Errorf("failed to unmarshal new tx notification: %w", err)
		}
	case types.TransactionStatusFeed:
		res = &OnTxStatusNotification{
			TxHash: string(v.GetStringBytes("params", "result", "tx_hash")),
			Status: string(v.GetStringBytes("params", "result", "status")),
		}
	case types.TxReceiptsFeed:
		res = &OnTxReceiptNotification{}
		err = json.Unmarshal(v.GetObject("params", "result").MarshalTo(nil), &res)
		if err != nil {
			err = fmt.Errorf("failed to unmarshal tx receipt notification: %w", err)
		}
	case types.UserIntentsFeed:
		res = &OnIntentsNotification{
			DappAddress:   string(v.GetStringBytes("params", "result", "dapp_address")),
			SenderAddress: string(v.GetStringBytes("params", "result", "sender_address")),
			IntentID:      string(v.GetStringBytes("params", "result", "intent_id")),
			Intent:        v.GetStringBytes("params", "result", "intent"),
			Timestamp:     string(v.GetStringBytes("params", "result", "timestamp")),
		}
	case types.UserIntentSolutionsFeed:
		res = &OnIntentSolutionsNotification{
			IntentID:       string(v.GetStringBytes("params", "result", "intent_id")),
			IntentSolution: v.GetStringBytes("params", "result", "intent_solution"),
			SolutionID:     string(v.GetStringBytes("params", "result", "solution_id")),
		}
	case types.QuotesFeed:
		res = &OnQuotesNotification{
			DappAddress:   string(v.GetStringBytes("params", "result", "dapp_address")),
			QuoteID:       string(v.GetStringBytes("params", "result", "quote_id")),
			SolverAddress: string(v.GetStringBytes("params", "result", "solver_address")),
			Quote:         v.GetStringBytes("params", "result", "quote"),
			Timestamp:     string(v.GetStringBytes("params", "result", "timestamp")),
		}
	}

	subscription.callback(ctx, err, res)

	return nil
}

func (h *wsHandler) handlePendingResponse(id jsonrpc2.ID, resChan chan requestResponse, v *fastjson.Value) error {
	defer func() {
		close(resChan)
		h.lock.Lock()
		delete(h.pendingResponse, id)
		h.lock.Unlock()
	}()

	errMessage := v.GetObject("error")
	if errMessage != nil {
		code, _ := errMessage.Get("code").Int64()
		data := json.RawMessage(errMessage.Get("data").GetStringBytes())
		select {
		case resChan <- requestResponse{Error: &RPCError{Code: code, Message: errMessage.Get("message").String(), Data: &data}}:
			return nil
		default:
			return nil
		}
	}

	// when result is a string, it is a subscription ID for a subscription request
	// otherwise it looks to be a response to a normal request
	result := v.Get("result")
	if result != nil {
		subscriptionId := v.GetStringBytes("result")
		if len(subscriptionId) > 0 {
			select {
			case resChan <- requestResponse{ID: string(subscriptionId)}:
				return nil
			default:
				return nil
			}
		} else {
			resBytes := result.MarshalTo(nil)
			select {
			case resChan <- requestResponse{Result: resBytes}:
				return nil
			default:
			}
		}
	}

	return nil
}

// subscribe subscribes to a feed
func (h *wsHandler) subscribe(ctx context.Context, f types.FeedType, subReq *jsonrpc2.Request) (chan requestResponse, error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	// check if already subscribed
	_, ok := h.feeds[f]
	if ok {
		return nil, fmt.Errorf("already subscribed to %s", f)
	}

	// connect if not connected
	if h.conn == nil {
		return nil, ErrNotConnected
	}

	resChan := make(chan requestResponse, 1)

	// add subscription
	h.feeds[f] = feed{subscriptionID: ""}
	h.pendingResponse[subReq.ID] = resChan

	err := h.conn.WriteJSON(ctx, subReq)
	if err != nil {
		return nil, fmt.Errorf("failed to write subscribe request for %s feed: %w", f, err)
	}

	return resChan, nil
}

// waitSubscriptionResponse waits for a subscription response
func (h *wsHandler) waitSubscriptionResponse(ctx context.Context, resChan chan requestResponse, feedType types.FeedType, subReq *jsonrpc2.Request, callback CallbackFunc[any]) (string, error) {
	wait := time.NewTimer(requestWaitTimeout)
	select {
	case <-ctx.Done():
		return "", nil
	case res, ok := <-resChan:
		if !ok {
			return "", nil
		}
		if res.Error != nil {
			return "", res.Error
		}

		// add a subscription
		h.lock.Lock()
		defer h.lock.Unlock()
		h.subscriptions[res.ID] = wsSubscription{subReq: subReq, callback: callback, feed: feedType}
		h.feeds[feedType] = feed{subscriptionID: res.ID}

		return res.ID, nil
	case <-wait.C:
		h.lock.Lock()
		defer h.lock.Unlock()

		delete(h.pendingResponse, subReq.ID)
		delete(h.feeds, feedType)

		return "", fmt.Errorf("didn't receive response for %s subscription request within %s", feedType, requestWaitTimeout)
	}
}

// makes a single rpc request and expects a single response
func (h *wsHandler) request(ctx context.Context, req *jsonrpc2.Request) (chan requestResponse, error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	// connect if not connected
	if h.conn == nil {
		return nil, ErrNotConnected
	}

	resChan := make(chan requestResponse, 1)

	h.pendingResponse[req.ID] = resChan

	err := h.conn.WriteJSON(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	return resChan, nil
}

// wait requestResponse waits for a request response and returns the response or error
func (h *wsHandler) waitRequestResponse(ctx context.Context, resChan chan requestResponse, req *jsonrpc2.Request) (*json.RawMessage, error) {
	wait := time.NewTimer(requestWaitTimeout)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res, ok := <-resChan:
		if !ok {
			return nil, ErrNoResponse
		}

		if res.Error != nil {
			return nil, res.Error
		}

		result := json.RawMessage(res.Result)
		return &result, nil
	case <-wait.C:
		h.lock.Lock()
		defer h.lock.Unlock()

		delete(h.pendingResponse, req.ID)

		return nil, fmt.Errorf("request timed out")
	}
}

func (h *wsHandler) resubscribeAll(ctx context.Context) {
	h.lock.Lock()

	subCopy := make([]wsSubscription, 0, len(h.subscriptions))
	for _, subscription := range h.subscriptions {
		subCopy = append(subCopy, subscription)
	}
	h.subscriptions = make(map[string]wsSubscription)
	h.feeds = make(map[types.FeedType]feed)

	h.lock.Unlock()

loop:
	for _, subscription := range subCopy {
		select {
		case <-ctx.Done():
			return
		default:
			// create new subscription ID
			subscription.subReq.ID = randomID()
			resChan, err := h.subscribe(ctx, subscription.feed, subscription.subReq)
			if err != nil {
				h.config.Logger.Errorf("failed to resubscribe to %s feed: %s", subscription.feed, err)
				continue loop
			}
			_, err = h.waitSubscriptionResponse(ctx, resChan, subscription.feed, subscription.subReq, subscription.callback)
			if err != nil {
				h.config.Logger.Errorf("failed to resubscribe to %s feed: %s", subscription.feed, err)
			}
		}
	}
}

func (h *wsHandler) unsubscribe(f types.FeedType) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	if h.conn == nil {
		return backoff.Permanent(fmt.Errorf("connection is not established"))
	}

	_, ok := h.feeds[f]
	if !ok {
		// no need to unsubscribe
		return nil
	}

	var subscriptionID string
	for id, subscription := range h.subscriptions {
		if subscription.feed != f {
			continue
		}
		subscriptionID = id
		break
	}

	if subscriptionID == "" {
		return fmt.Errorf("no subscription ID is defined for %s yet", types.NewTxsFeed)
	}

	raw, err := json.Marshal([]interface{}{subscriptionID})
	if err != nil {
		return backoff.Permanent(fmt.Errorf("failed to marshal params: %w", err))
	}

	unsubscribeRequest := &jsonrpc2.Request{
		ID:     randomID(),
		Method: string(jsonrpc.RPCUnsubscribe),
		Params: (*json.RawMessage)(&raw),
	}

	ctx, cancel := context.WithTimeout(context.Background(), unsubscribeTimeout)
	defer cancel()

	err = h.conn.WriteJSON(ctx, unsubscribeRequest)
	if err != nil {
		return fmt.Errorf("failed to write unsubscribe request for %s feed: %w", f, err)
	}

	delete(h.feeds, f)
	delete(h.subscriptions, subscriptionID)

	return nil
}
