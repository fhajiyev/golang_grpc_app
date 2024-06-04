package model

const (
	// RAW const definition
	RAW = "R"

	// StaffOrganizationID const definition
	StaffOrganizationID = 1

	// DisplayWeightMultiplier const definition
	DisplayWeightMultiplier = 10000 // settings

	// CampaignTypeCast const definition
	CampaignTypeCast = "C"

	// CleanModeForceDisabled const definition
	CleanModeForceDisabled = 0 // campaign.models

	// DefaultBaseReward const definition
	DefaultBaseReward = 2

	// APIAesKey const definition
	APIAesKey = "4i3jfkdie0998754"
	// APIAesIv const definition
	APIAesIv = APIAesKey

	// ESGlobString const definition
	ESGlobString = "__GLOB__"
	// ESNullShortMin const definition
	ESNullShortMin = -32768
	// ESNullShortMax const definition
	ESNullShortMax = 32767

	// ESNullIntMin const definition
	ESNullIntMin = -2147483648
	// ESNullIntMax const definition
	ESNullIntMax = 2147483647
)

// Status type definition
type Status int

// Status constants
const (
	_                      = iota
	StatusManual    Status = 1
	StatusCuratable Status = 2
	StatusSelected  Status = 3
	StatusAccepted  Status = 4
	StatusRejected  Status = 5
	StatusComplete  Status = 6
	StatusFeedOnly  Status = 7
)

// StatusesForLockscreen type definition
var StatusesForLockscreen = []Status{StatusManual, StatusComplete}

// StatusesForFeed type definition
var StatusesForFeed = []Status{StatusCuratable, StatusSelected, StatusAccepted, StatusFeedOnly, StatusComplete}

// LandingType type definition
type LandingType int

// LandingTypeBrowser constants
const (
	_                              = iota
	LandingTypeBrowser LandingType = 1
	LandingTypeOverlay LandingType = 2
	LandingTypeCard    LandingType = 3
	LandingTypeYoutube LandingType = 8 | LandingTypeOverlay
)

// Carrier type definition
type Carrier string

const (
	// Japan
	au       Carrier = "au"
	softbank Carrier = "softbank"
	docomo   Carrier = "docomo"

	// Korea
	kt  Carrier = "kt"
	skt Carrier = "skt"
	lgt Carrier = "lgt"

	// Taiwan
	twMobile        Carrier = "twm"
	chungwhaTelecom Carrier = "cht"
	apTelecom       Carrier = "apt"
	farEasTone      Carrier = "fet"
	tStar           Carrier = "tst"

	inAirtel   Carrier = "airtel"
	inVodafone Carrier = "vodafone"
	inIdea     Carrier = "idea"
)

