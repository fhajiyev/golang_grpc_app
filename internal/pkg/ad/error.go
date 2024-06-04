package ad

var (
	_ error = RemoteError{}
)

// RemoteError will be returned when the buzzad doesn't give ad response
type RemoteError struct {
	Err error
}

// Error will return message of an error
func (e RemoteError) Error() string {
	return e.Err.Error()
}
