package bloxroute_sdk_go

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/bloXroute-Labs/gateway/v2/jsonrpc"
	pb "github.com/bloXroute-Labs/gateway/v2/protobuf"
	"github.com/bloXroute-Labs/gateway/v2/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type grpcHandler struct {
	hst           handlerSourceType
	config        *Config
	conn          *grpc.ClientConn
	client        pb.GatewayClient
	md            metadata.MD
	stop          chan struct{}
	wg            *sync.WaitGroup
	subscriptions map[types.FeedType]grpcSubscription
	lock          *sync.Mutex
}

type grpcSubscription struct {
	callback CallbackFunc[any]
	subReq   any
	cancel   context.CancelFunc
	wait     chan struct{}
}

// Type returns the handler type
func (h *grpcHandler) Type() handlerSourceType {
	return h.hst
}

// Subscribe subscribes to a feed
func (h *grpcHandler) Subscribe(ctx context.Context, feed types.FeedType, req any, callback CallbackFunc[any]) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	_, ok := h.subscriptions[feed]
	if ok {
		return fmt.Errorf("feed %v already subscribed", feed)
	}

	ctx = metadata.NewOutgoingContext(ctx, h.md)

	subCtx, cancel := context.WithCancel(ctx)
	var wrapStream func() (any, error)
	var err error

	switch feed {
	case types.NewTxsFeed:
		params := req.(*NewTxParams)
		var stream pb.Gateway_NewTxsClient
		stream, err = h.client.NewTxs(subCtx, &pb.TxsRequest{Filters: params.Filters, Includes: params.Include})
		if err != nil {
			cancel()
			return fmt.Errorf("failed to subscribe to %s: %w", feed, err)
		}
		wrapStream = func() (any, error) {
			return stream.Recv()
		}
	case types.PendingTxsFeed:
		params := req.(*PendingTxParams)
		var stream pb.Gateway_PendingTxsClient
		stream, err = h.client.PendingTxs(subCtx, &pb.TxsRequest{Filters: params.Filters, Includes: params.Include})
		if err != nil {
			cancel()
			return fmt.Errorf("failed to subscribe to %s: %w", feed, err)
		}

		wrapStream = func() (any, error) {
			return stream.Recv()
		}
	case types.NewBlocksFeed:
		params := req.(*NewBlockParams)
		var stream pb.Gateway_NewBlocksClient
		stream, err = h.client.NewBlocks(subCtx, &pb.BlocksRequest{Includes: params.Include})
		if err != nil {
			cancel()
			return fmt.Errorf("failed to subscribe to %s: %w", feed, err)
		}
		wrapStream = func() (any, error) {
			return stream.Recv()
		}
	case types.BDNBlocksFeed:
		params := req.(*BdnBlockParams)
		var stream pb.Gateway_BdnBlocksClient
		stream, err = h.client.BdnBlocks(subCtx, &pb.BlocksRequest{Includes: params.Include})
		if err != nil {
			cancel()
			return fmt.Errorf("failed to subscribe to %s: %w", feed, err)
		}
		wrapStream = func() (any, error) {
			return stream.Recv()
		}
	default:
		cancel()
		return fmt.Errorf("%s feed type is not yet supported", feed)
	}

	h.sub(subCtx, cancel, feed, wrapStream, req, callback)

	return nil
}

// Request sends a gRPC request
func (h *grpcHandler) Request(ctx context.Context, method jsonrpc.RPCRequestType, params any) (*json.RawMessage, error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	ctx = metadata.NewOutgoingContext(ctx, h.md)

	switch method {
	case jsonrpc.RPCTx:
		sendTxParams, ok := params.(*SendTxParams)
		if !ok {
			return nil, fmt.Errorf("failed to cast params: expected %T, got %T", &SendTxParams{}, params)
		}

		reply, err := h.client.BlxrTx(ctx, &pb.BlxrTxRequest{
			Transaction:     sendTxParams.Transaction,
			NonceMonitoring: sendTxParams.NonceMonitoring,
			NextValidator:   sendTxParams.NextValidator,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to send tx: %w", err)
		}

		responseMap := map[string]string{"tx_hash": reply.TxHash}
		responseJSON, err := json.Marshal(responseMap)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tx hash: %w", err)
		}

		responseRawMessage := json.RawMessage(responseJSON)
		return &responseRawMessage, nil

	default:
		return nil, fmt.Errorf("%s grpc request is not yet supported", method)
	}
}

