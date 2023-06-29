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

func (h *handler) connect(ctx context.Context) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	url := h.config.GatewayURL
	if h.config.CloudAPIURL != "" {
		url = h.config.CloudAPIURL
	}

	conn, _, err := h.config.Dialer.DialContext(ctx, url, http.Header{authHeaderKey: []string{h.config.AuthHeader}})
	if err != nil {
		return err
	}

	h.conn = conn
	return nil
}

func (h *handler) reconnect(ctx context.Context) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	conn, err := h.config.ReconnectFunc(ctx, h.config.Dialer, h.config.CloudAPIURL, h.config.AuthHeader)
	if err != nil {
		return err
	}

	h.conn = conn
	return nil
}

func (h *handler) read(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return h.close()
		case <-h.stop:
			return nil
		default:
			msgType, msg, err := h.conn.ReadMessage()
			if err != nil {
				select {
				case <-h.stop:
					// client is closed
					return nil
				default:
					// pass through
				}

				if !isWSClosedError(err) || !*h.config.Reconnect {
					return fmt.Errorf("failed to read message from cloud API: %w", err)
				}

				// connection is closed, try to reconnect
				if err := h.reconnect(ctx); err != nil {
					return err
				}

				// try to re-subscribe to all feeds
				go h.resubscribeAll(ctx)

				continue
			}

			// TODO: handle non-text messages
			if msgType != websocket.TextMessage {
				continue
			}

			if err := h.handleMessage(ctx, msg); err != nil {
				h.config.Logger.Errorf("failed to handle message: %s", err)
			}
		}
	}
}

func (h *handler) handleMessage(ctx context.Context, message []byte) error {
	parse, err := fastjson.ParseBytes(message)
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	// check the subscription ID exists
	id := jsonrpc2.ID{
		Str:      string(parse.GetStringBytes("id")),
		IsString: true,
	}

	resChan, ok := h.pendingResponse[id]
	if ok {
		defer delete(h.pendingResponse, id)
		defer close(resChan)

		if err := parse.GetObject("error"); err != nil {
			code, _ := err.Get("code").Int64()
			data := json.RawMessage(err.Get("data").GetStringBytes())
			params := requestResponse{Error: &RPCError{Code: code, Message: err.Get("message").String(), Data: &data}}
			select {
			case resChan <- params:
				return nil
			default:
				return nil
			}
		}

		// when result is a string, it is a subscription ID for a subscription request
		// otherwise it looks to be a response to a normal request
		if res := parse.Get("result"); res != nil {
			subscriptionId := parse.GetStringBytes("result")
			if len(subscriptionId) > 0 {
				params := requestResponse{ID: string(subscriptionId)}
				select {
				case resChan <- params:
					return nil
				default:
					return nil
				}
			} else {
				params := requestResponse{Result: res.MarshalTo(nil)}
				select {
				case resChan <- params:
					return nil
				default:
					// drop the response
				}
			}
		}
		return nil
	}

	method := string(parse.GetStringBytes("method"))
	if method != "subscribe" {
		return nil
	}

	sub, ok := h.subscriptions[string(parse.GetStringBytes("params", "subscription"))]
	if !ok {
		// subscription not found
		return nil
	}

	params := json.RawMessage(parse.GetObject("params", "result").String())
	sub.callback(ctx, &params)
	return nil
}

// subscribe subscribes to a feed
func (h *handler) subscribe(ctx context.Context, f types.FeedType, subReq *jsonrpc2.Request) (chan requestResponse, error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	// check if already subscribed
	if _, ok := h.feeds[f]; ok {
		err := fmt.Errorf("already subscribed to %s", types.NewTxsFeed)
		return nil, err
	}

	// connect if not connected
	if h.conn == nil {
		return nil, ErrNotConnected
	}

	// add subscription
	ch := make(chan requestResponse, 1)
	h.feeds[f] = struct{}{}
	h.pendingResponse[subReq.ID] = ch
	if err := h.conn.WriteJSON(subReq); err != nil {
		return nil, fmt.Errorf("failed to write subscribe request for %s feed: %w", f, err)
	}
	return ch, nil
}

