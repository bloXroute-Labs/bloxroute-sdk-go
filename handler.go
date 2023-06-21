package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bloXroute-Labs/gateway/v2/jsonrpc"
	"github.com/bloXroute-Labs/gateway/v2/types"
	"github.com/cenkalti/backoff/v4"
	"github.com/fasthttp/websocket"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/valyala/fastjson"
)

const (
	requestWaitTimeout = time.Minute

	unsubscribeTimeout         = time.Second * 10
	unsubscribeInitialInterval = time.Millisecond * 100

	reconnectTimeout         = time.Second * 30
	reconnectInitialInterval = time.Millisecond * 100
)

var (
	ErrNotConnected = errors.New("WS connection not established")
	ErrNoResponse   = errors.New("no response")
)

type handler struct {
	config          *Config
	conn            *websocket.Conn
	feeds           map[types.FeedType]struct{}
	subscriptions   map[string]subscription
	pendingResponse map[jsonrpc2.ID]chan requestResponse
	lock            *sync.Mutex
	stop            chan struct{}
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
type subscription struct {
	callback CallbackFunc
	feedType types.FeedType
	subReq   *jsonrpc2.Request
}

func (h *handler) connect(ctx context.Context) (err error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	var url string
	if h.config.CloudAPIURL != "" {
		url = h.config.CloudAPIURL
	} else {
		url = h.config.GatewayURL
	}

	h.conn, _, err = h.config.Dialer.DialContext(ctx, url, http.Header{authHeaderKey: []string{h.config.AuthHeader}})

	return
}

func (h *handler) reconnect(ctx context.Context) (err error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.conn, err = h.config.ReconnectFunc(ctx, h.config.Dialer, h.config.CloudAPIURL, h.config.AuthHeader)

	return
}

func (h *handler) read(ctx context.Context) error {
	for {
		select {
		case <-h.stop:
			return nil
		case <-ctx.Done():
			return h.close()
		default:
			messageType, message, err := h.conn.ReadMessage()
			if err != nil {
				select {
				case <-h.stop:
					// client is closed, so just return
					return nil
				default:
				}

				if !isWSClosedError(err) || !*h.config.Reconnect {
					return fmt.Errorf("failed to read message from cloud API: %w", err)
				}

				// connection is closed, try to reconnect
				err := h.reconnect(ctx)
				if err != nil {
					return err
				}

				// try to re-subscribe to all feeds
				go h.resubscribeAll(ctx)

				continue
			}
			if messageType != websocket.TextMessage {
				// TODO: handle non-text messages
				continue
			}

			err = h.handleMessage(ctx, message)
			if err != nil {
				h.config.Logger.Errorf("failed to handle message: %s", err)
			}
		}
	}
}

func (h *handler) handleMessage(ctx context.Context, message []byte) error {
	var p fastjson.Parser
	v, err := p.ParseBytes(message)
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	// check the subscription ID exists
	id := jsonrpc2.ID{Str: string(v.GetStringBytes("id")), IsString: true}
	resChan, ok := h.pendingResponse[id]
	if ok {
		defer delete(h.pendingResponse, id)
		defer close(resChan)

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

	method := v.GetStringBytes("method")
	if string(method) != "subscribe" {
		return nil
	}

	subscription, ok := h.subscriptions[string(v.GetStringBytes("params", "subscription"))]
	if !ok {
		// subscription not found
		return nil
	}

	paramsResult := json.RawMessage(v.GetObject("params", "result").String())

	subscription.callback(ctx, &paramsResult)

	return nil
}

// subscribe subscribes to a feed
func (h *handler) subscribe(ctx context.Context, f types.FeedType, subReq *jsonrpc2.Request) (chan requestResponse, error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	// check if already subscribed
	_, ok := h.feeds[f]
	if ok {
		return nil, fmt.Errorf("already subscribed to %s", types.NewTxsFeed)
	}

	// connect if not connected
	if h.conn == nil {
		return nil, ErrNotConnected
	}

	resChan := make(chan requestResponse, 1)

	// add subscription
	h.feeds[f] = struct{}{}
	h.pendingResponse[subReq.ID] = resChan

	err := h.conn.WriteJSON(subReq)
	if err != nil {
		return nil, fmt.Errorf("failed to write subscribe request for %s feed: %w", f, err)
	}

	return resChan, nil
}

// waitSubscriptionResponse waits for a subscription response
func (h *handler) waitSubscriptionResponse(ctx context.Context, resChan chan requestResponse, feedType types.FeedType, subReq *jsonrpc2.Request, callback CallbackFunc) error {
	wait := time.NewTimer(requestWaitTimeout)
	select {
	case <-ctx.Done():
		return nil
	case res, ok := <-resChan:
		if !ok {
			return nil
		}
		if res.Error != nil {
			return res.Error
		}

		// add a subscription
		h.lock.Lock()
		defer h.lock.Unlock()
		h.subscriptions[res.ID] = subscription{subReq: subReq, callback: callback, feedType: feedType}

		return nil
	case <-wait.C:
		h.lock.Lock()
		defer h.lock.Unlock()

		delete(h.pendingResponse, subReq.ID)
		delete(h.feeds, feedType)

		return fmt.Errorf("didn't receive response for %s subscribtion request within %s", feedType, requestWaitTimeout)
	}
}

// makes a single rpc request and expects a single response
func (h *handler) request(ctx context.Context, req *jsonrpc2.Request) (chan requestResponse, error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	// connect if not connected
	if h.conn == nil {
		return nil, ErrNotConnected
	}

	resChan := make(chan requestResponse, 1)

	h.pendingResponse[req.ID] = resChan

	err := h.conn.WriteJSON(req)
	if err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	return resChan, nil
}

// wait requestResponse waits for a request response and returns the response or error
func (h *handler) waitRequestResponse(ctx context.Context, resChan chan requestResponse, req *jsonrpc2.Request) (*json.RawMessage, error) {
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

func (h *handler) resubscribeAll(ctx context.Context) {
	subCopy := make([]subscription, 0, len(h.subscriptions))
	for _, subscription := range h.subscriptions {
		subCopy = append(subCopy, subscription)
	}
	h.subscriptions = make(map[string]subscription)
	h.feeds = make(map[types.FeedType]struct{})

loop:
	for _, subscription := range subCopy {
		select {
		case <-ctx.Done():
			return
		default:
			// create new subscription ID
			subscription.subReq.ID = randomID()
			resChan, err := h.subscribe(ctx, subscription.feedType, subscription.subReq)
			if err != nil {
				h.config.Logger.Errorf("failed to resubscribe to %s feed: %s", subscription.feedType, err)
				continue loop
			}
			err = h.waitSubscriptionResponse(ctx, resChan, subscription.feedType, subscription.subReq, subscription.callback)
			if err != nil {
				h.config.Logger.Errorf("failed to resubscribe to %s feed: %s", subscription.feedType, err)
			}
		}
	}
}

// unsubscribeRetry unsubscribes from a feed with retries
func (h *handler) unsubscribeRetry(f types.FeedType) error {
	backOff := backoff.NewExponentialBackOff()
	backOff.MaxElapsedTime = unsubscribeTimeout
	backOff.InitialInterval = unsubscribeInitialInterval

	fn := func() error {
		return h.unsubscribe(f)
	}

	return backoff.Retry(fn, backOff)
}

func (h *handler) unsubscribe(f types.FeedType) error {
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
		if subscription.feedType != f {
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

	err = h.conn.WriteJSON(unsubscribeRequest)
	if err != nil {
		return fmt.Errorf("failed to write unsubscribe request for %s feed: %w", f, err)
	}

	delete(h.feeds, f)
	delete(h.subscriptions, subscriptionID)

	return nil
}

func (h *handler) close() (err error) {
	close(h.stop)

	// no connection, so just return
	if h.conn == nil {
		return
	}

	unsubscribeRequest := &jsonrpc2.Request{
		ID:     randomID(),
		Method: string(jsonrpc.RPCUnsubscribe),
	}

	err = errors.Join(err, h.conn.WriteJSON(unsubscribeRequest), h.conn.Close())

	return
}
