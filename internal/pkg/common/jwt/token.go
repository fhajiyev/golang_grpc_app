package jwt

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

// GetServiceToken func definition
func GetServiceToken() string {
	expMin := 5
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"service": "bs_go",
		"exp":     time.Now().Add(time.Minute * time.Duration(expMin)).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte("rtgFGE4037yfrqFcgwc036BceGmRadYnmslOxuZ2"))
	if err != nil {
		panic(err)
	}
	return tokenString
}
