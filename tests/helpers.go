package tests

import (
	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/route"

	"net/http/httptest"
	"os"
	"testing"

	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
)

// GetTestServer func definition
func GetTestServer(m *testing.M) *httptest.Server {
	bsSvc := buzzscreen.New()
	server := core.NewServer(bsSvc)
	configPath := os.Getenv("GOPATH") + "/src/github.com/Buzzvil/buzzscreen-api/config/"
	server.Init(configPath, &env.Config)
	route.InitLegacyRoute(server.Engine())
	return httptest.NewServer(server.Engine())
}

// DeleteLogFiles func definition
func DeleteLogFiles() {
	for _, logger := range env.Config.Loggers {
		os.Remove(logger.File)
	}
}
