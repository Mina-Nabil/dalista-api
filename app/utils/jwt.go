package utils

import (
	"errors"
	"fmt"
	"time"

	userModels "dalista/dalista-be/app/repository/models/userModels"
	"dalista/dalista-be/config"

	"github.com/golang-jwt/jwt/v5"
)

// Token types carried in the `typ` claim so an access token can't be replayed
// as a refresh token or vice versa.
const (
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"

	// Refresh tokens outlive access tokens; access expiry is JWT_EXPIRY_HOURS.
	refreshTokenDays = 30
)

// ErrWrongTokenType is returned when a token's `typ` claim doesn't match the
// expected use (e.g. a refresh token presented to the access middleware).
var ErrWrongTokenType = errors.New("unexpected token type")

// GenerateAccessToken mints a short-lived access JWT for the user.
func GenerateAccessToken(user *userModels.User) (string, time.Time, error) {
	expiry := time.Now().Add(time.Duration(config.GetConfig().JWTExpiryHours) * time.Hour)
	return signToken(user, tokenTypeAccess, expiry)
}

// GenerateRefreshToken mints a long-lived refresh JWT for the user.
func GenerateRefreshToken(user *userModels.User) (string, time.Time, error) {
	expiry := time.Now().Add(refreshTokenDays * 24 * time.Hour)
	return signToken(user, tokenTypeRefresh, expiry)
}

func signToken(user *userModels.User, typ string, expiry time.Time) (string, time.Time, error) {
	claims := jwt.MapClaims{
		"sub": user.ID.String(),
		"typ": typ,
		"exp": expiry.Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(config.GetConfig().JWTKey))
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiry, nil
}

func parseClaims(tokenString string) (jwt.MapClaims, error) {
	parsed, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.GetConfig().JWTKey), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// ParseAccessToken validates signature + expiry and requires typ=access,
// returning the subject (user id).
func ParseAccessToken(tokenString string) (string, error) {
	return parseTokenOfType(tokenString, tokenTypeAccess)
}

// ParseRefreshToken validates signature + expiry and requires typ=refresh,
// returning the subject (user id).
func ParseRefreshToken(tokenString string) (string, error) {
	return parseTokenOfType(tokenString, tokenTypeRefresh)
}

func parseTokenOfType(tokenString, want string) (string, error) {
	claims, err := parseClaims(tokenString)
	if err != nil {
		return "", err
	}
	if typ, _ := claims["typ"].(string); typ != want {
		return "", ErrWrongTokenType
	}
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", errors.New("invalid token subject")
	}
	return sub, nil
}
