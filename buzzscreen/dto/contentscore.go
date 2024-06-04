package dto

import (
	"fmt"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
)

type (
	// ContentScoresRequest type definition
	ContentScoresRequest struct {
		Scores map[int]int `json:"scores"`
	}

	// ContentTarget type definition
	ContentTarget struct {
		Age     int
		Country string
		Gender  string
	}
)

func (ct *ContentTarget) getCacheKeyContentTarget() string {
	return fmt.Sprintf("CACHE_GO_CONTENT_TARGET_%s_%s_%d", ct.Country, ct.Gender, ct.Age)
}

// SaveContentScores func definition
func (ct *ContentTarget) SaveContentScores(scores *map[int]int) {
	core.Logger.Infof("SaveContentScores() - target: %v, scores: %v", *ct, *scores)
	env.SetCache(ct.getCacheKeyContentTarget(), *scores, time.Hour*2)
}

// LoadContentScores func definition
func (ct *ContentTarget) LoadContentScores() *map[int]int {
	scores := make(map[int]int)
	scoresCacheKey := ct.getCacheKeyContentTarget()
	if err := env.GetCache(scoresCacheKey, &scores); err != nil {
		core.Logger.Warnf("LoadContentScores() - No score for target %+v", ct)
	}
	core.Logger.Debugf("LoadContentScores() - target: %v, scores: %v", *ct, scores)
	return &scores
}
