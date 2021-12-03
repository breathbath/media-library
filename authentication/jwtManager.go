package authentication

import (
	"fmt"
	"github.com/breathbath/go_utils/utils/env"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type JwtManager struct {
	issuer        string
	tokenDuration time.Duration
	secret        string
}

const (
	tokenAudience = "media_service"
)

func NewJwtManager() (*JwtManager, error) {
	issuer, err := env.ReadEnvOrError("TOKEN_ISSUER")
	if err != nil {
		return nil, err
	}

	secret, err := env.ReadEnvOrError("TOKEN_SECRET")
	if err != nil {
		return nil, err
	}

	tokenDuration := env.ReadEnvInt("TOKEN_DURATION_DAYS", 30)

	return &JwtManager{
		secret:        secret,
		issuer:        issuer,
		tokenDuration: time.Hour * 24 * time.Duration(tokenDuration),
	}, nil
}

func (jwtm *JwtManager) GenerateToken(appName string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = jwt.MapClaims{
		"exp": time.Now().UTC().Add(jwtm.tokenDuration).Unix(),
		"iat": time.Now().UTC().Unix(),
		"sub": appName,
		"iss": jwtm.issuer,
		"aud": tokenAudience,
	}
	tokenString, err := token.SignedString([]byte(jwtm.secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (jwtm *JwtManager) ParseToken(rawToken string) (*jwt.Token, error) {
	token, err := jwt.Parse(rawToken, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(jwtm.secret), nil
	})

	return token, err
}
