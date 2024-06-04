package controller

import (
	"net/http"
	"os"

	"strconv"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"
)

// PostContentScores func definition
func PostContentScores(c core.Context) error {
	authValues := c.Request().Header["Authorization"]
	if len(authValues) < 1 || authValues[0] != os.Getenv("BASIC_AUTHORIZATION_VALUE") {
		return c.NoContent(http.StatusForbidden)
	}

	var contentReq dto.ContentScoresRequest
	if err := bindValue(c, &contentReq); err != nil {
		return err
	}

	age, err := strconv.ParseFloat(c.QueryParam("age"), 32)

	if err != nil {
		return err
	}

	core.Logger.Debugf("PostContentScores() contentReq: %+v", contentReq)

	if len(contentReq.Scores) > 0 {
		(&dto.ContentTarget{
			Age:     int(age),
			Country: c.QueryParam("country"),
			Gender:  c.QueryParam("gender"),
		}).SaveContentScores(&contentReq.Scores)
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code": dto.CodeOk,
		})
	}
	return c.JSON(http.StatusBadRequest, map[string]interface{}{
		"code":    dto.CodeBadRequest,
		"message": "Score is not provided.",
	})
}
