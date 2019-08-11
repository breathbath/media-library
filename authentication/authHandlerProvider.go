package authentication

import (
	"context"
	"github.com/breathbath/go_utils/utils/env"
	"github.com/breathbath/go_utils/utils/io"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"strings"
)

type AuthHandlerProvider struct {
	jwtManager          *JwtManager
	acceptedTokenIssuer string
}

func NewAuthHandlerProvider(jwtManager *JwtManager) *AuthHandlerProvider {
	return &AuthHandlerProvider{
		jwtManager:          jwtManager,
		acceptedTokenIssuer: env.ReadEnv("TOKEN_ISSUER", ""),
	}
}

func (ahp *AuthHandlerProvider) GetHandlerFunc() func(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	return ahp.AuthenticateClient
}

func (ahp *AuthHandlerProvider) AuthenticateClient(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	rawTokens, ok := req.Header["Authorization"]
	if !ok {
		next(rw, req)
		return
	}

	rawToken := rawTokens[0]
	if rawToken == "" {
		next(rw, req)
		return
	}

	rawToken = strings.Replace(rawToken, "Bearer ", "", -1)

	token, err := ahp.jwtManager.ParseToken(rawToken)

	if err != nil {
		io.OutputError(err, "Auth handler", "Invalid token: "+rawToken)
		next(rw, req)
		return
	}

	if !token.Valid || token.Claims.(jwt.MapClaims)["iss"].(string) != ahp.acceptedTokenIssuer {
		next(rw, req)
		return
	}

	ctx := context.WithValue(req.Context(), "token", token)
	next(rw, req.WithContext(ctx))
}
