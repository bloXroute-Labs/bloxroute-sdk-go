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
import "github.com/bloxroute-inc/bloxroute-cloud-go-sdk/sdk"
```

Initialize the SDK Client using Auth Header and WS URL (either cloud API or a gateway):

```go
// create a config
config := &Config{
	AuthHeader: os.Getenv("af84h0p4TR79MKqh909b9yj4BwxxGL4ueWm0QZiCB88OzYelc7QOG2GB9QPMUefZ01wsgu7efSL4Mj6m6KPp0qFhN74m"),
	CloudAPIURL: os.Getenv("wss://8.210.133.198/ws"),
}

// create a new client
c, err := NewClient(config)
if err != nil {
    // handle error
}

ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// run the client
go c.Run(ctx)
```

Subscribe to a feed:

```go

// subscribe to new transactions
err := c.OnNewTx(ctx, nil, func(ctx context.Context, result *json.RawMessage) {
    fmt.Println("new tx", string(*result))
})
if err != nil {
    // handle error
}
```

Unsubscribe from a feed:

```go
// unsubscribe from new transactions
err := c.UnsubscribeFromNewTxs()
if err != nil {
    // handle error
}
```

Stop the client:

```go
// stop the client
err = c.Close()
if err != nil {
    // handle error
}
```

Another option would be to cancel the context to stop the client:

```go
ctx, cancel := context.WithCancel(context.Background())

wg := &sync.WaitGroup{}
wg.Add(1)

go func() {
	defer wg.Done()
	
	err := c.Run(ctx)
	if err != nil { 
	    // handle error 
	}
}()

// cancel the context
cancel()

// wait for the client to stop
wg.Wait()
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
[contributing guide]: https://github.com/bloXroute-Labs/bloxroute-sdk-go-private/blob/master/CONTRIBUTING.md
