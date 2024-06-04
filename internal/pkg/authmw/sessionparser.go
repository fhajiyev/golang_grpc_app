package authmw

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbdevice"

	"github.com/Buzzvil/buzzlib-go/crypto"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/auth"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
)

// keyValue value has been hardcoded in the file below.
// https://github.com/Buzzvil/buzzscreen-api/blob/master/buzzscreen/model/device.go#L98
var keyValue = []byte("BuzzscreenServer")

type sessionReq struct {
	SessionKey string `form:"session_key" query:"session_key" validate:"required"`
}

type session struct {
	appID           int64
	publisherUserID string
	deviceID        int64
}

// SessionParser creates function for parsing session key.
func SessionParser(db *gorm.DB) AccountParser {
	return func(c echo.Context) (*auth.Identifier, error) {
		req := new(sessionReq)
		if err := c.Bind(req); err != nil {
			return nil, common.NewBindError(err)
		}

		if err := c.Validate(req); err != nil {
			return nil, common.NewBindError(err)
		}

		s, err := decodeSessionKey(req.SessionKey)
		if err != nil {
			return nil, err
		}

		var d dbdevice.Device
		if err := db.Where(&dbdevice.Device{ID: s.deviceID}).First(&d).Error; err != nil {
			return nil, err
		}

		return &auth.Identifier{AppID: s.appID, PublisherUserID: s.publisherUserID, IFA: d.IFA}, nil
	}
}

func decodeSessionKey(sessionKey string) (session, error) {
	var s session
	decodedBytes, err := crypto.Base64Decoding(sessionKey)
	if err != nil {
		return s, err
	}

	decoded, err := crypto.AesDecrpyter(keyValue, decodedBytes)
	if err != nil {
		return s, err
	}

	parts := strings.Split(decoded, "&&")

	if len(parts) < 5 {
		return s, errors.New("Cannot parse session key")
	}

	if s.appID, err = strconv.ParseInt(parts[0], 10, 64); err != nil {
		return s, err
	}
	s.publisherUserID = parts[1]
	if s.deviceID, err = strconv.ParseInt(parts[2], 10, 64); err != nil {
		return s, err
	}

	return s, nil
}
