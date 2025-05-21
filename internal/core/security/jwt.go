package security

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rshafikov/gophermart/internal/app"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"go.uber.org/zap"
	"strings"
	"time"
)

const TokenExpTime = 60 * time.Minute
const TokenType = "Bearer"

var ErrTokenInvalid = errors.New("token is invalid")
var ErrUnableToParseToken = errors.New("unable to parse token")

type JWTToken struct {
	Token     string    `json:"token"`
	TokenType string    `json:"token_type"`
	ExpiresAt time.Time `json:"expires_at"`
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
	expires := jwt.NewNumericDate(time.Now().Add(TokenExpTime))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, TokenPayload{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: expires,
			Subject:   login,
		},
	})

	tokenString, err := token.SignedString([]byte(app.Env.Secret))
	if err != nil {
		return nil, err
	}

	return &JWTToken{Token: tokenString, TokenType: TokenType, ExpiresAt: expires.Time}, nil
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
		logger.L.Debug("unable to parse token:", zap.Error(err))
		return nil, ErrUnableToParseToken
	}

	if !token.Valid {
		logger.L.Debug("token is not valid")
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

type MockJWTHandler struct{}

func (m *MockJWTHandler) GenerateJWT(login string) (*JWTToken, error) {
	return &JWTToken{Token: "fake-token " + login, TokenType: TokenType}, nil
}

func (m *MockJWTHandler) ParseJWT(token string) (*TokenPayload, error) {
	splitedToken := strings.Split(token, "fake-token ")
	if len(splitedToken) != 2 || splitedToken[0] != "" {
		return nil, ErrTokenInvalid
	}

	return &TokenPayload{jwt.RegisteredClaims{Subject: splitedToken[1]}}, nil
}
