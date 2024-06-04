package payload

const (
	payloadAESKey  = "ejuvbej14828vjdp"
	payloadURLSafe = true
)

// Payload struct definition
type Payload struct {
	Country     string  `json:"cou"`
	EndedAt     int64   `json:"ended_at"`
	Gender      *string `json:"sex,omitempty"`
	OrgID       int64   `json:"oid"`
	Time        int64   `json:"t"`
	Timezone    string  `json:"tz"`
	YearOfBirth *int    `json:"yob,omitempty"`

	UnitID *int64 `json:"unit_id,omitempty"`
}

var acceptedUnitIDs = map[int64]struct{}{
	100000043:       {}, // HSKR
	100000045:       {}, // HSJP
	100000050:       {}, // HSTW
	210342277740215: {}, // SJ
}

// SetUnitID assigns unitID if acceptable conditions
/*
HS/SJ sdk_version 3500,3600 에서 app_id, unit_id를 request param에 넣어주지 않는 문제. HS-704
user-agent에 있는 package_name과 sdk_version으로 app_id를 복구함
*/
// TODO HS/SJ 앱에서 sdkversion 3500/3600을 사용하지 않을때 코드 제거
func (p *Payload) SetUnitID(unitID int64, sdkVersion int) {
	_, acceptUnitID := acceptedUnitIDs[unitID]
	if acceptUnitID && (sdkVersion == 3500 || sdkVersion == 3600) {
		p.UnitID = &unitID
	}
}
