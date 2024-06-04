package ad

// BAUserRepository interface definition
type BAUserRepository interface {
	GetBAUserByID(id int64) (*BAUser, error)
}

// Repository interface definition
type Repository interface {
	GetAdDetail(adID int64, accessToken string) (*Detail, error)
	GetAdsV1(v1Req V1AdsRequest) (*V1AdsResponse, error)
	GetAdsV2(v2Req V2AdsRequest) (*V2AdsResponse, error)
}
