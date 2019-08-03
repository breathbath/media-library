package authentication

import (
	"fmt"
	"github.com/breathbath/go_utils/utils/env"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type JwtManager struct {
	issuer        string
	tokenDuration time.Duration
	secret        string
}

const (
	tokenExpireSecondsAfterLogout = 3600
	tokenAudience                 = "token_blug"
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

	tokenDuration := env.ReadEnvInt("TOKEN_DURATION_HOURS", 72)

	return &JwtManager{
		secret:        secret,
		issuer:        issuer,
		tokenDuration: time.Hour * time.Duration(tokenDuration),
	}, nil
}

func (jwtm *JwtManager) GenerateToken(login string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = jwt.MapClaims{
		"exp": time.Now().UTC().Add(jwtm.tokenDuration).Unix(),
		"iat": time.Now().UTC().Unix(),
		"sub": login,
		"iss": jwtm.issuer,
		"aud": tokenAudience,
	}
	tokenString, err := token.SignedString([]byte(jwtm.secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (jwtm *JwtManager) getTokenRemainingValidity(timestamp interface{}) time.Duration {
	if validity, ok := timestamp.(float64); ok {
		tm := time.Unix(int64(validity), 0)
		remainer := tm.Sub(time.Now().UTC())
		if remainer > 0 {
			return time.Second * time.Duration(remainer.Seconds()+tokenExpireSecondsAfterLogout)
		}
	}
	return time.Second * time.Duration(tokenExpireSecondsAfterLogout)
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