// UnsubscribeRetry unsubscribes from a feed
func (h *grpcHandler) UnsubscribeRetry(f types.FeedType) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	sub, ok := h.subscriptions[f]
	if !ok {
		return fmt.Errorf("feed %v not subscribed", f)
	}

	sub.cancel()

	select {
	case <-sub.wait:
	case <-time.After(5 * time.Second):
	}

	delete(h.subscriptions, f)

	return nil
}

// Close closes the gRPC connection
func (h *grpcHandler) Close() error {
	err := h.conn.Close()

	h.wg.Wait()

	return err
}

func (h *grpcHandler) sub(ctx context.Context, cancel context.CancelFunc, f types.FeedType, stream func() (any, error), req any, callback CallbackFunc[any]) {
	wait := make(chan struct{})

	h.subscriptions[f] = grpcSubscription{
		callback: callback,
		subReq:   req,
		cancel:   cancel,
		wait:     wait,
	}

	h.wg.Add(1)
	go func() {
		defer h.wg.Done() // global wait group
		defer close(wait) // local signal channel for the subscription

		for {
			rawResult, err := stream()
			if err != nil {
				rpcErr, ok := status.FromError(err)
				if (ok && rpcErr.Code() == codes.Canceled) || errors.Is(err, io.EOF) {
					break
				}

				h.config.Logger.Errorf("failed to receive response from %s stream: %v", f, err)

				callback(ctx, err, nil)
				continue
			}

			var result any

			switch f {
			case types.NewTxsFeed, types.PendingTxsFeed:
				resp := rawResult.(*pb.TxsReply)
				for i := range resp.Tx {
					result = &NewTxNotification{
						TxContents: &NewTxNotificationTxContents{
							From: string(resp.Tx[i].From),
						},
						LocalRegion: resp.Tx[i].LocalRegion,
						Time:        strconv.FormatInt(resp.Tx[i].Time, 10),
						RawTx:       string(resp.Tx[i].RawTx),
					}
					callback(ctx, nil, result)
				}
				continue
			case types.NewBlocksFeed, types.BDNBlocksFeed:
				resp := rawResult.(*pb.BlocksReply)
				var futureValidatorInfo []FutureValidatorInfo
				if resp.FutureValidatorInfo != nil {
					futureValidatorInfo = make([]FutureValidatorInfo, len(resp.FutureValidatorInfo))
					for i, fv := range resp.FutureValidatorInfo {
						futureValidatorInfo[i] = FutureValidatorInfo{
							BlockHeight: fv.BlockHeight,
							WalletId:    fv.WalletId,
							Accessible:  fv.Accessible,
						}
					}
				}
				var transactions []OnNewBlockTransaction
				if resp.Transaction != nil {
					transactions = make([]OnNewBlockTransaction, len(resp.Transaction))
					for i, fv := range resp.Transaction {
						transactions[i] = OnNewBlockTransaction{
							From:  fv.From,
							RawTx: fv.RawTx,
						}
					}
				}
				var header *Header
				if resp.Header != nil {
					header = &Header{
						ParentHash:       resp.Header.WithdrawalsRoot,
						Sha3Uncles:       resp.Header.Sha3Uncles,
						Miner:            resp.Header.Miner,
						StateRoot:        resp.Header.StateRoot,
						TransactionsRoot: resp.Header.TransactionsRoot,
						ReceiptsRoot:     resp.Header.ReceiptsRoot,
						LogsBloom:        resp.Header.LogsBloom,
						Difficulty:       resp.Header.Difficulty,
						Number:           resp.Header.Number,
						GasLimit:         resp.Header.GasLimit,
						GasUsed:          resp.Header.GasUsed,
						Timestamp:        resp.Header.Timestamp,
						ExtraData:        resp.Header.ExtraData,
						MixHash:          resp.Header.MixHash,
						Nonce:            resp.Header.Nonce,
					}
				}
				result = &OnBdnBlockNotification{
					Hash:                resp.Hash,
					Header:              header,
					FutureValidatorInfo: futureValidatorInfo,
					Transactions:        transactions,
				}
			}

			callback(ctx, nil, result)
		}
	}()
}
