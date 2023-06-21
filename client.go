package bloxroute_sdk_go

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/bloXroute-Labs/gateway/v2/types"
	"github.com/fasthttp/websocket"
	"github.com/sourcegraph/jsonrpc2"
)

var (
	ErrClientNotInitialized = errors.New("client not initialized")
)

// Client is a client for the bloXroute cloud API
type Client struct {
	cloudAPIHandler *handler
	gatewayHandler  *handler
	initialized     bool
}

// NewClient creates a new SDK client
// Note: the client is not connected to the cloud API until Connect() is called
// or a subscription is made
func NewClient(config *Config) (*Client, error) {
	err := config.validate()
	if err != nil {
		return nil, err
	}

	if config.AuthHeader == "" {
		config.AuthHeader = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", config.AccountID, config.Secret)))
	}

	if config.Reconnect == nil {
		reconnect := true
		config.Reconnect = &reconnect
	}

	if config.Reconnect != nil && *config.Reconnect && config.ReconnectFunc == nil {
		config.ReconnectFunc = reconnect
	}

	if config.Logger == nil {
		config.Logger = &noopLogger{}
	}

	if config.BlockchainNetwork == "" {
		config.BlockchainNetwork = "Mainnet"
	}

	c := &Client{
		initialized: true,
	}

	if config.Dialer == nil {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}
		config.Dialer = websocket.DefaultDialer
		config.Dialer.TLSClientConfig = tlsConfig
	}

	// TODO check if we need to change the handler structure for gateway and cloudAPI
	h := &handler{
		config:          config,
		feeds:           make(map[types.FeedType]struct{}),
		subscriptions:   make(map[string]subscription),
		pendingResponse: make(map[jsonrpc2.ID]chan requestResponse),
		lock:            &sync.Mutex{},
		stop:            make(chan struct{}),
	}

	if config.CloudAPIURL != "" {
		c.cloudAPIHandler = h
	} else {
		c.gatewayHandler = h
	}

	return c, nil
}

// Close closes the connection to the cloud API
func (c *Client) Close() (err error) {
	if !c.initialized {
		return ErrClientNotInitialized
	}

	if c.cloudAPIHandler != nil {
		return c.cloudAPIHandler.close()
	}

	return c.gatewayHandler.close()
}

// Run reads messages from the cloud API and handles them
// Run blocks until the context is canceled or an error occurs
func (c *Client) Run(ctx context.Context) error {
	if !c.initialized {
		return ErrClientNotInitialized
	}

	if c.cloudAPIHandler != nil {
		err := c.cloudAPIHandler.connect(ctx)
		if err != nil {
			return fmt.Errorf("failed to connect to cloud API: %w", err)
		}

		return c.cloudAPIHandler.read(ctx)
	}

	err := c.gatewayHandler.connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to gateway: %w", err)
	}

	return c.gatewayHandler.read(ctx)
}

func (c *Client) subscribe(ctx context.Context, f types.FeedType, subReq *jsonrpc2.Request, callback CallbackFunc) error {
	if c.cloudAPIHandler != nil {
		resChan, err := c.cloudAPIHandler.subscribe(ctx, f, subReq)
		if err != nil {
			return err
		}

		return c.cloudAPIHandler.waitSubscriptionResponse(ctx, resChan, f, subReq, callback)
	}

	resChan, err := c.gatewayHandler.subscribe(ctx, f, subReq)
	if err != nil {
		return err
	}

	return c.gatewayHandler.waitSubscriptionResponse(ctx, resChan, f, subReq, callback)
}

// make a request to the cloud API and wait for the response
func (c *Client) request(ctx context.Context, req *jsonrpc2.Request) (*json.RawMessage, error) {
	if c.cloudAPIHandler != nil {
		resChan, err := c.cloudAPIHandler.request(ctx, req)
		if err != nil {
			return nil, err
		}
		return c.cloudAPIHandler.waitRequestResponse(ctx, resChan, req)
	}

	resChan, err := c.gatewayHandler.request(ctx, req)
	if err != nil {
		return nil, err
	}
	return c.gatewayHandler.waitRequestResponse(ctx, resChan, req)
}

func (c *Client) unsubscribeRetry(f types.FeedType) error {
	if c.cloudAPIHandler != nil {
		return c.cloudAPIHandler.unsubscribeRetry(f)
	}

	return c.gatewayHandler.unsubscribeRetry(f)
}
