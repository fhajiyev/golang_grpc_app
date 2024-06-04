package cookiehandler

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	uuid "github.com/satori/go.uuid"
)

// CookieHandler handles cookie-related actions
type CookieHandler struct {
	Ctx core.Context
}

const (
	// cookiePath const definition
	cookiePath = "/"
	// cookieDomain const definition
	cookieDomain = "buzzvil.com"
	// expireOffset const definition
	cookieIDKey = "cookie_id"
	// cookieChecksumKey const definition
	cookieChecksumKey = "checksum"
	// cookieVersionKey const definition
	cookieVersionKey = "cookie_version"
	// cookieVersion const definition
	cookieVersion = "3"
	// expireOffset const definition
	expireOffset = 365 * 24 * time.Hour
	// timeFormat const definition
	timeFormat = "Mon, 02 Jan 2006 15:00:00 GMT"
)

// SetCookieIDAndChecksum sets cookieID and checksum based on both cookieID and IFA into response and emits a log whenever cookie is updated
func (ch *CookieHandler) SetCookieIDAndChecksum(ifa string) (bool, string) {
	cookieUpdated := false
	expiration := time.Now().Add(expireOffset).UTC().Format(timeFormat)
	userAgent := ch.Ctx.Request().Header.Get("User-Agent")

	cookieID, exist := ch.getCookieValue(cookieIDKey)
	if !exist {
		cookieID = generateCookieID(ifa)
		ch.setCookie(cookieIDKey, cookieID, expiration)
		cookieUpdated = true
	}

	storedCookieVersion, exist := ch.getCookieValue(cookieVersionKey)
	if !exist || storedCookieVersion != cookieVersion {
		ch.setCookie(cookieVersionKey, cookieVersion, expiration)
		cookieUpdated = true
	}

	storedChecksum, exist := ch.getCookieValue(cookieChecksumKey)
	checksum := generateChecksum(cookieID + ifa)
	if !exist || storedChecksum != checksum {
		ch.setCookie(cookieChecksumKey, checksum, expiration)
		cookieUpdated = true
	}

	if cookieUpdated {
		profileLogJSON, err := json.Marshal(map[string]interface{}{
			"type": "profile",
			"payload": map[string]interface{}{
				"cookie_id":  cookieID,
				"ifa":        ifa,
				"user_agent": userAgent,
			},
		})
		if err != nil {
			core.Logger.Warnf("SetCookieIDAndChecksum() err: %s", err)
			return false, ""
		}

		fmt.Println(string(profileLogJSON))
	}

	return cookieUpdated, cookieID
}

// GetCookieValue returns true if cookieID exists in cookie
func (ch *CookieHandler) getCookieValue(key string) (string, bool) {
	cookie, err := ch.Ctx.Request().Cookie(key)
	if err != nil || cookie == nil {
		return "", false
	}

	return cookie.Value, true
}

// SetCookie sets cookieID into response
func (ch *CookieHandler) setCookie(key string, cookieVal string, expiration string) {
	path := cookiePath
	domain := cookieDomain
	cookieString := generateCookieHeaders(key, cookieVal, domain, expiration, path)
	ch.Ctx.Response().Header().Add("Set-Cookie", cookieString)
}

// generateCookieHeaders generates the cookie headers to be set under "set-cookie" key into response
func generateCookieHeaders(key string, value string, domain string, expiration string, path string) string {
	return fmt.Sprintf("%s=%s;Domain=%s;Expires=%s;Path=%s;SameSite=None;Secure", key, value, domain, expiration, path)
}

// generateCookieID generates cookie ID based on given seed value
func generateCookieID(seed string) string {
	return uuid.NewV5(uuid.UUID{}, seed).String()
}

// generateChecksum generates checksum based on given seed value
func generateChecksum(seed string) string {
	hasher := md5.New()
	hasher.Write([]byte(seed))
	return hex.EncodeToString(hasher.Sum(nil))
}
