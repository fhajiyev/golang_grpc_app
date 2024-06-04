package log

import (
	"fmt"
)

// FmtWrapper wraps fmt to fit logger interface
type FmtWrapper struct {
}

// Logf definition
func (*FmtWrapper) Logf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

// NewFmtWrapper returns FmtWrapper
func NewFmtWrapper() *FmtWrapper {
	return &FmtWrapper{}
}
