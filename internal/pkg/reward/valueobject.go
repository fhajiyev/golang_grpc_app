package reward

// ClickType definition
type ClickType string

// ClickType const definition
const (
	ClickTypeLanding ClickType = "l"
	ClickTypeUnlock  ClickType = "u"
)

// ReceivedStatus definition
type ReceivedStatus string

// ReceivedStatus const definition
const (
	StatusUnknown  ReceivedStatus = "unknown"
	StatusReceived ReceivedStatus = "received"
)

// Period represents reward period as seconds
type Period int64

// Hours Returns reward period as hours with ignoring residuals
func (rp Period) Hours() int64 {
	if rp%3600 != 0 {
		return int64(rp)/3600 + 1
	}
	return int64(rp) / 3600
}

// PeriodForCampaign represents map of Period for campaignID
type PeriodForCampaign map[int64]Period

// MaxPeriod returns the maximum value in Periods
func (pfc PeriodForCampaign) MaxPeriod() Period {
	max := Period(0)
	for _, p := range pfc {
		if max < p {
			max = p
		}
	}
	return max
}

const (
	pointTypeImpression = "imp"
)
