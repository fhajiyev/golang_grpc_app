package session_test

import (
	"testing"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/session"
	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/suite"
)

func (ts *UseCaseTestSuite) Test_Session() {
	var err error
	var s *session.Session
	err = faker.FakeData(&s)
	ts.NoError(err)

	sessionKey := ts.u.GetNewSessionKey(s.AppID, s.UserID, s.DeviceID, s.AndroidID, s.CreatedSeconds)
	resultSession, err := ts.u.GetSessionFromKey(sessionKey)

	ts.NoError(err)
	ts.Equal(s.AppID, resultSession.AppID)
	ts.Equal(s.UserID, resultSession.UserID)
	ts.Equal(s.DeviceID, resultSession.DeviceID)
	ts.Equal(s.AndroidID, resultSession.AndroidID)
	ts.Equal(s.CreatedSeconds, resultSession.CreatedSeconds)
}

var (
	_ suite.SetupTestSuite = &UseCaseTestSuite{}
)

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseTestSuite))
}

type UseCaseTestSuite struct {
	suite.Suite
	u session.UseCase
}

func (ts *UseCaseTestSuite) SetupTest() {
	ts.u = session.NewUseCase()
}
