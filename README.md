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
