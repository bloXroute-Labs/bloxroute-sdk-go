package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/bloXroute-Labs/gateway/v2/jsonrpc"
	pb "github.com/bloXroute-Labs/gateway/v2/protobuf"
	"github.com/bloXroute-Labs/gateway/v2/types"
	"github.com/sourcegraph/jsonrpc2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type handlerSourceType int

const (
	handlerSourceTypeCloudAPIWS handlerSourceType = iota
	handlerSourceTypeGatewayWS
	handlerSourceTypeCloudAPIGRPC // not yet supported
	handlerSourceTypeGatewayGRPC
)

var ErrClientNotInitialized = errors.New("client not initialized")

type handler interface {
	Type() handlerSourceType
	Subscribe(ctx context.Context, f types.FeedType, req any, callback CallbackFunc[any]) error
	Request(ctx context.Context, method jsonrpc.RPCRequestType, params any) (*json.RawMessage, error)
	UnsubscribeRetry(f types.FeedType) error
	Close() error
}

// Client is a client for the bloXroute cloud API
type Client struct {
	handler           handler
	blockchainNetwork string
	initialized       bool
}

// NewClient creates a new SDK client
// Note: the client is not connected to the cloud API until Connect() is called
// or a subscription is made
func NewClient(ctx context.Context, config *Config) (*Client, error) {
	err := config.validate()
	if err != nil {
		return nil, err
	}

	config.setDefaults()

	c := &Client{
		blockchainNetwork: config.BlockchainNetwork,
	}

	err = c.connect(ctx, config)
	if err != nil {
		return nil, err
	}

	c.initialized = true

	return c, nil
}

// Close closes the connection to the cloud API
func (c *Client) Close() (err error) {
	if !c.initialized {
		return ErrClientNotInitialized
	}

	return c.handler.Close()
}

func (c *Client) connect(ctx context.Context, config *Config) error {
	if config.GRPCGatewayURL != "" {
		grpcConn, err := grpc.DialContext(ctx, strings.TrimPrefix(config.GRPCGatewayURL, "grpc://"), config.GRPCDialOptions...)
		if err != nil {
			return fmt.Errorf("failed to create GRPC connection: %w", err)
		}

		c.handler = &grpcHandler{
			hst:    handlerSourceTypeGatewayGRPC,
			config: config,
			conn:   grpcConn,
			client: pb.NewGatewayClient(grpcConn),
			md: metadata.New(map[string]string{
				blockchainHeaderKey: config.BlockchainNetwork,
				sdkVersionHeaderKey: buildVersion,
				languageHeaderKey:   runtime.Version(),
			}),
			stop:          make(chan struct{}),
			wg:            &sync.WaitGroup{},
			subscriptions: make(map[types.FeedType]grpcSubscription),
			lock:          &sync.Mutex{},
		}

		return nil
	}

	var hst handlerSourceType
	if config.WSCloudAPIURL != "" {
		hst = handlerSourceTypeCloudAPIWS
	} else {
		hst = handlerSourceTypeGatewayWS
	}

	h := &wsHandler{
		hst:             hst,
		config:          config,
		feeds:           make(map[types.FeedType]feed),
		subscriptions:   make(map[string]wsSubscription),
		pendingResponse: make(map[jsonrpc2.ID]chan requestResponse),
		lock:            &sync.Mutex{},
		stop:            make(chan struct{}),
		wg:              &sync.WaitGroup{},
		readErr:         make(chan error, 1),
	}

	err := h.reconnect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to WS: %w", err)
	}

	h.wg.Add(1)

	go h.read(ctx)

	c.handler = h

	return nil
}
