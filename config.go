package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"

	"github.com/fasthttp/websocket"
)

const (
	authHeaderKey = "Authorization"
)

var (
	ErrorNilConfig           = errors.New("config is nil")
	ErrEndpointNotProvided   = errors.New("either cloud API or gateway URL must be provided")
	ErrAuthHeaderNotProvided = errors.New("either auth header or account ID and secret must be provided")
	ErrInvalidCloudApiURL    = errors.New("invalid cloud API URL")
)

type CallbackFunc func(ctx context.Context, result *json.RawMessage)

type ReconnectFunc func(ctx context.Context, dialer *websocket.Dialer, url, authHeader string) (*websocket.Conn, error)

// Config is the configuration for the SDK
// It contains necessary information for creating SDK client
type Config struct {
	// CloudAPIURL is the URL of the cloud API
	// Required if GatewayURL is not provided
	CloudAPIURL string

	// GatewayURL is your gateway's websocket URL
	// Required if CloudAPIURL is not provided
	GatewayURL string

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

	// Dialer is the websocket dialer
	// Optional
	Dialer *websocket.Dialer

	// Reconnect is a flag that indicates whether the SDK should reconnect to the cloud API in case of disconnection
	// Optional (default: true)
	Reconnect *bool

	// ReconnectFunc is a function that is called when the SDK needs to reconnect to the cloud API
	// Optional (default: exponential backoff with 30s timeout)
	ReconnectFunc ReconnectFunc

	// Logger is the logger used by the SDK
	// Optional (default: no logging)
	Logger logger
}

type logger interface {
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
		return ErrorNilConfig
	}

	if c.CloudAPIURL == "" && c.GatewayURL == "" {
		return ErrEndpointNotProvided
	}

	if c.CloudAPIURL != "" {
		if _, err := url.ParseRequestURI(c.CloudAPIURL); err != nil {
			return ErrInvalidCloudApiURL
		}
	}

	if c.AuthHeader == "" && (c.AccountID == "" || c.Secret == "") {
		return ErrAuthHeaderNotProvided
	}
	return nil
}
