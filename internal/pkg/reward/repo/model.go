package repo

// DBPoint type definition
// TODO: make it private
type DBPoint struct {
	DeviceID        int64  `dynamo:"did,hash"`
	Version         int64  `dynamo:"v,range"`
	UnitID          int64  `dynamo:"ui"`
	UnitDeviceToken string `dynamo:"udt"`
	Requested       bool   `dynamo:"r"`
	Type            string `dynamo:"pt"`
	Title           string `dynamo:"ti"`
	SubType         string `dynamo:"st"`
	ReferKey        string `dynamo:"rk"` // CampaignID
	Amount          int64  `dynamo:"am"`
	BaseReward      int    `dynamo:"br"`
	DepositSum      int64  `dynamo:"ds"`
	CreatedAt       int64  `dynamo:"ca"`
	UpdatedAt       int64  `dynamo:"ua"`
	Slot            int64  `dynamo:"slot"`
	Scatter         int64  `dynamo:"sc"`
}

const (
	keyDeviceID = "did"
)

const (
	pointTypeImpression = "imp"
)
