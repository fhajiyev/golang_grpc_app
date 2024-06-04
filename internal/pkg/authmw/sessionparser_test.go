package authmw_test

import (
	"net/url"
	"regexp"
	"testing"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/authmw"
	"github.com/jinzhu/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v2"
)

func TestSessionParser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.Nil(t, err)
	gormDB, err := gorm.Open("mysql", db)
	require.Nil(t, err)

	expectedAppID := int64(100001)
	expectedPublisherUserID := "TestPublisherUserID"
	expectedIFA := "TestIFA"
	deviceID := int64(101)
	// Encrypted token for "100001&&TestPublisherUserID&&101&&TestAndroidID&&1557305050".
	token := "yWWJcFSgrQYPRpqFuhy7cKVm7u+DVb/ZJcvG1n7XDcDbIXRs6/WEStlI5hYxS2QgpdWdkk5EyJprN1dGSUrCGOsFnmoFO/shKVU3"
	rows := sqlmock.NewRows([]string{"ifa"}).AddRow(expectedIFA)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `device` WHERE (`device`.`id` = ?) ORDER BY `device`.`id` ASC LIMIT 1")).WithArgs(deviceID).WillReturnRows(rows)

	f := make(url.Values)
	f.Set("session_key", token)
	c := newApplicationFormContext(f)

	parser := authmw.SessionParser(gormDB)
	identifier, err := parser(c)
	require.Nil(t, err, err)
	require.NotNil(t, identifier)

	assert.Equal(t, expectedAppID, identifier.AppID)
	assert.Equal(t, expectedPublisherUserID, identifier.PublisherUserID)
	assert.Equal(t, expectedIFA, identifier.IFA)
	err = mock.ExpectationsWereMet()
	assert.Nil(t, err, err)
}

func TestSessionParserNoSessionKey(t *testing.T) {
	f := make(url.Values)
	f.Set("dummy", "value")
	c := newApplicationFormContext(f)

	parser := authmw.SessionParser(nil)
	a, err := parser(c)
	require.NotNil(t, err, err)
	require.Nil(t, a)

	assert.Contains(t, err.Error(), "Field validation for 'SessionKey' failed on the 'required' tag")
}
