package bloxroute_sdk_go

type NoopLogger struct{}

func (n *NoopLogger) Debug(...interface{}) {}

func (n *NoopLogger) Debugf(string, ...interface{}) {}

func (n *NoopLogger) Info(...interface{}) {}

func (n *NoopLogger) Infof(string, ...interface{}) {}

func (n *NoopLogger) Warn(...interface{}) {}

func (n *NoopLogger) Warnf(string, ...interface{}) {}

func (n *NoopLogger) Error(...interface{}) {}

func (n *NoopLogger) Errorf(string, ...interface{}) {}
