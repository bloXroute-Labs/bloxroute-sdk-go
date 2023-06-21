package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClient_OnBlock(t *testing.T) {
	// t.Run("eth_call", testOnBlock(t, "eth_call"))
	t.Run("eth_getBalance", testOnBlock(t, "eth_getBalance"))
	t.Run("eth_getTransactionCount", testOnBlock(t, "eth_getTransactionCount"))
	t.Run("eth_getStorageAt", testOnBlock(t, "eth_getStorageAt"))
	t.Run("eth_blockNumber", testOnBlock(t, "eth_blockNumber"))

}

func testOnBlock(t *testing.T, method string) func(t *testing.T) {

	testAddress := "0xCbe321c620071307Ba5d0381c886B7359763735E"

	return func(t *testing.T) {
		config := &Config{
			AuthHeader: os.Getenv("AUTH_HEADER"),
		}
		config.GatewayURL = os.Getenv("GATEWAY_URL")

		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
		defer cancel()

		c, err := NewClient(config)
		require.NoError(t, err)

		started := make(chan struct{})
		wg := &sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer wg.Done()
			close(started)
			err := c.Run(ctx)
			require.NoError(t, err)
		}()

		<-started

		receive := make(chan struct{})

		params := &OnBlockParams{
			Include:    []string{"name", "response", "block_height", "tag"},
			CallParams: []OnBlockParamsCallParams{},
		}

		switch method {
		case "eth_call":
			params.CallParams = append(params.CallParams, &OnBlockParamsEthCall{
				OnBlockParamsCallParamsCommon: OnBlockParamsCallParamsCommon{
					Name:   "eth_call",
					Method: "eth_call",
					Tag:    "latest",
				},
			})
		case "eth_getBalance":
			params.CallParams = append(params.CallParams, &OnBlockParamsGetBalance{
				OnBlockParamsCallParamsCommon: OnBlockParamsCallParamsCommon{
					Name:   "eth_getBalance",
					Method: "eth_getBalance",
					Tag:    "latest",
				},
				Address: testAddress,
			})
		case "eth_getTransactionCount":
			params.CallParams = append(params.CallParams, &OnBlockParamsGetTransactionCount{
				OnBlockParamsCallParamsCommon: OnBlockParamsCallParamsCommon{
					Name:   "eth_getTransactionCount",
					Method: "eth_getTransactionCount",
					Tag:    "latest",
				},
				Address: testAddress,
			})
		case "eth_getCode":
			params.CallParams = append(params.CallParams, &OnBlockParamsGetCode{
				OnBlockParamsCallParamsCommon: OnBlockParamsCallParamsCommon{
					Name:   "eth_getCode",
					Method: "eth_getCode",
					Tag:    "latest",
				},
				Address: testAddress,
			})
		case "eth_getStorageAt":
			params.CallParams = append(params.CallParams, &OnBlockParamsGetStorageAt{
				OnBlockParamsCallParamsCommon: OnBlockParamsCallParamsCommon{
					Name:   "eth_getStorageAt",
					Method: "eth_getStorageAt",
					Tag:    "latest",
				},
				Address: testAddress,
			})
		case "eth_blockNumber":
			params.CallParams = append(params.CallParams, &OnBlockParamsBlockNumber{
				OnBlockParamsCallParamsCommon: OnBlockParamsCallParamsCommon{
					Name:   "eth_blockNumber",
					Method: "eth_blockNumber",
					Tag:    "latest",
				},
			})
		}

		done := false

		err = c.OnBlock(ctx, params, func(ctx context.Context, result *json.RawMessage) {
			var response struct {
				Name        string `json:"name"`
				Response    string `json:"response"`
				BlockHeight string `json:"block_height"`
				Tag         string `json:"tag"`
			}

			if !done {
				// needs some improvement
				if method == "eth_getBalance" {
					err := json.Unmarshal(*result, &response)
					require.NoError(t, err)
					require.Equal(t, "0x23708c8bd551c12", response.Response)
				} else if method == "eth_getTransactionCount" {
					err := json.Unmarshal(*result, &response)
					require.NoError(t, err)
					require.Equal(t, "0x96", response.Response)
				} else if method == "eth_getStorageAt" {
					err := json.Unmarshal(*result, &response)
					require.NoError(t, err)
					require.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000000000", response.Response)
				} else if method == "eth_blockNumber" {
					err := json.Unmarshal(*result, &response)
					require.NoError(t, err)
					_, err = strconv.ParseUint(response.Response, 0, 64)
					require.NoError(t, err)
				}
				done = true
				close(receive)
			}
		})

		require.NoError(t, err)

		select {
		case <-receive:
		case <-time.After(15 * time.Second):
			require.Failf(t, "timeout waiting for the %s event", method)
		}

		err = c.UnsubscribeFromEthOnBlock()
		require.NoError(t, err)

		require.NoError(t, c.Close())
		wg.Wait()
	}
}
