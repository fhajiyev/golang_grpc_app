package dto

import (
	"time"
)

// PromotionType type definition
type PromotionType string

//PromotionType const definition
const (
	WelcomePromotion  PromotionType = "welcome"
	ReferralPromotion PromotionType = "referral"
)

// Promotion type definition
type Promotion struct {
	ID        int64         `json:"-"`
	Type      PromotionType `json:"type"`
	Amount    int           `json:"amount"`
	EndTime   *time.Time    `json:"end_time,omitempty"`
	StartTime *time.Time    `json:"start_time,omitempty"`
}

// GetCampaignsRes type definition
type GetCampaignsRes struct {
	Promotions []*Promotion `json:"promotions"`
}
