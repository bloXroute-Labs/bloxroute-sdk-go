package bloxroute_sdk_go

import (
	"fmt"

	"github.com/sourcegraph/jsonrpc2"
)

type RPCError jsonrpc2.Error

// Error implements the Go error interface.
func (e *RPCError) Error() string {
	if e == nil {
		return ""
	}

	if e.Data == nil {
		return fmt.Sprintf("code: %v message: %s", e.Code, e.Message)
	}

	bytes, err := e.Data.MarshalJSON()
	if err != nil {
		return fmt.Sprintf("code: %v message: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("code: %v message: %s, data: %s", e.Code, e.Message, string(bytes))
}
