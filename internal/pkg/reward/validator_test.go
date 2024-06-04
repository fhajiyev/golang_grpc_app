package reward

import (
	"fmt"
	"testing"

	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/suite"
)

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(validatorTestSuite))
}

type validatorTestSuite struct {
	suite.Suite
	v *validator
}

func (ts *validatorTestSuite) SetupSuite() {
}

func (ts *validatorTestSuite) SetupTest() {
	ts.v = &validator{}
	ts.NotNil(ts.v)
}

func (ts *validatorTestSuite) TestValidateCheckSum() {
	ts.Run("test on valid ifa", func() {
		req := RequestIngredients{}
		faker.FakeData(&req)

		expected := GetMD5Hash(fmt.Sprintf("buz:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v",
			req.DeviceID, req.IFA, req.UnitDeviceToken, req.AppID,
			req.CampaignID, req.CampaignType, req.CampaignName, *req.CampaignOwnerID, req.CampaignIsMedia,
			req.Slot, req.Reward, req.BaseReward))

		req.Checksum = expected

		ts.Require().True(ts.v.validateChecksum(req))
	})
}
