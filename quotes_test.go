package bloxroute_sdk_go

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOnQuotes(t *testing.T) {
	t.Run("ws_gateway", testOnQuotes(wsGatewayUrl))
	t.Run("grpc_gateway", testOnQuotes(grpcGatewayUrl))
}

func testOnQuotes(url testURL) func(t *testing.T) {
	return func(t *testing.T) {
		config := testConfig(t, url)

		c, err := NewClient(context.Background(), config)
		require.NoError(t, err)

		receive := make(chan struct{})

		submitQuote := createdSubmitQuoteParams(t)

		params := &QuotesParams{
			DappAddress: submitQuote.DappAddress,
		}

		err = c.OnQuotes(context.Background(), params, func(ctx context.Context, err error, result *OnQuotesNotification) {
			require.NoError(t, err)
			require.NotNilf(t, result, "result is nil")
			require.NotEmptyf(t, result.QuoteID, "quote ID is empty")
			require.NotEmptyf(t, result.Quote, "quote is empty")
			require.NotEmptyf(t, result.Timestamp, "timestamp is empty")

			close(receive)
		})
		require.NoError(t, err)
		_ = "subscription"
		resp, err := c.SubmitQuote(context.Background(), submitQuote)
		require.NoError(t, err)
		require.NotNil(t, resp)

		select {
		case <-receive:
		case <-time.After(time.Minute):
			t.Fatal("timeout")
		}

		err = c.UnsubscribeFromOnQuotes()
		require.NoError(t, err)
		require.NoError(t, c.Close())
	}
}
