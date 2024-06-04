package contentcampaign

var (
	_ error = RemoteESError{}
)

// RemoteESError will be returned when the Content campaign response is invalid from the content elastic search.
type RemoteESError struct {
	Err error
}

// Error func definition
func (ree RemoteESError) Error() string {
	return ree.Err.Error()
}
