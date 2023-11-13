package bloxroute_sdk_go

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonitorTxs(t *testing.T) {
	t.Run("ws_cloud_api", testMonitorTxs(wsCloudApiUrl))
}

func testMonitorTxs(url testURL) func(t *testing.T) {
	return func(t *testing.T) {
		config := testConfig(t, url)

		c, err := NewClient(context.Background(), config)
		require.NoError(t, err)

		var wg sync.WaitGroup
		txs := make(map[string]string)
		mutex := sync.Mutex{}
		wg.Add(1)
		newTxErrChan := make(chan error, 1)

		err = c.OnNewTx(
			context.Background(),
			&NewTxParams{Include: []string{"tx_hash", "raw_tx"}},
			func(ctx context.Context, err error, result *NewTxNotification) {
				require.NoError(t, err)
				mutex.Lock()
				defer mutex.Unlock()
				txs[result.TxHash] = result.RawTx
				if len(txs) > 2 {
					err = c.UnsubscribeFromNewTxs()
					newTxErrChan <- err
					wg.Done()
				}
			},
		)

		require.NoError(t, err)
		select {
		case err := <-newTxErrChan:
			require.NoError(t, err)
		case <-time.After(30 * time.Second):
			require.Fail(t, "timeout waiting for transactions")
		}
		wg.Wait()

		receive := make(chan struct{})

		txHashes := make([]string, 0, len(txs))
		for txHash := range txs {
			txHashes = append(txHashes, txHash)
		}

		onTxStatusParams := &OnTxStatusParams{
			Transactions: []string{txs[txHashes[0]]},
		}

		// test monitor without existing subscription
		err = c.OnTxStatus(context.Background(), *onTxStatusParams)
		require.Error(t, err)

		onTxStatusParams.Callback = func(ctx context.Context, err error, result *OnTxStatusNotification) {
			require.NoError(t, err)
			if result.TxHash == txHashes[0] {
				// add second transaction to monitor
				monitorTxsParams := &MonitorTxsParams{
					Transactions: []string{txs[txHashes[1]]},
				}
				err = c.MonitorTxs(context.Background(), monitorTxsParams)
				require.NoError(t, err)
			} else if result.TxHash == txHashes[1] {
				close(receive)
			}
		}

		err = c.OnTxStatus(context.Background(), *onTxStatusParams)
		require.NoError(t, err)

		select {
		case <-receive:
		case <-time.After(time.Minute):
			require.Fail(t, "timeout waiting for the second transaction")
		}

		// stop monitoring both txs
		stopMonitoringTxParams := &StopMonitoringTxParams{
			TransactionHash: txHashes,
			Transactions:    []string{txs[txHashes[0]], txs[txHashes[1]]},
		}

		err = c.StopMonitoringTx(context.Background(), stopMonitoringTxParams)
		require.NoError(t, err)

		// close the client
		err = c.Close()
		assert.NoError(t, err)
	}
}
