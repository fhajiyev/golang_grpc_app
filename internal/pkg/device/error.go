package device

import "fmt"

var (
	_ error = &RemoteProfileError{}
)

// RemoteProfileError struct definition
type RemoteProfileError struct {
	Err error
}

// InvalidArgumentError struct definition
type InvalidArgumentError struct {
	ArgName  string
	ArgValue interface{}
}

// Error func definition
func (pre RemoteProfileError) Error() string {
	return pre.Err.Error()
}

// Error func definition
func (iae InvalidArgumentError) Error() string {
	return fmt.Sprintf("Argument %v : %v is invalid", iae.ArgName, iae.ArgValue)
}
