package ifa

import "fmt"

const (
	// EmptyIFA defines empty ifa caused by iOS
	EmptyIFA = "00000000-0000-0000-0000-000000000000"
)

// ShouldReplaceIFAWithIFV func definition
// https://buzzvil.atlassian.net/browse/PO-649
func ShouldReplaceIFAWithIFV(IFA string, IFV *string) bool {
	return IFV != nil && (IFA == "" || IFA == EmptyIFA) && *IFV != ""
}

// GetDeviceIFV func definition
func GetDeviceIFV(IFV string) string {
	return fmt.Sprintf("ifv%v", IFV)
}
