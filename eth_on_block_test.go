package bloxroute_sdk_go

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClient_OnBlock(t *testing.T) {
	// t.ReadWS("eth_call", testOnBlock(t, "eth_call"))
	t.Run("ws_gateway_eth_getBalance", testOnBlock("eth_getBalance"))
	t.Run("ws_gateway_eth_getTransactionCount", testOnBlock("eth_getTransactionCount"))
	t.Run("ws_gateway_eth_getStorageAt", testOnBlock("eth_getStorageAt"))
	t.Run("ws_gateway_eth_blockNumber", testOnBlock("eth_blockNumber"))
}

func testOnBlock(method string) func(t *testing.T) {
	testAddress := "0xCbe321c620071307Ba5d0381c886B7359763735E"

	return func(t *testing.T) {
		config := testConfig(t, wsGatewayUrl)

		c, err := NewClient(context.Background(), config)
		require.NoError(t, err)

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

		err = c.OnBlock(context.Background(), params, func(ctx context.Context, err error, response *OnBlockNotification) {
			select {
			case <-receive:
				return
			default:
			}

			require.NoError(t, err)

			switch method {
			case "eth_getBalance":
				require.Equal(t, "0x23708c8bd551c12", response.Response)
			case "eth_getTransactionCount":
				require.Equal(t, "0x96", response.Response)
			case "eth_getStorageAt":
				require.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000000000", response.Response)
			case "eth_blockNumber":
				_, err = strconv.ParseUint(response.Response, 0, 64)
				require.NoError(t, err)
			default:
				t.FailNow()
			}

			close(receive)
		})

		require.NoError(t, err)

		select {
		case <-receive:
		case <-time.After(time.Minute):
			require.Fail(t, fmt.Sprintf("timeout waiting for the %s event", method))
		}

		err = c.UnsubscribeFromEthOnBlock()
		require.NoError(t, err)

		require.NoError(t, c.Close())
	}
}
