package bloxroute_sdk_go

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"testing"
)

type testURL string

const (
	wsCloudApiUrl  testURL = "WS_CLOUD_API_URL"
	wsGatewayUrl   testURL = "WS_GATEWAY_URL"
	grpcGatewayUrl testURL = "GRPC_GATEWAY_URL"
)

func testConfig(t *testing.T, url testURL) *Config {
	t.Helper()

	c := &Config{
		AuthHeader: os.Getenv("AUTH_HEADER"),
	}
	switch url {
	case wsCloudApiUrl:
		c.WSCloudAPIURL = os.Getenv(string(wsCloudApiUrl))
	case wsGatewayUrl:
		c.WSGatewayURL = os.Getenv(string(wsGatewayUrl))
	case grpcGatewayUrl:
		c.GRPCGatewayURL = os.Getenv(string(grpcGatewayUrl))
	default:
		t.Fatalf("unknown test url: %s", url)
	}

	return c
}

// contextWithSignal returns a context that is cancelled when the process receives the given termination signal.
func contextWithSignal(parent context.Context, s ...os.Signal) context.Context {
	if len(s) == 0 {
		s = []os.Signal{syscall.SIGTERM, syscall.SIGINT}
	}
	ctx, cancel := context.WithCancel(parent)
	c := make(chan os.Signal, 1)
	signal.Notify(c, s...)

	go func() {
		// wait for either the signal, or for the context to be cancelled
		select {
		case <-c:
		case <-parent.Done():
		}

		// cancel the context
		// and stop waiting for more signals (next sigterm will terminate the process)
		cancel()
		signal.Stop(c)
	}()

	return ctx
}
