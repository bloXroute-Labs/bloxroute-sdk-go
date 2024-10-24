package bloxroute_sdk_go

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	bxgateway "github.com/bloXroute-Labs/gateway/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bloXroute-Labs/bloxroute-sdk-go/connection/ws"
)

var (
	ErrNilConfig             = errors.New("config is nil")
	ErrEndpointNotProvided   = errors.New("either cloud API or gateway URL must be provided")
	ErrAuthHeaderNotProvided = errors.New("either auth header or account ID and secret must be provided")
)

// CallbackFunc is the function used to handle the result of a subscription
type CallbackFunc[T any] func(ctx context.Context, err error, result T)

// WSConnectFunc is the function used to connect to the WS endpoint
type WSConnectFunc func(ctx context.Context, url string, headers http.Header, dialOpts *ws.DialOptions) (ws.Conn, error)

// Config is the configuration for the SDK
// It contains necessary information for creating SDK client
type Config struct {
	// WSCloudAPIURL is the URL of the cloud API
	// Required if either WSGatewayURL or GRPCGatewayURL is not provided
	WSCloudAPIURL string

	// WSGatewayURL is your gateway's URL
	// Required if either WSCloudAPIURL or GRPCGatewayURL is not provided
	WSGatewayURL string

	// GRPCGatewayURL is your gateway's URL
	// Required if either WSCloudAPIURL or WSGatewayURL is not provided
	GRPCGatewayURL string

	// AuthHeader is the authorization header for the cloud and gateway APIs
	// In case AccountID and Secret are provided, this field be set automatically
	// Optional (if AccountID and Secret are provided)
	AuthHeader string

	// Account ID received when registering the account
	// Optional (if AuthHeader is provided)
	AccountID string

	// Secret hash received when registering the account
	// Optional (if AuthHeader is provided)
	Secret string

	// BlockchainNetwork
	// Optional (default: "Mainnet")
	BlockchainNetwork string

	// WSDialOptions is the websocket dialer options
	// Optional
	WSDialOptions *ws.DialOptions

	// WSConnectFunc is a function that is called when the SDK creates a connection or needs to reconnect to the endpoint
	// Optional (default: exponential backoff with 30s timeout)
	WSConnectFunc WSConnectFunc

	// GRPCDialOptions is the grpc dialer options
	GRPCDialOptions []grpc.DialOption

	// GRPCDialTimeout is the grpc dialer timeout
	GRPCDialTimeout time.Duration

	// Reconnect is a flag that indicates whether the SDK should reconnect to the cloud API in case of disconnection
	// Optional (default: true)
	Reconnect *bool

	// Logger is the Logger used by the SDK
	// Optional (default: no logging)
	Logger Logger
}

type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

func (c *Config) validate() error {
	if c == nil {
		return ErrNilConfig
	}

	if c.WSCloudAPIURL == "" && c.WSGatewayURL == "" && c.GRPCGatewayURL == "" {
		return ErrEndpointNotProvided
	}

	if c.AuthHeader == "" && (c.AccountID == "" || c.Secret == "") {
		return ErrAuthHeaderNotProvided
	}

	return nil
}

func (c *Config) setDefaults() {
	if c.AuthHeader == "" {
		c.AuthHeader = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.AccountID, c.Secret)))
	}

	if c.Reconnect == nil {
		reconnect := true
		c.Reconnect = &reconnect
	}

	if c.WSConnectFunc == nil {
		c.WSConnectFunc = reconnect
	}

	if c.Logger == nil {
		c.Logger = &NoopLogger{}
	}

	if c.BlockchainNetwork == "" {
		c.BlockchainNetwork = bxgateway.Mainnet
	}

	if len(c.GRPCDialOptions) == 0 {
		c.GRPCDialOptions = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	}
	c.GRPCDialOptions = append(c.GRPCDialOptions, grpc.WithPerRPCCredentials(grpcCredentials{authorization: c.AuthHeader}))
}

type grpcCredentials struct {
	authorization string
}

func (bc grpcCredentials) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": bc.authorization,
	}, nil
}

func (bc grpcCredentials) RequireTransportSecurity() bool {
	return false
}
