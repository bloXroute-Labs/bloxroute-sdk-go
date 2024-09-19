package bloxroute_sdk_go

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/bloXroute-Labs/gateway/v2/protobuf"
	"github.com/bloXroute-Labs/gateway/v2/types"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	ErrQuotesGatewayOnly               = errors.New("OnQuotes is only supported on the gateway GRPC and WS handlers")
	ErrDappAddressOrPrivateKeyRequired = errors.New("dApp address or dApp private key is required")
)

// QuotesParams is the params object for the OnQuotes subscription
type QuotesParams struct {
	// DappAddress is the ETH address of the DApp
	// Required if DappPrivateKey is not provided
	DappAddress string

	// DappPrivateKey is the private key of the DApp
	// Required if DappAddress is not provided
	DappPrivateKey string
}

// OnQuotesNotification is the notification object for the OnQuotes subscription
type OnQuotesNotification struct {
	QuoteID       string `json:"quote_id"`
	DappAddress   string `json:"dapp_address"`
	SolverAddress string `json:"solver_address"`
	Quote         []byte `json:"quote"`
	Timestamp     string `json:"timestamp"`
}

// OnQuotes subscribes to a stream of new quotes that match the dappAddress of the subscription as they are propagated in the BDN.
func (c *Client) OnQuotes(ctx context.Context, params *QuotesParams, callbackFunc CallbackFunc[*OnQuotesNotification]) error {
	if c.handler.Type() != handlerSourceTypeGatewayGRPC && c.handler.Type() != handlerSourceTypeGatewayWS {
		return ErrQuotesGatewayOnly
	}

	if params == nil {
		return ErrNilParams
	}

	if params.DappAddress == "" && params.DappPrivateKey == "" {
		return ErrDappAddressOrPrivateKeyRequired
	}

	if params.DappPrivateKey != "" {
		privateKey, err := crypto.HexToECDSA(params.DappPrivateKey)
		if err != nil {
			return fmt.Errorf("failed to parse dApp private key: %v", err)
		}
		dappAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
		params.DappAddress = dappAddress
	}

	wrap := func(ctx context.Context, err error, result any) {
		if err != nil {
			callbackFunc(ctx, err, nil)
			return
		}
		callbackFunc(ctx, nil, result.(*OnQuotesNotification))
	}

	var req interface{}

	if c.handler.Type() == handlerSourceTypeGatewayGRPC {
		req = &pb.QuotesRequest{
			DappAddress: params.DappAddress,
		}
	} else {
		req = map[string]interface{}{
			"dapp_address": params.DappAddress,
		}
	}

	return c.handler.Subscribe(ctx, types.QuotesFeed, req, wrap)
}

// UnsubscribeFromOnQuotes unsubscribes from the OnQuotes subscription
func (c *Client) UnsubscribeFromOnQuotes() error {
	return c.handler.UnsubscribeRetry(types.QuotesFeed)
}
