package security

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rshafikov/gophermart/internal/app"
	"log"
	"time"
)

const TokenExpTime = 60 * time.Minute
const TokenType = "Bearer"

type JWTToken struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
}

type TokenPayload struct {
	jwt.RegisteredClaims
}

type JWTHandler interface {
	GenerateJWT(login string) (*JWTToken, error)
	ParseJWT(tokenString string) (*TokenPayload, error)
}

type jwtHandler struct{}

func NewJWTHandler() JWTHandler {
	return &jwtHandler{}
}

func (j *jwtHandler) GenerateJWT(login string) (*JWTToken, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, TokenPayload{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpTime)),
			Subject:   login,
		},
	})

	tokenString, err := token.SignedString([]byte(app.Env.Secret))
	if err != nil {
		return nil, err
	}

	return &JWTToken{Token: tokenString, TokenType: TokenType}, nil
}

func (j *jwtHandler) ParseJWT(tokenString string) (*TokenPayload, error) {
	claims := &TokenPayload{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(app.Env.Secret), nil
		})
	if err != nil {
		log.Println("unable to parse token:", err)
		return nil, errors.New("unable to parse token")
	}

	if !token.Valid {
		log.Println("JWTToken is not valid")
		return nil, errors.New("JWTToken is not valid")
	}

	log.Println("JWTToken os valid")
	return claims, nil
}
