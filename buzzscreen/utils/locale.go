package utils

import (
	"regexp"
	"strings"

	"github.com/Buzzvil/buzzlib-go/core"
)

// SplitLocale func definition
func SplitLocale(localeString string) (string, string) {
	parts := regexp.MustCompile("[-_, ]").Split(localeString, 3)
	switch len(parts) {
	case 1:
		return strings.ToLower(parts[0]), ""
	case 3:
		if len(parts[1]) == 0 {
			core.Logger.Warnf("SplitLocale() local has scripts or extensions without country")
			break
		}
		parts[0] = strings.ToLower(parts[0]) + "_" + strings.ToUpper(string(parts[1][0])) + strings.ToLower(parts[1][1:])
		parts[1] = strings.ToUpper(parts[2])
		return parts[0], parts[1]
	}
	return strings.ToLower(parts[0]), strings.ToUpper(parts[1])
}
