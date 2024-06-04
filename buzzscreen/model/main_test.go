package model_test

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Buzzvil/buzzscreen-api/tests"
)

var ts *httptest.Server

func TestMain(m *testing.M) {
	ts = tests.GetTestServer(m)
	tearDownDatabase := tests.SetupDatabase()

	code := m.Run()

	tearDownDatabase()
	tests.DeleteLogFiles()
	ts.Close()

	os.Exit(code)
}
