package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_IsRegisterSecondsRewardableNow(t *testing.T) {
	now := time.Now()
	startTime := now.Add(-24 * time.Hour)
	endTime := now.Add(time.Hour)
	wrc := WelcomeRewardConfig{
		RetentionDays: 0,
		Amount:        10,
		StartTime:     &startTime,
		EndTime:       &endTime,
	}

	assert.Equal(t, wrc.IsRegisterSecondsRewardable(now.Unix()), true)
}

func Test_IsRegisterSecondsRewardableNow_InfiniteEnddate(t *testing.T) {
	now := time.Now()
	twoDaysBefore, threeDaysBefore := now.Add(-48*time.Hour), now.Add(-72*time.Hour)

	wrc := WelcomeRewardConfig{
		RetentionDays: 2,
		Amount:        10,
		StartTime:     &threeDaysBefore,
		EndTime:       nil,
	}
	assert.Equal(t, wrc.IsRegisterSecondsRewardable(twoDaysBefore.Unix()), true)
}

// 캠페인이 EndTime 을 지났어도 RetentionDays 옵션이 있을 경우 해당 기간 만큼 더 허용해줘야 한다.
func Test_IsRegisterSecondsRewardableNow_RetentionDays(t *testing.T) {
	now := time.Now()
	oneDayBefore, twoDaysBefore, threeDaysBefore := now.Add(-24*time.Hour), now.Add(-48*time.Hour), now.Add(-72*time.Hour)

	wrc := WelcomeRewardConfig{
		RetentionDays: 2,
		Amount:        10,
		StartTime:     &threeDaysBefore,
		EndTime:       &oneDayBefore,
	}

	assert.Equal(t, wrc.IsRegisterSecondsRewardable(twoDaysBefore.Unix()), true)
}

