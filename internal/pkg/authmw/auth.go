package authmw

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/header"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/auth"
	"github.com/labstack/echo"
)

// AccountParser function parses echo.Context to extract auth.Account data.
type AccountParser func(echo.Context) (*auth.Identifier, error)

const (
	// ContextKeyAuth is a key for stored `auth.Auth` value in context.
	contextKeyAuth = "authmw.auth"

	// headerKeyAuthorization is a key for stored `token` value in header.
	headerKeyAuthorization = "authorization"
	bearer                 = "Bearer "
)

type accountHandler struct {
	ap AccountParser
	uc auth.UseCase
}

type authHandler struct {
	uc auth.UseCase
}

// RequestLogIn creates new echo.MiddlewareFunc for creating authentication.
func RequestLogIn(authUseCase auth.UseCase, ap AccountParser) echo.MiddlewareFunc {
	ah := &accountHandler{ap: ap, uc: authUseCase}
	return ah.RequestLogIn
}

// RequestLogIn func definition
func (ah *accountHandler) RequestLogIn(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		acc, err := ah.ap(c)
		if err != nil {
			return err
		}

		token, err := ah.uc.CreateAuth(*acc)
		if err != nil {
			return &core.HttpError{Code: http.StatusInternalServerError, Message: err.Error()}
		}

		ah.setupToken(c, token)

		return next(c)
	}
}

func (ah *accountHandler) setupToken(c echo.Context, token string) {
	c.Request().Header.Set(headerKeyAuthorization, bearer+token)
}

// Authenticate creates new echo.MiddlewareFunc to authenticate request
// TODO replace to envoy ext_authz
func Authenticate(authUseCase auth.UseCase) echo.MiddlewareFunc {
	ah := &authHandler{uc: authUseCase}
	return ah.Authenticate
}

// AppendAuthToHeader creates new echo.MiddlewareFunc to fill auth from request
func AppendAuthToHeader(authUseCase auth.UseCase) echo.MiddlewareFunc {
	ah := &authHandler{uc: authUseCase}
	return ah.AppendAuthToHeader
}

// Authenticate func definition
func (ah *authHandler) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Get(contextKeyAuth) != nil {
			return next(c)
		}

		token, err := ah.getTokenFromheader(c.Request().Header)
		if err != nil {
			return &core.HttpError{Code: http.StatusUnauthorized, Message: err.Error()}
		}

		a, err := ah.uc.GetAuth(token)
		if err != nil {
			return &core.HttpError{Code: http.StatusUnauthorized, Message: err.Error()}
		}

		ah.setupAuth(c, *a)
		return next(c)
	}
}

// AppendAuthToHeader func definition
func (ah *authHandler) AppendAuthToHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Get(contextKeyAuth) != nil {
			return next(c)
		}

		token, err := ah.getTokenFromheader(c.Request().Header)
		if err != nil {
			return next(c)
		}

		a, err := ah.uc.GetAuth(token)
		if err != nil {
			return next(c)
		}

		ah.setupAuth(c, *a)
		return next(c)
	}
}

func (ah *authHandler) getTokenFromheader(header http.Header) (string, error) {
	authorizationValue := header.Get(headerKeyAuthorization)
	if authorizationValue == "" {
		return "", errors.New("no authorization")
	}

	if !strings.HasPrefix(authorizationValue, bearer) {
		return "", errors.New("authorization does not have prefix")
	}

	return strings.TrimPrefix(authorizationValue, bearer), nil
}

func (ah *authHandler) setupAuth(c echo.Context, a auth.Auth) {
	req := c.Request()
	req.Header = header.AppendAuthToHeader(req.Header, header.Auth{
		AccountID:       a.AccountID,
		AppID:           a.AppID,
		PublisherUserID: a.PublisherUserID,
		IFA:             a.IFA,
	})
	c.Set(contextKeyAuth, a)
}