// CarrierMap var definition
var CarrierMap = map[string]Carrier{
	// Korea
	"SKTelecom":      skt,
	"olleh":          kt,
	"LG U+":          lgt,
	"KT":             kt,
	"SK Telecom":     skt,
	"LG UPLUS":       lgt,
	"SKT":            skt,
	"KOR SK Telecom": skt,
	"KTF":            kt,
	"3G olleh":       kt,
	"KR KTF":         kt,
	"LGU+":           lgt,

	// Japan
	"KDDI":       au,
	"NTT DOCOMO": docomo,
	"ソフトバンクモバイル": softbank,
	"ドコモ":        docomo,
	"SoftBank":   softbank,
	"DOCOMO":     docomo,
	"JP DOCOMO":  docomo,

	// Taiwan

	// Taiwan - Taiwan mobile
	"台灣大哥大":                 twMobile,
	"台灣大哥大,null":            twMobile,
	"TW Mobile":             twMobile,
	"Taiwan Mobile Co. Ltd": twMobile,
	"台湾大哥大":                 twMobile,
	"Taiwan Mobile":         twMobile,
	"TWM":                   twMobile,
	"myfone":                twMobile,
	"Taiwanmobile":          twMobile,

	// Taiwan - Chunghwa telecom
	"中華電信":                 chungwhaTelecom,
	"Chunghwa Telecom":     chungwhaTelecom,
	"Chungwa":              chungwhaTelecom,
	"Chunghwa":             chungwhaTelecom,
	"中華電信,null":            chungwhaTelecom,
	"中華電信,中華電信":            chungwhaTelecom,
	"CHT":                  chungwhaTelecom,
	"null,中華電信":            chungwhaTelecom,
	"Chunghw":              chungwhaTelecom,
	"Chunghwa Telecom LDM": chungwhaTelecom,
	"中华电信":                 chungwhaTelecom,

	// Taiwan - Fas EasTone
	"遠傳電信":                             farEasTone,
	"Far EasTone":                      farEasTone,
	"FarEasTone":                       farEasTone,
	"Far EasTone Telecommunications C": farEasTone,
	"遠傳":                               farEasTone,
	"远传电信":                             farEasTone,
	"遠傳電信,null":                        farEasTone,
	"遠傳電信 3G":                          farEasTone,
	"Far EasTone,null":                 farEasTone,
	"FarEasTon":                        farEasTone,
	"null,遠傳電信":                        farEasTone,
	"FET":                              farEasTone,

	// Taiwan - T Star
	"台灣之星":                tStar,
	"台灣之星電信股份有限公司":        tStar,
	"Taiwan Star Telecom": tStar,
	"T Star":              tStar,
	"TStar":               tStar,
	"臺灣之星":                tStar,
	"威寶電信":                tStar,
	"VIBO Telecom Inc":    tStar,
	"VIBO Telecom":        tStar,
	"威寶電信 3G":             tStar,
	"VIBO":                tStar,
	"威宝电信":                tStar,

	// Taiwan - Asia Pacific Telecom
	"亞太電信":                 apTelecom,
	"Asia Pacific Telecom": apTelecom,
	"GT 4G":                apTelecom,
	"亞太行動":                 apTelecom,
	"亞太電信,null":            apTelecom,

	// India - Air tel
	"airtel":             inAirtel,
	"AirTel":             inAirtel,
	"Airtel Assam":       inAirtel,
	"Airtel A.P.":        inAirtel,
	"Airtel Mumbai":      inAirtel,
	"Airtel | airtel":    inAirtel,
	"IND Airtel":         inAirtel,
	"Airtel":             inAirtel,
	"Airtel Maharashtra": inAirtel,
	"AIRTEL | airtel":    inAirtel,
	"Airtel Delhi":       inAirtel,
	"airtel | Airtel":    inAirtel,
	"Airtel Orissa":      inAirtel,
	"IND AirTel":         inAirtel,
	"Airtel UP East":     inAirtel,

	// India - Vodafone
	"Vodafone IN":               inVodafone,
	"Vodafone In":               inVodafone,
	"VODAFONE | Vodafone IN":    inVodafone,
	"VODAFONEIN":                inVodafone,
	"Vodafone":                  inVodafone,
	"Supernet Vodafone":         inVodafone,
	"vodafone":                  inVodafone,
	"VODAFONE IN | Vodafone IN": inVodafone,
	"VODAFONE":                  inVodafone,
	"VODAFONE IN":               inVodafone,
	"VODAFONEIN | Vodafone IN":  inVodafone,
	"Vodafone Punjab":           inVodafone,

	// India - idea
	"Jio 4G,IDea":    inIdea,
	"IDea":           inIdea,
	"IDea 2G":        inIdea,
	"IDEA IND":       inIdea,
	"IDEA | IDea":    inIdea,
	"IDEA A.P.":      inIdea,
	"IDea Karnataka": inIdea,
	"IDEA U.P.(E)":   inIdea,
	"IDEA MP":        inIdea,
	"IDEA":           inIdea,
	"IDEA U.P.(W)":   inIdea,
}
