package authmw

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/suite"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/header"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/auth"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/mock"
)

func TestAuthMiddleware(t *testing.T) {
	suite.Run(t, new(AuthMiddlewareTestSuite))
}

const (
	headerKeyAppID           = "Buzz-App-ID"
	headerKeyPublisherUserID = "Buzz-Publisher-User-ID"
	headerKeyAccountID       = "Buzz-Account-ID"
	headerKeyIFA             = "Buzz-Ifa"
)

func (ts *AuthMiddlewareTestSuite) TestRequestLogIn() {
	token := ts.createToken()
	rawToken := ts.buildRawToken(token)
	a := ts.createAuth()
	identifier := ts.buildIdentifier(a)
	ctx := ts.createContext()

	parser := func(echo.Context) (*auth.Identifier, error) {
		return &identifier, nil
	}
	ah := accountHandler{uc: ts.uc, ap: parser}

	ts.uc.On("CreateAuth", identifier).Return(rawToken, nil).Once()

	err := ah.RequestLogIn(ts.next)(ctx)
	ts.NoError(err)

	ctxToken := ctx.Request().Header.Get(headerKeyAuthorization)
	ts.Equal(token, ctxToken)
}

func (ts *AuthMiddlewareTestSuite) TestAuthenticate() {
	token := ts.createToken()
	rawToken := ts.buildRawToken(token)
	a := ts.createAuth()
	ctx := ts.createContext()

	ah := authHandler{ts.uc}
	ts.uc.On("GetAuth", rawToken).Return(&a, nil).Once()

	ctx.Request().Header.Set(headerKeyAuthorization, token)

	err := ah.Authenticate(ts.next)(ctx)
	ts.NoError(err)
	ts.assertHeaderAuth(a, ctx.Request().Header)
}

func (ts *AuthMiddlewareTestSuite) TestAppendAuthToHeader() {
	token := ts.createToken()
	rawToken := ts.buildRawToken(token)
	a := ts.createAuth()
	ctx := ts.createContext()

	ah := authHandler{ts.uc}
	ts.uc.On("GetAuth", rawToken).Return(&a, nil).Once()

	ctx.Request().Header.Set(headerKeyAuthorization, token)

	err := ah.AppendAuthToHeader(ts.next)(ctx)
	ts.NoError(err)
	ts.assertHeaderAuth(a, ctx.Request().Header)
}

func (ts *AuthMiddlewareTestSuite) TestAppendAuthToHeader_NoToken() {
	ctx := ts.createContext()

	ah := authHandler{ts.uc}

	err := ah.AppendAuthToHeader(ts.next)(ctx)
	ts.NoError(err)
	ts.assertHeaderNoAuth(ctx.Request().Header)
}

func (ts *AuthMiddlewareTestSuite) assertHeaderNoAuth(header http.Header) {
	ts.Equal("", header.Get(headerKeyAccountID))
	ts.Equal("", header.Get(headerKeyAppID))
	ts.Equal("", header.Get(headerKeyPublisherUserID))
	ts.Equal("", header.Get(headerKeyIFA))
}

func (ts *AuthMiddlewareTestSuite) assertHeaderAuth(a auth.Auth, header http.Header) {
	ts.Equal(strconv.FormatInt(a.AccountID, 10), header.Get(headerKeyAccountID))
	ts.Equal(strconv.FormatInt(a.AppID, 10), header.Get(headerKeyAppID))
	ts.Equal(a.PublisherUserID, header.Get(headerKeyPublisherUserID))
	ts.Equal(a.IFA, header.Get(headerKeyIFA))
}

func (ts *AuthMiddlewareTestSuite) createContext() core.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	return e.NewContext(req, nil)
}

func (ts *AuthMiddlewareTestSuite) buildRawToken(token string) string {
	return strings.TrimPrefix(token, bearer)
}

func (ts *AuthMiddlewareTestSuite) createToken() string {
	var s string
	ts.NoError(faker.FakeData(&s))
	return bearer + s
}

func (ts *AuthMiddlewareTestSuite) buildIdentifier(a auth.Auth) auth.Identifier {
	return auth.Identifier{
		AppID:           a.AppID,
		PublisherUserID: a.PublisherUserID,
		IFA:             a.IFA,
	}
}

func (ts *AuthMiddlewareTestSuite) buildHeaderAuth(a auth.Auth) header.Auth {
	return header.Auth{
		AccountID:       a.AccountID,
		AppID:           a.AppID,
		PublisherUserID: a.PublisherUserID,
		IFA:             a.IFA,
	}
}

func (ts *AuthMiddlewareTestSuite) createAuth() auth.Auth {
	a := auth.Auth{}
	ts.NoError(faker.FakeData(&a))
	return a
}

type AuthMiddlewareTestSuite struct {
	suite.Suite
	next func(echo.Context) error
	uc   *mockUseCase
}

func (ts *AuthMiddlewareTestSuite) SetupTest() {
	ts.next = func(echo.Context) error { return nil }
	ts.uc = &mockUseCase{}
}

func (ts *AuthMiddlewareTestSuite) AfterTest(_, _ string) {
	ts.uc.AssertExpectations(ts.T())
}

type mockUseCase struct {
	mock.Mock
}

func (u *mockUseCase) CreateAuth(identifier auth.Identifier) (string, error) {
	args := u.Called(identifier)
	return args.Get(0).(string), args.Error(1)
}

func (u *mockUseCase) GetAuth(token string) (*auth.Auth, error) {
	args := u.Called(token)
	return args.Get(0).(*auth.Auth), args.Error(1)
}
