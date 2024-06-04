package mq

// ClosedSessionError definition
type ClosedSessionError struct {
}

func (ClosedSessionError) Error() string {
	return "closed session"
}

// ClosedConnectionError definition
type ClosedConnectionError struct {
}

func (ClosedConnectionError) Error() string {
	return "closed connection"
}

// ClosedChannelError definition
type ClosedChannelError struct {
}

func (ClosedChannelError) Error() string {
	return "closed channel"
}
