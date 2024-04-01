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

	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/bloXroute-Labs/gateway/v2/jsonrpc"
	pb "github.com/bloXroute-Labs/gateway/v2/protobuf"
	"github.com/bloXroute-Labs/gateway/v2/types"
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
	case types.UserIntentsFeed:
		params := req.(*pb.IntentsRequest)
		var stream pb.Gateway_IntentsClient
		stream, err = h.client.Intents(subCtx, params)
		if err != nil {
			cancel()
			return fmt.Errorf("failed to subscribe to %s: %w", feed, err)
		}
		wrapStream = func() (any, error) {
			return stream.Recv()
		}
	case types.UserIntentSolutionsFeed:
		params := req.(*pb.IntentSolutionsRequest)
		var stream pb.Gateway_IntentSolutionsClient
		stream, err = h.client.IntentSolutions(subCtx, params)
		if err != nil {
			cancel()
			return fmt.Errorf("failed to subscribe to %s: %w", feed, err)
		}
		wrapStream = func() (any, error) {
			return stream.Recv()
		}
	case types.TxReceiptsFeed:
		params := req.(*TxReceiptParams)
		var stream pb.Gateway_TxReceiptsClient
		stream, err = h.client.TxReceipts(subCtx, &pb.TxReceiptsRequest{Includes: params.Include})
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

	case RPCSubmitIntent:
		submitIntentReq, ok := params.(*pb.SubmitIntentRequest)
		if !ok {
			return nil, fmt.Errorf("failed to cast params: expected %T, got %T", &pb.SubmitIntentRequest{}, params)
		}
		reply, err := h.client.SubmitIntent(ctx, submitIntentReq)
		if err != nil {
			return nil, fmt.Errorf("failed to submit intent: %w", err)
		}

		responseJSON, err := json.Marshal(reply)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal submit intent response: %w", err)
		}

		responseRawMessage := json.RawMessage(responseJSON)
		return &responseRawMessage, nil

	case RPCSubmitIntentSolution:
		submitIntentSolutionReq, ok := params.(*pb.SubmitIntentSolutionRequest)
		if !ok {
			return nil, fmt.Errorf("failed to cast params: expected %T, got %T", &pb.SubmitIntentSolutionRequest{}, params)
		}
		reply, err := h.client.SubmitIntentSolution(ctx, submitIntentSolutionReq)
		if err != nil {
			return nil, fmt.Errorf("failed to submit intent solution: %w", err)
		}

		responseJSON, err := json.Marshal(reply)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal submit intent solution response: %w", err)
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
							From:  string(fv.From),
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
						BlobGasUsed:      resp.Header.BlobGasUsed,
						ExcessBlobGas:    resp.Header.ExcessBlobGas,
					}
				}
				if resp.Header.BaseFeePerGas != "" {
					baseFee, err := strconv.Atoi(resp.Header.BaseFeePerGas)
					if err != nil {
						callback(ctx, err, nil)
						continue
					}
					header.BaseFeePerGas = &baseFee
				}
				if resp.Header.WithdrawalsRoot != "" {
					withdrawalsRoot := common.BytesToHash([]byte(resp.Header.WithdrawalsRoot))
					header.WithdrawalsRoot = &withdrawalsRoot
				}
				if resp.Header.ParentBeaconRoot != "" {
					parentBeaconRoot := common.BytesToHash([]byte(resp.Header.ParentBeaconRoot))
					header.ParentBeaconRoot = &parentBeaconRoot
				}
				var withdrawals []OnBlockWithdrawal
				if resp.Withdrawals != nil {
					withdrawals = make([]OnBlockWithdrawal, len(resp.Withdrawals))
					for i, w := range resp.Withdrawals {
						withdrawals[i] = OnBlockWithdrawal{
							Address:        w.Address,
							Amount:         w.Amount,
							Index:          w.Index,
							ValidatorIndex: w.ValidatorIndex,
						}
					}
				}

				result = &OnBdnBlockNotification{
					Hash:                resp.Hash,
					Header:              header,
					FutureValidatorInfo: futureValidatorInfo,
					Transactions:        transactions,
					Withdrawals:         withdrawals,
				}
			case types.TxReceiptsFeed:
				resp := rawResult.(*pb.TxReceiptsReply)

				var logs []OnTxReceiptNotificationLog
				for _, log := range resp.Logs {
					logs = append(logs, OnTxReceiptNotificationLog{
						Address:          log.Address,
						Topics:           log.Topics,
						Data:             log.Data,
						BlockNumber:      log.BlockNumber,
						TransactionHash:  log.TransactionHash,
						TransactionIndex: log.TransactionIndex,
						BlockHash:        log.BlockHash,
						LogIndex:         log.LogIndex,
						Removed:          log.Removed,
					})
				}

				result = &OnTxReceiptNotification{
					BlockHash:         resp.BlocKHash,
					BlockNumber:       resp.BlockNumber,
					ContractAddress:   resp.ContractAddress,
					CumulativeGasUsed: resp.CumulativeGasUsed,
					EffectiveGasUsed:  resp.EffectiveGasUsed,
					From:              resp.From,
					GasUsed:           resp.GasUsed,
					Logs:              logs,
					LogsBloom:         resp.LogsBloom,
					Status:            resp.Status,
					To:                resp.To,
					TransactionHash:   resp.TransactionHash,
					TransactionIndex:  resp.TransactionIndex,
					Type:              resp.Type,
					TxsCount:          resp.TxsCount,
					BlobGasUsed:       resp.BlobGasUsed,
					BlobGasPrice:      resp.BlobGasPrice,
				}

			case types.UserIntentsFeed:
				resp := rawResult.(*pb.IntentsReply)
				result = &OnIntentsNotification{
					DappAddress:   resp.DappAddress,
					SenderAddress: resp.SenderAddress,
					IntentID:      resp.IntentId,
					Intent:        resp.Intent,
					Timestamp:     resp.Timestamp.String(),
				}
			case types.UserIntentSolutionsFeed:
				resp := rawResult.(*pb.IntentSolutionsReply)
				result = &OnIntentSolutionsNotification{
					IntentID:       resp.IntentId,
					IntentSolution: resp.IntentSolution,
				}
			}

			callback(ctx, nil, result)
		}
	}()
}
