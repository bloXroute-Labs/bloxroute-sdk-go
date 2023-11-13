package bloxroute_sdk_go

type noopLogger struct{}

func (n *noopLogger) Debug(...interface{}) {}

func (n *noopLogger) Debugf(string, ...interface{}) {}

func (n *noopLogger) Info(...interface{}) {}

func (n *noopLogger) Infof(string, ...interface{}) {}

func (n *noopLogger) Warn(...interface{}) {}

func (n *noopLogger) Warnf(string, ...interface{}) {}

func (n *noopLogger) Error(...interface{}) {}

func (n *noopLogger) Errorf(string, ...interface{}) {}
