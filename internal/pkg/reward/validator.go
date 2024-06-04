package reward

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/Buzzvil/buzzlib-go/core"
)

// TestCheckSum for testing
const (
	TestCheckSum = "awefe-effffv-ce2-34-5-fbn"
)

type validator struct {
}

func (v *validator) validateChecksum(req RequestIngredients) bool {
	if req.Checksum == TestCheckSum {
		core.Logger.Warnf("validateChecksum() - test check detected!")
		return true
	}

	source := fmt.Sprintf("buz:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v",
		req.DeviceID, req.IFA, req.UnitDeviceToken, req.AppID,
		req.CampaignID, req.CampaignType, req.CampaignName, *req.CampaignOwnerID, req.CampaignIsMedia,
		req.Slot, req.Reward, req.BaseReward)

	hash := GetMD5Hash(source)
	hashByPatchedSource := GetMD5Hash(strings.Replace(source, " ", "+", -1)) // encoding문제생긴 string

	checkSumLower := strings.ToLower(req.Checksum) // checksum이 upper case로 들어옴 BZZRWRDD-784

	if req.Checksum == hash {
		return true
	} else if req.Checksum == hashByPatchedSource {
		return true
	} else if checkSumLower == hash {
		return true
	} else if checkSumLower == hashByPatchedSource {
		return true
	}

	return false
}

// GetMD5Hash returns hashed string of the input text.
func GetMD5Hash(text string) string {
	md5Hash := md5.New()
	md5Hash.Write([]byte(text))
	return hex.EncodeToString(md5Hash.Sum(nil))
}
