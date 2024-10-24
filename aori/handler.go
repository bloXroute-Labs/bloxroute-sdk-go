package aori

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	solsha3 "github.com/miguelmota/go-solidity-sha3"
	"github.com/valyala/fastjson"

	"github.com/bloXroute-Labs/gateway/v2/logger"

	sdk "github.com/bloXroute-Labs/bloxroute-sdk-go"
)

func (s *Solver) handleAoriRfqQuoteRequest(ctx context.Context, quote *RfqQuoteRequest, intentID string) error {
	solution, err := s.handlers.QuoteRequested(ctx, quote)
	if err != nil {
		return err
	}

	// force the offerer to be the solver address
	solution.Offerer = s.onIntentsRequest.SolverAddress

	respHash := getOrderHash(solution)
	prefixed := crypto.Keccak256Hash([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(respHash), respHash)))

	quoteSig, err := crypto.Sign(prefixed.Bytes(), s.solverPrivateKey) // signing the hash
	if err != nil {
		return err
	}

	quoteResponse := &RfqQuoteResponse{
		RfqId:     quote.RfqId,
		Order:     solution,
		Signature: hexutil.Encode(quoteSig),
	}

	payload := &sendPayload{
		Id:      1,
		Jsonrpc: "2.0",
		Method:  "aori_respond",
		Params:  []interface{}{quoteResponse},
	}

	solutionBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload with quoteResponse: %v", err)
	}

	params := &sdk.SubmitIntentSolutionParams{
		IntentSolution: solutionBytes,
		IntentID:       intentID,
	}

	params.SolverAddress = crypto.PubkeyToAddress(s.solverPrivateKey.PublicKey).Hex()
	params.Hash = crypto.Keccak256Hash(params.IntentSolution).Bytes()
	params.Signature, err = crypto.Sign(params.Hash, s.solverPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to sign intentSolutionHash: %v", err)
	}

	resp, err := s.client.SubmitIntentSolution(ctx, params)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return fmt.Errorf("failed to submit intent solution: %v", err)
	}

	var p fastjson.Parser
	v, err := p.ParseBytes(*resp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal submit intent solution response: %v", err)
	}

	solutionID := string(v.GetStringBytes("solution_id"))

	logger.Debugf("submitted intent solution for intentID %s, solverAddress %s, solutionID %s", intentID, params.SolverAddress, solutionID)

	return nil
}

func (s *Solver) handleAoriRfqQuoteReceived(ctx context.Context, req *RfqQuoteReceived) error {
	return s.handlers.QuoteReceived(ctx, req)
}

func (s *Solver) handleAoriRfqCallDataToExecute(ctx context.Context, req *RfqCallDataToExecute) error {
	return s.handlers.CallData(ctx, req)
}

func getOrderHash(resp *RfqSolution) []byte {
	hash := solsha3.SoliditySHA3(
		[]string{
			"address",
			"address",
			"uint256",
			"uint256",
			"address",
			"address",
			"uint256",
			"uint256",
			"address",
			"uint256",
			"uint256",
			"uint256",
			"uint256",
			"bool",
		},
		[]interface{}{
			resp.Offerer,
			resp.InputToken,
			resp.InputAmount,
			strconv.Itoa(resp.InputChainId),
			resp.InputZone,
			resp.OutputToken,
			resp.OutputAmount,
			strconv.Itoa(resp.OutputChainId),
			resp.OutputZone,
			resp.StartTime,
			resp.EndTime,
			resp.Salt,
			resp.Counter,
			resp.ToWithdraw,
		},
	)
	return hash
}
