package jwt_test

import (
	"testing"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/common/jwt"
)

func TestGetServiceToken(t *testing.T) {
	token := jwt.GetServiceToken()
	if token == "" {
		t.Fatalf("Expected to get a service token, got: %s.", token)
	}
}
