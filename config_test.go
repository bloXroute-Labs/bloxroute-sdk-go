package bloxroute_sdk_go

import (
	"os"
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
