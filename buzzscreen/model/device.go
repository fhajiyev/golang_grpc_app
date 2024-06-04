package model

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/Buzzvil/buzzscreen-api/buzzscreen/utils"
)

type (
	// DeviceContentConfig type definition
	DeviceContentConfig struct {
		ID        int64     `gorm:"primary_key" json:"id"`
		DeviceID  int64     `json:"device_id"`
		Category  string    `json:"category"`
		Channel   string    `json:"channel"`
		Campaign  string    `json:"campaign"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	// DeviceContentConfigChannel type definition
	DeviceContentConfigChannel struct {
		FollowStr   string `json:"follow"`
		UnfollowStr string `json:"unfollow"`
		followSlice []int
	}

	// DeviceUser type definition
	DeviceUser struct {
		ID         int64     `gorm:"primary_key" json:"id"`
		DeviceID   int64     `json:"device_id"`
		Code       string    `json:"code"`
		ReferrerID int64     `json:"referrer_id"`
		IsVerified bool      `json:"is_verified"`
		CreatedAt  time.Time `json:"-"`
		UpdatedAt  time.Time `json:"-"`
	}
)

// GetFollowings func definition
func (dccc *DeviceContentConfigChannel) GetFollowings() ([]int, error) {
	if dccc.FollowStr != "" && len(dccc.followSlice) == 0 {
		strSlice := strings.Split(strings.Trim(dccc.FollowStr, ","), ",")
		cIDs, err := utils.SliceAtoi(strSlice)
		if err != nil {
			return nil, err
		}
		dccc.followSlice = cIDs
	}
	return dccc.followSlice, nil
}

// GetJSONStr func definition
func (dcc *DeviceContentConfig) GetJSONStr() string {
	actJSON, err := json.Marshal(dcc)
	if err != nil {
		panic(err)
	}
	return string(actJSON)
}

// TableName func definition
func (DeviceUser) TableName() string {
	return "device_user"
}
