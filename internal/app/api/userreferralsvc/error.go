package userreferralsvc

import "fmt"

// DeviceValidationError will be returned when it is old or invalid device
type DeviceValidationError struct {
	Message string
}

// NotFoundError will be returned when each usecase returs nil with no error
type NotFoundError struct {
	Message string
}

// Error func definition
func (e DeviceValidationError) Error() string {
	return fmt.Sprintf("device validaiton failed - %s", e.Message)
}

// Error func definition
func (e NotFoundError) Error() string {
	return fmt.Sprintf("%s not found", e.Message)
}
