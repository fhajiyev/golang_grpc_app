package dto

import (
	"net/http"
)

type (
	// GetPrivacyPolicyRequest holds request parameters for Privacy Policy Request
	GetPrivacyPolicyRequest struct {
		AppID      int64  `form:"appId" query:"appId"`
		UnitIDV1   int64  `form:"unit_id" query:"unit_id"`
		CountryReq string `form:"country" query:"country"`

		country string
		Request *http.Request `form:"-"`
	}

	// GetPrivacyPolicyResponse is a response for Privacy Policy Request
	GetPrivacyPolicyResponse struct {
		IsRequired                bool                                 `json:"isRequired"`
		PrivacyPolicyTranslations *map[string]PrivacyPolicyTranslation `json:"translations,omitempty"`
	}

	// PrivacyPolicyTranslation is an object to hold Privacy Policy translation
	PrivacyPolicyTranslation struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Consent string `json:"consent"`
	}

	// PostPrivacyRequest holds request parameters for Post Privacy Request
	PostPrivacyRequest struct {
		IsConsented bool   `form:"isConsented" query:"isConsented" required:"true"`
		SessionKey  string `form:"sessionKey" query:"sessionKey"`
		DeviceID    int64  `form:"deviceId" query:"deviceId"`
	}
)

// GetCountry returns country in string
func (ppReq *GetPrivacyPolicyRequest) GetCountry() string {
	return ppReq.country
}
