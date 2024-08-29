package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestClient_GetSolutionsForIntent(t *testing.T) {
	t.Run("ws_gateway", testGetSolutionsForIntentWithDAppAddr(wsGatewayUrl))
	t.Run("ws_gateway", testGetSolutionsForIntentWithSenderAddr(wsGatewayUrl))
}

func testGetSolutionsForIntentWithDAppAddr(url testURL) func(t *testing.T) {
	return testGetSolutionsForIntent(url, true)
}

func testGetSolutionsForIntentWithSenderAddr(url testURL) func(t *testing.T) {
	return testGetSolutionsForIntent(url, false)
}

func testGetSolutionsForIntent(url testURL, useDApp bool) func(t *testing.T) {
	return func(t *testing.T) {
		config := testConfig(t, url)

		ctx := contextWithSignal(context.Background())

		c, err := NewClient(ctx, config)
		require.NoError(t, err)

		submitIntentParams := createdSubmitIntentParams(t)

		subRep, err := c.SubmitIntent(ctx, submitIntentParams)
		require.NoError(t, err)

		var resp map[string]string
		err = json.Unmarshal(*subRep, &resp)
		require.NoError(t, err)
		intentID := resp["intent_id"]
		require.NotEmpty(t, intentID)

		group, gCtx := errgroup.WithContext(ctx)

		solutionsNum := 5

		for i := 0; i < solutionsNum; i++ {
			group.Go(func() error {
				r := rand.Intn(50)
				time.Sleep(time.Duration(r) * time.Millisecond)

				_, err = c.SubmitIntentSolution(gCtx, createSubmitIntentSolutionParams(t, intentID, []byte(fmt.Sprintf("test intent solution %d", r))))

				return err
			})
		}

		getSolutionsForIntentParams := &GetSolutionsForIntentParams{
			IntentID: intentID,
		}
		if useDApp {
			getSolutionsForIntentParams.DAppOrSenderPrivateKey = submitIntentParams.DappPrivateKey
		} else {
			getSolutionsForIntentParams.DAppOrSenderPrivateKey = submitIntentParams.SenderPrivateKey
		}

		group.Go(func() error {
			ticker := time.NewTicker(10 * time.Millisecond)

			for {
				select {
				case <-ticker.C:
					solutions, err := c.GetSolutionsForIntent(gCtx, getSolutionsForIntentParams)
					if err != nil {
						return err
					}

					var m []map[string]interface{}
					err = json.Unmarshal(*solutions, &m)
					if err != nil {
						return fmt.Errorf("failed to unmarshal get solutions for intent response: %v", err)
					}
					if len(m) == solutionsNum {
						return nil
					}
				case <-time.After(time.Minute):
					return fmt.Errorf("timeout waiting for %d solutions", solutionsNum)
				case <-gCtx.Done():
					return nil
				}
			}
		})

		require.NoError(t, group.Wait())
	}
}
