package authmw_test

import (
	"net/url"
	"testing"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/authmw"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIFAParserWithNoIFAValue(t *testing.T) {
	f := make(url.Values)
	f.Set("app_id", "1000001")
	c := newApplicationFormContext(f)

	parser := authmw.IFAParser
	acc, err := parser(c)
	require.Nil(t, acc)
	require.NotNil(t, err)

	assert.Contains(t, err.Error(), "Field validation for 'IFA' failed on the 'required' tag")
}

func TestIFAParserWithNoBothAppIDAndUnitID(t *testing.T) {
	f := make(url.Values)
	f.Set("ifa", "TESTIFAVALUE")
	c := newApplicationFormContext(f)

	parser := authmw.IFAParser
	acc, err := parser(c)
	require.Nil(t, acc)
	require.NotNil(t, err)

	assert.Contains(t, err.Error(), "`app_id` is required")
}

func TestIFAParserWithValidData(t *testing.T) {
	f := make(url.Values)
	f.Set("ifa", "TESTIFAVALUE")
	f.Set("app_id", "1000001")
	c := newApplicationFormContext(f)

	parser := authmw.IFAParser
	acc, err := parser(c)
	require.Nil(t, err)
	require.NotNil(t, acc)

	assert.Equal(t, "TESTIFAVALUE", acc.IFA)
	assert.Equal(t, int64(1000001), acc.AppID)
}
