package dto_test

import (
	"testing"

	"time"

	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"
	"github.com/Buzzvil/go-test/test"
)

func TestContentScores(t *testing.T) {
	scores := map[int]int{
		1111: 10,
		2222: 20,
	}
	target := dto.ContentTarget{
		Age:     34,
		Country: "KR",
		Gender:  "M",
	}
	target.SaveContentScores(&scores)
	time.Sleep(time.Millisecond * 100)
	savedScores := target.LoadContentScores()
	t.Logf("TestContentScores() - ori: %v, cache: %v", scores, *savedScores)
	for k, v := range scores {
		test.AssertEqual(t, (*savedScores)[k], v, "TestContentScores()")
	}
}
