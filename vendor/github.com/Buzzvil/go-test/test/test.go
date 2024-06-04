package test

import (
	"fmt"
	"runtime/debug"
	"testing"
)

// AssertEqual check a and b are same.
// If they are different, this will call Fatal error with message.
func AssertEqual(t *testing.T, a interface{}, b interface{}, message string) {
	if a == b {
		return
	}
	message += fmt.Sprintf("\n%v != %v\n%s", a, b, debug.Stack())
	t.Fatal(message)
}
