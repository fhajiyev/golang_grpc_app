package log

import "encoding/json"

// StructuredLogger definition
type StructuredLogger struct {
	logger logger
}

// Logger interface
type logger interface {
	Logf(string, ...interface{})
}

// Log method
func (l *StructuredLogger) Log(m map[string]interface{}) {
	marshaled, _ := json.Marshal(m)
	l.logger.Logf("%s\n", string(marshaled))
}

// NewStructuredLogger creates structuredLogger
func NewStructuredLogger(logger logger) *StructuredLogger {
	return &StructuredLogger{
		logger: logger,
	}
}
