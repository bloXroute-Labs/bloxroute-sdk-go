package bloxroute_sdk_go

type noopLogger struct{}

func (n *noopLogger) Debug(args ...interface{}) {}

func (n *noopLogger) Debugf(format string, args ...interface{}) {}

func (n *noopLogger) Info(args ...interface{}) {}

func (n *noopLogger) Infof(format string, args ...interface{}) {}

func (n *noopLogger) Warn(args ...interface{}) {}

func (n *noopLogger) Warnf(format string, args ...interface{}) {}

func (n *noopLogger) Error(args ...interface{}) {}

func (n *noopLogger) Errorf(format string, args ...interface{}) {}
