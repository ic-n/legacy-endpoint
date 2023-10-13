package media

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type ctxKey string // reasoning: https://go.dev/blog/context#package-userip

var userDataKey = ctxKey("userdata")

type userData struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func getUserData(r *http.Request) (*userData, error) {
	auth := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	token, err := jwt.ParseWithClaims(auth, &userData{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("AllYourBase"), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*userData); ok && token.Valid {
		return claims, nil
	}

	return nil, nil
}
