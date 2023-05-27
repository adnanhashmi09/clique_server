package utils

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func Generate_JWT_Token(id string, username string) (string, error) {
	secretKey := []byte(Get_Env_Variable("SECRET"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomClaims{
		ID:       id,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    id,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	})

	ss, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return ss, err
}

func Verify_JWT_Token(token_value string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(token_value, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		secretKey := []byte(Get_Env_Variable("SECRET"))
		return secretKey, nil
	})

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		log.Printf("%v %v", claims, claims.RegisteredClaims.Issuer)
		return claims, nil
	} else {
		log.Println(err)
		return nil, err
	}
}
