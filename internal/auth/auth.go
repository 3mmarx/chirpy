package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}

	return hashedPassword, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}

	return match, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now().UTC()

	rc := jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, rc)

	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	var c jwt.RegisteredClaims
	token, err := jwt.ParseWithClaims(tokenString, &c, func(t *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return uuid.Nil, err
	}

	id, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.Parse(id)
}

func GetBearerToken(headers http.Header) (string, error) {
	value := headers.Get("Authorization")
	if value == "" {
		return "", errors.New("authorization header not found")
	}
	valueS := strings.Fields(value)
	if len(valueS) != 2 || valueS[0] != "Bearer" { // ["Bearer", "TOKEN_STRING"]
		return "", errors.New("invalid Authorization header")
	}

	return valueS[1], nil
}

func MakeRefreshToken() string {
	key := make([]byte, 32)
	rand.Read(key)

	return hex.EncodeToString(key)
}
