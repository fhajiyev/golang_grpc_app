package userreferral

import "fmt"

// APICallError happens when calling verifying API or giving referral reward API
type APICallError struct {
	Message string
}

// UserValidationError happens when referrer or referee validation fails
type UserValidationError struct {
	Message string
}

// NotFoundError happens when config or user is not found
type NotFoundError struct {
	Message string
}

// InvalidArgumentError struct definition
type InvalidArgumentError struct {
	ArgName  string
	ArgValue interface{}
}

// Error func definition
func (e APICallError) Error() string {
	return fmt.Sprintf("calling API Error - %s", e.Message)
}

// Error func definition
func (e UserValidationError) Error() string {
	return fmt.Sprintf("user verification failed - %s", e.Message)
}

// Error func definition
func (e NotFoundError) Error() string {
	return fmt.Sprintf("record not found - %s", e.Message)
}

// Error func definition
func (e InvalidArgumentError) Error() string {
	return fmt.Sprintf("Argument %v : %v is invalid", e.ArgName, e.ArgValue)
}
