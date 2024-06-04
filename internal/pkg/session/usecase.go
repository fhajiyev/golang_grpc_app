package session

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Buzzvil/buzzlib-go/crypto"
)

// UseCase interface definition
type UseCase interface {
	GetNewSessionKey(appID int64, userID string, deviceID int64, androidID string, createdMillis int64) string
	GetSessionFromKey(sessionKey string) (*Session, error)
	NewAnonymousSession(appID int64) *Session
}

type useCase struct {
}

// GetNewSessionKey func definition
func (u *useCase) GetNewSessionKey(appID int64, userID string, deviceID int64, androidID string, createdMillis int64) string {
	result, err := crypto.AesEncrypter(keyValue, fmt.Sprintf("%v&&%v&&%v&&%v&&%v", appID, userID, deviceID, androidID, createdMillis))
	if err != nil {
		return ""
	}

	return crypto.Base64Encoding(result)
}

// GetSessionFromKey func definition
func (u *useCase) GetSessionFromKey(sessionKey string) (*Session, error) {
	sessionInfo := new(Session)
	decodedBytes, err := crypto.Base64Decoding(sessionKey)
	if err != nil {
		return nil, err
	}

	decoded, err := crypto.AesDecrpyter(keyValue, decodedBytes)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(decoded, "&&")
	if len(parts) < 5 {
		err = errors.New("cannot parse sessionKey")
		return nil, err
	}

	appID, _ := strconv.ParseInt(parts[0], 10, 64)
	deviceID, _ := strconv.ParseInt(parts[2], 10, 64)
	createdSeconds, _ := strconv.ParseInt(parts[4], 10, 64)
	sessionInfo.AppID = appID
	sessionInfo.UserID = parts[1]
	sessionInfo.DeviceID = deviceID
	sessionInfo.AndroidID = parts[3]
	sessionInfo.CreatedSeconds = createdSeconds
	return sessionInfo, nil
}

func (u *useCase) NewAnonymousSession(appID int64) *Session {
	return &Session{AppID: appID}
}

// NewUseCase returns UseCase interface
func NewUseCase() UseCase {
	return &useCase{}
}
