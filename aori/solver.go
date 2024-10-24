package aori

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/valyala/fastjson"
	"golang.org/x/sync/errgroup"

	sdk "github.com/bloXroute-Labs/bloxroute-sdk-go"
)

var (
	ErrSDKClientIsRequired = errors.New("SDK client is required")
	ErrMissingHandler      = errors.New("one or more aori event handlers are missing")
	ErrMissingDataField    = errors.New("'data' field is missing from the intent")
)

// Solver represents aori solver
type Solver struct {
	client           *sdk.Client
	logger           sdk.Logger
	onIntentsRequest *sdk.IntentsParams
	solverPrivateKey *ecdsa.PrivateKey
	handlers         SolverHandlers
}

// SolverHandlers represents aori solver event handlers
type SolverHandlers struct {
	QuoteRequested func(ctx context.Context, quote *RfqQuoteRequest) (*RfqSolution, error)
	QuoteReceived  func(ctx context.Context, quote *RfqQuoteReceived) error
	CallData       func(ctx context.Context, data *RfqCallDataToExecute) error
}

// NewSolver creates a new solver instance
func NewSolver(client *sdk.Client, solverPrivateKey, aoriDAppAddress string, handlers SolverHandlers) (*Solver, error) {
	if client == nil {
		return nil, ErrSDKClientIsRequired
	}

	if handlers.QuoteRequested == nil || handlers.QuoteReceived == nil || handlers.CallData == nil {
		return nil, ErrMissingHandler
	}

	key, err := crypto.HexToECDSA(solverPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse solver private key: %v", err)
	}

	solverAddress := crypto.PubkeyToAddress(key.PublicKey).String()
	solverHash := crypto.Keccak256Hash([]byte(solverAddress)).Bytes()
	solverSignature, err := crypto.Sign(solverHash, key)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate signature: %v", err)
	}

	return &Solver{
		client: client,
		logger: &sdk.NoopLogger{},
		onIntentsRequest: &sdk.IntentsParams{
			DappAddress:   aoriDAppAddress,
			SolverAddress: solverAddress,
			Hash:          solverHash,
			Signature:     solverSignature,
		},
		solverPrivateKey: key,
		handlers:         handlers,
	}, nil
}

// WithLogger sets the logger for the solver
func (s *Solver) WithLogger(logger sdk.Logger) *Solver {
	s.logger = logger

	return s
}

// Start starts the solver
func (s *Solver) Start(ctx context.Context) error {
	eventsGroup := &errgroup.Group{}

	s.logger.Infof("starting intents subscription for dApp %s", s.onIntentsRequest.DappAddress)

	err := s.client.OnIntents(ctx, s.onIntentsRequest, func(ctx context.Context, err error, notification *sdk.OnIntentsNotification) {
		if err != nil {
			s.logger.Errorf("failed to receive intents notification: %v", err)
			return
		}

		s.logger.Debugf("received intent notification %s for dApp %s", notification.IntentID, notification.DappAddress)

		event, err := s.parseEvent(notification)
		if err != nil {
			s.logger.Errorf("failed to parse intent: %v", err)
			return
		}

		err = s.handleEvent(ctx, eventsGroup, event)
		if err != nil {
			s.logger.Error(err.Error())
		}
	})
	if err != nil {
		return fmt.Errorf("failed to start intents subscription: %v", err)
	}

	<-ctx.Done()

	return eventsGroup.Wait()
}

func (s *Solver) handleEvent(ctx context.Context, eventsGroup *errgroup.Group, event *aoriEvent) error {
	s.logger.Debugf("handling event '%s'", event.name)

	switch event.name {
	case EventRfqQuoteRequested:
		quoteRequested, err := unmarshalAny[RfqQuoteRequest](event.data)
		if err != nil {
			return fmt.Errorf("failed to unmarshal %s event: %v", EventRfqQuoteRequested, err)
		}

		eventsGroup.Go(func() error {
			err := s.handleAoriRfqQuoteRequest(ctx, quoteRequested, event.intentID)
			if err != nil {
				s.logger.Errorf("failed to handle %s event: %v", EventRfqQuoteRequested, err)
			}

			return nil
		})

	case EventRfqQuoteReceived:
		quoteReceived, err := unmarshalAny[RfqQuoteReceived](event.data)
		if err != nil {
			return fmt.Errorf("failed to unmarshal %s event: %v", EventRfqQuoteReceived, err)
		}

		eventsGroup.Go(func() error {
			err := s.handleAoriRfqQuoteReceived(ctx, quoteReceived)
			if err != nil {
				s.logger.Errorf("failed to handle %s event: %v", EventRfqQuoteReceived, err)
			}

			return nil
		})
	case EventRfqCallDataToExecute:
		dataToExecute, err := unmarshalAny[RfqCallDataToExecute](event.data)
		if err != nil {
			return fmt.Errorf("failed to unmarshal %s event: %v", EventRfqCallDataToExecute, err)
		}

		eventsGroup.Go(func() error {
			err := s.handleAoriRfqCallDataToExecute(ctx, dataToExecute)
			if err != nil {
				s.logger.Errorf("failed to handle %s event: %v", EventRfqCallDataToExecute, err)
			}

			return nil
		})
	default:
		s.logger.Errorf("event %s is not supported", event.name)
	}

	return nil
}

func (s *Solver) parseEvent(notification *sdk.OnIntentsNotification) (*aoriEvent, error) {
	rawIntent := make([]byte, base64.StdEncoding.DecodedLen(len(notification.Intent)))
	_, err := base64.StdEncoding.Decode(rawIntent, notification.Intent)
	if err == nil {
		notification.Intent = rawIntent
	}

	// remove null bytes from the input data
	notification.Intent = bytes.Trim(notification.Intent, "\x00")

	var p fastjson.Parser
	v, err := p.ParseBytes(notification.Intent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse intent: %v", err)
	}

	eventData := v.GetObject("data")
	if eventData == nil {
		return nil, ErrMissingDataField
	}

	return &aoriEvent{
		name:     string(v.GetStringBytes("event")),
		intentID: notification.IntentID,
		data:     []byte(eventData.String()),
	}, nil
}

func unmarshalAny[T any](bytes []byte) (*T, error) {
	out := new(T)
	if err := json.Unmarshal(bytes, out); err != nil {
		return nil, err
	}

	return out, nil
}
