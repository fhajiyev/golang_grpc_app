package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBindHeaders(t *testing.T) {
	type TestModel struct {
		TestKey string `header:"Test-Header-Key"`
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Test-Header-Key", "test-header-value")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var tm TestModel
	err := bindHeaders(c, &tm)
	require.Nil(t, err, err)
	assert.Equal(t, "test-header-value", tm.TestKey)
}
