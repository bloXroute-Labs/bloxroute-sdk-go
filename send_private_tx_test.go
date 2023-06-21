package bloxroute_sdk_go

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSendPrivateTx(t *testing.T) {
	t.Run("cloud api", testSendPrivateTx(t, "CLOUD_API_URL"))
}

func testSendPrivateTx(t *testing.T, url string) func(t *testing.T) {
	return func(t *testing.T) {
		config := &Config{
			AuthHeader: os.Getenv("AUTH_HEADER"),
		}
		switch url {
		case "CLOUD_API_URL":
			config.CloudAPIURL = os.Getenv("CLOUD_API_URL")
		case "GATEWAY_URL":
			config.GatewayURL = os.Getenv("GATEWAY_URL")
		}

		// get tx bytes from os env and error if not found
		txBytes := os.Getenv("TX_BYTES")
		require.NotEmpty(t, txBytes)

		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		defer cancel()

		c, err := NewClient(config)
		require.NoError(t, err)

		wg := &sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer wg.Done()
			err := c.Run(ctx)
			require.NoError(t, err)
		}()

		sendTxParams := &SendPrivateTxParams{
			Transaction: txBytes,
		}

		_, err = c.SendPrivateTx(ctx, sendTxParams)
		fmt.Println(err)
		require.NoError(t, err)

		require.NoError(t, c.Close())
		wg.Wait()
	}
}
