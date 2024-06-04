package reward

var (
	_ error = DuplicatedError{}
	_ error = UnprocessableError{}
	_ error = RemoteError{}
)

// BadRequestError will be returned when the reward request is invalid
type BadRequestError struct {
	cause string
}

// NewBadRequestError returns new BadRequestError with the cause
func NewBadRequestError(cause string) error {
	return &BadRequestError{cause}
}

// Error func definition
func (bre BadRequestError) Error() string {
	return bre.cause
}

// DuplicatedError will be returned when the reward request is duplicated
type DuplicatedError struct {
}

// Error func definition
func (DuplicatedError) Error() string {
	return "duplicated request"
}

// UnprocessableError will be returned when the reward request can't be served
type UnprocessableError struct {
}

// Error func definition
func (UnprocessableError) Error() string {
	return "unprocessable request"
}

// RemoteError will be returned when the reward response is not valid
type RemoteError struct {
	Err error
}

// Error func definition
func (rre RemoteError) Error() string {
	return rre.Err.Error()
}
