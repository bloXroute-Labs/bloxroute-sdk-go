# bloXroute Golang SDK

The bloXroute Cloud Golang SDK is a software development kit that allows developers to interact with the bloXroute Cloud API and Gateway.
It provides an easy-to-use interface to integrate bloXroute's blockchain infrastructure optimization services into your Golang applications.

## Prerequisites

Before using the bloXroute Cloud Golang SDK, make sure you have the following prerequisites:

- bloXroute credentials: you need to sign up for an [account][account] and the ["Authorization"][authorization] header from the Account Portal.
- Go programming language: you should have Go installed on your system. The bloXroute Golang SDK is compatible with Go versions 1.20 and above.

## Usage

To get started with the Bloxroute Cloud Golang SDK, follow these steps:

Import the SDK into your project:

```go
import (
    sdk "github.com/bloXroute-Labs/bloxroute-sdk-go"
)
```

Initialize the SDK Client using Auth Header and WS/gRPC URL (either cloud API or a gateway):

```go
// create a config
config := &sdk.Config{
	AuthHeader: "af84h0p4TR79MKqh909b9yj4BwxxGL4ueWm0QZiCB88OzYelc7QOG2GB9QPMUefZ01wsgu7efSL4Mj6m6KPp0qFhN74m",
	WSCloudAPIURL: "wss://8.210.133.198/ws",
}

// create a new client
c, err := sdk.NewClient(context.Background(), config)
if err != nil {
    log.Fatal(err)
}
```

Subscribe to a feed:

```go

// subscribe to new transactions
if err := c.OnNewTx(ctx, &sdk.NewTxParams{Include: []string{"raw_tx"}}, func(ctx context.Context, err error, result *sdk.NewTxNotification) {
    if err != nil {
        log.Fatal(err)
    }

    // handle result
}); err != nil {
    log.Fatal(err)
}
```

Unsubscribe from a feed:

```go
// unsubscribe from new transactions
if err := c.UnsubscribeFromNewTxs(); err != nil {
    log.Fatal(err)
}
```

Stop the client:

```go
// stop the client
if err = c.Close(); err != nil {
    log.Fatal(err)
}
```


## Intents example