// waitSubscriptionResponse waits for a subscription response
func (h *handler) waitSubscriptionResponse(ctx context.Context, resChan chan requestResponse, feedType types.FeedType, subReq *jsonrpc2.Request, cb CallbackFunc) error {
	timer := time.NewTimer(requestWaitTimeout)
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
		h.subscriptions[res.ID] = subscription{subReq: subReq, callback: cb, feedType: feedType}

		return nil
	case <-timer.C:
		h.lock.Lock()
		defer h.lock.Unlock()

		delete(h.pendingResponse, subReq.ID)
		delete(h.feeds, feedType)

		err := fmt.Errorf("didn't receive response for %s subscribtion request within %s", feedType, requestWaitTimeout)
		return err
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

	ch := make(chan requestResponse, 1)
	h.pendingResponse[req.ID] = ch

	if err := h.conn.WriteJSON(req); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}
	return ch, nil
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
		err := fmt.Errorf("request timed out")
		return nil, err
	}
}

func (h *handler) resubscribeAll(ctx context.Context) {
	subs := make([]subscription, 0, len(h.subscriptions))
	for _, sub := range h.subscriptions {
		subs = append(subs, sub)
	}
	h.subscriptions = make(map[string]subscription)
	h.feeds = make(map[types.FeedType]struct{})

loop:
	for _, sub := range subs {
		select {
		case <-ctx.Done():
			return
		default:
			// create new subscription ID
			sub.subReq.ID = randomID()
			ch, err := h.subscribe(ctx, sub.feedType, sub.subReq)
			if err != nil {
				h.config.Logger.Errorf("failed to resubscribe to %s feed: %s", sub.feedType, err)
				continue loop
			}
			if err := h.waitSubscriptionResponse(ctx, ch, sub.feedType, sub.subReq, sub.callback); err != nil {
				h.config.Logger.Errorf("failed to resubscribe to %s feed: %s", sub.feedType, err)
			}
		}
	}
}

// unsubscribeRetry unsubscribes from a feed with retries
func (h *handler) unsubscribeRetry(f types.FeedType) error {
	backOff := backoff.NewExponentialBackOff()
	backOff.MaxElapsedTime = unsubscribeTimeout
	backOff.InitialInterval = unsubscribeInitialInterval
	return backoff.Retry(func() error {
		return h.unsubscribe(f)
	}, backOff)
}

func (h *handler) unsubscribe(f types.FeedType) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	if h.conn == nil {
		return backoff.Permanent(fmt.Errorf("connection is not established"))
	}

	if _, ok := h.feeds[f]; !ok {
		// no need to unsubscribe
		return nil
	}

	var subID string
	for id, sub := range h.subscriptions {
		if sub.feedType != f {
			continue
		}
		subID = id
		break
	}

	if subID == "" {
		return fmt.Errorf("no subscription ID is defined for %s yet", types.NewTxsFeed)
	}

	raw, err := json.Marshal([]interface{}{subID})
	if err != nil {
		return backoff.Permanent(fmt.Errorf("failed to marshal params: %w", err))
	}

	// send unsubscribe request
	if err = h.conn.WriteJSON(&jsonrpc2.Request{
		ID:     randomID(),
		Method: string(jsonrpc.RPCUnsubscribe),
		Params: (*json.RawMessage)(&raw),
	}); err != nil {
		return fmt.Errorf("failed to write unsubscribe request for %s feed: %w", f, err)
	}

	delete(h.feeds, f)
	delete(h.subscriptions, subID)
	return nil
}

func (h *handler) close() error {
	close(h.stop)
	if h.conn != nil {
		var err error
		unsubReq := h.conn.WriteJSON(&jsonrpc2.Request{
			ID:     randomID(),
			Method: string(jsonrpc.RPCUnsubscribe),
		})
		err = errors.Join(err, unsubReq, h.conn.Close())
		return err
	}
	return nil
}
