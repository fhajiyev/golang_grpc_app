package model

import (
	"time"
)

// StatusType type definition
type StatusType int

//noinspection GoUnusedConst
const (
	StatusUnknown   StatusType = iota
	StatusPending   StatusType = 1
	StatusProcessed StatusType = 2
	StatusCompleted StatusType = 3
)

type (
	// WelcomeReward type definition
	WelcomeReward struct {
		ID              int64 `gorm:"primary_key"`
		DeviceID        int64
		Amount          int
		UnitDeviceToken string
		UnitID          int64 `gorm:"column:unit_id"`
		Status          StatusType
		Version         *int
		ConfigID        int64 `gorm:"column:config_id"`

		CreatedAt *time.Time `json:"-"`
		UpdatedAt *time.Time `json:"-"`
	}
)