```go
package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/bloXroute-Labs/bloxroute-sdk-go"
	"github.com/bloXroute-Labs/bloxroute-sdk-go/connection/ws"
	"github.com/ethereum/go-ethereum/crypto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var authHeader = "af84h0p4TR79MKqh909b9yj4BwxxGL4ueWm0QZiCB88OzYelc7QOG2GB9QPMUefZ01wsgu7efSL4Mj6m6KPp0qFhN74m"

func intentsGRPC(ctx context.Context) error {
	creds := credentials.NewClientTLSFromCert(nil, "")

	config := &sdk.Config{
		AuthHeader:     authHeader,
		GRPCGatewayURL: "grpc://germany-intents.blxrbdn.com:5005",
		GRPCDialOptions: []grpc.DialOption{
			grpc.WithTransportCredentials(creds),
		},
	}

	c, err := sdk.NewClient(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	return intents(ctx, c)
}

func intentsWS(ctx context.Context) error {
	credsWS := &ws.DialOptions{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	config := &sdk.Config{
		AuthHeader:    authHeader,
		WSGatewayURL:  "ws://virginia-intents.blxrbdn.com:28334/ws",
		WSDialOptions: credsWS,
	}

	c, err := sdk.NewClient(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	return intents(ctx, c)
}

func intents(ctx context.Context, c *sdk.Client) error {
	defer c.Close()

	onIntentsReq, solverPrivKey, err := onIntentsRequest()
	if err != nil {
		return fmt.Errorf("failed to create onIntents request: %v", err)
	}

	err = c.OnIntents(ctx, onIntentsReq, func(ctx context.Context, err error, notification *sdk.OnIntentsNotification) {
		fmt.Println("got intent")
		fmt.Println(notification)
		fmt.Println()
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to intents: %v", err)
	}

	onIntentSolutionsReq, dappPrivateKey, err := onIntentSolutionRequest()
	if err != nil {
		return fmt.Errorf("failed to create onIntentSolutions request: %v", err)
	}

	err = c.OnIntentSolutions(ctx, onIntentSolutionsReq, func(ctx context.Context, err error, notification *sdk.OnIntentSolutionsNotification) {
		fmt.Println("got intent solution")
		fmt.Println(notification)
		fmt.Println()
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to intent solutions: %v", err)
	}

	intent := []byte("test intent")
	intentHash := crypto.Keccak256Hash(intent).Bytes()
	intentSignature, err := crypto.Sign(intentHash, dappPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to sign intentHash: %v", err)
	}

	resp, err := c.SubmitIntent(ctx, &sdk.SubmitIntentParams{
		DappAddress:   onIntentSolutionsReq.DappAddress,
		SenderAddress: onIntentSolutionsReq.DappAddress,
		Intent:        intent,
		Hash:          intentHash,
		Signature:     intentSignature,
	})
	if err != nil {
		return fmt.Errorf("failed to submit intent: %v", err)
	}

	fmt.Println("submitted intent", string(*resp))

	var submitIntentResponse map[string]string
	err = json.Unmarshal(*resp, &submitIntentResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal submit intent response: %v", err)
	}

	intentSolution := []byte("test intent solution")
	intentSolutionHash := crypto.Keccak256Hash(intentSolution).Bytes()
	intentSolutionSignature, err := crypto.Sign(intentSolutionHash, solverPrivKey)
	if err != nil {
		return fmt.Errorf("failed to sign intentSolutionHash: %v", err)
	}

	resp, err = c.SubmitIntentSolution(ctx, &sdk.SubmitIntentSolutionParams{
		SolverAddress:  onIntentsReq.SolverAddress,
		IntentID:       submitIntentResponse["intent_id"],
		IntentSolution: intentSolution,
		Hash:           intentSolutionHash,
		Signature:      intentSolutionSignature,
	})

	if err != nil {
		return fmt.Errorf("failed to submit intent solution: %v", err)
	}

	fmt.Println("submitted intent solution", string(*resp))

	time.Sleep(5 * time.Second)

	return nil
}

func onIntentsRequest() (*sdk.IntentsParams, *ecdsa.PrivateKey, error) {
	/*
		privKey, err := crypto.HexToECDSA("703ba5e914dg0075c991ee19f2cd02d16929db029045a6e0d720ba8fbcd32222")
		if err != nil {
			return fmt.Errorf("failed to parse private key: %v", err)
		}
	*/

	privKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %v", err)
	}
	solverAddress := crypto.PubkeyToAddress(privKey.PublicKey).String()
	solverHash := crypto.Keccak256Hash([]byte(solverAddress)).Bytes()
	solverSignature, err := crypto.Sign(solverHash, privKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to sign solver hash: %v", err)
	}

	return &sdk.IntentsParams{
		SolverAddress:    solverAddress,
		Hash:             solverHash,
		Signature:        solverSignature,
	}, privKey, nil
}

func onIntentSolutionRequest() (*sdk.IntentSolutionsParams, *ecdsa.PrivateKey, error) {
	/*
		privKey, err := crypto.HexToECDSA("703ba5e914dg0075c991ee19f2cd02d16929db029045a6e0d720ba8fbcd32222")
		if err != nil {
			return fmt.Errorf("failed to parse private key: %v", err)
		}
	*/
	
	privKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %v", err)
	}
	signerAddress := crypto.PubkeyToAddress(privKey.PublicKey).String()
	hash := crypto.Keccak256Hash([]byte(signerAddress)).Bytes()
	sig, err := crypto.Sign(hash, privKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to sign hash: %v", err)
	}

	return &sdk.IntentSolutionsParams{
		DappAddress: signerAddress,
		Hash:        hash,
		Signature:   sig,
	}, privKey, nil
}
```

## Tests

To run the tests, the following environment variables should be set:

```bash
export AUTH_HEADER=af84h0p4TR79MKqh909b9yj4BwxxGL4ueWm0QZiCB88OzYelc7QOG2GB9QPMUefZ01wsgu7efSL4Mj6m6KPp0qFhN74m
export WS_CLOUD_API_URL=wss://api.blxrbdn.com/ws
export WS_GATEWAY_URL=ws://localhost:28334/ws
export GRPC_GATEWAY_URL=grpc://localhost:5001
```

## Contributing

Please read our [contributing guide] contributing guide

## Documentation

You can find our full technical documentation and architecture [on our website][documentation].

## Troubleshooting

Contact us at [our Discord] for further questions.

[account]: https://portal.bloxroute.com/register
[authorization]: https://docs.bloXroute.com/apis/authorization-headers
[white paper]: https://bloxroute.com/wp-content/uploads/2019/01/whitepaper-V1.1-1.pdf
[documentation]: https://docs.bloxroute.com/
[our Discord]: https://discord.gg/jHgpN8b
[contributing guide]: https://github.com/bloXroute-Labs/bloxroute-sdk-go/blob/master/CONTRIBUTING.md
