package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	if len(password) > 71 {
		return "", errors.New("password too long")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", errors.New("unable to hash password")
	}

	stringHashed := string(hashed)

	return stringHashed, nil
}

func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return err
	}
	return nil
}

type Token string

const (
	TokenTypeAccess Token = "terminus-access"
)

func MakeJWT(
	userID uuid.UUID,
	tokenSecret string,
	expiresIn time.Duration,
) (string, error) {
	signingKey := []byte(tokenSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    string(TokenTypeAccess),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	})
	return token.SignedString(signingKey)
}

//func CheckPasswordHash(password, hash string) error

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil },
	)
	if err != nil {
		return uuid.Nil, err
	}

	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}
	if issuer != string(TokenTypeAccess) {
		return uuid.Nil, errors.New("invalid issuer")
	}

	id, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID: %w", err)
	}
	return id, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authorization := headers.Get("Authorization")
	if authorization == "" {
		return "", errors.New("missing authorization header")
	}

	if !strings.HasPrefix(authorization, "Bearer ") {
		return "", errors.New("expected Bearer authorization scheme")
	}

	// Trim off the prefix and whitespace
	tokenString := strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
	if tokenString == "" {
		return "", errors.New("missing bearer token")
	}
	log.Printf("returning token: %s", tokenString)
	return tokenString, nil
}

// It should use the following to generate a random 256-bit (32-byte) hex-encoded string:
// rand.Read to generate 32 bytes (256 bits) of random data from the crypto/rand package (math/rand's Read function is deprecated).
// hex.EncodeToString to convert the random data to a hex string

func MakeRefreshToken() (string, error) {
	key := make([]byte, 32)
	rand.Read(key)
	encodedString := hex.EncodeToString(key)

	return encodedString, nil
}

func GetAPIKey(headers http.Header) (string, error) {
	// shape Authorization: ApiKey THE_KEY_HERE
	authorization := headers.Get("Authorization")
	if authorization == "" {
		return "", errors.New("missing authorization header")
	}
	if !strings.HasPrefix(authorization, "ApiKey ") {
		return "", errors.New("expected ApiKey ")
	}
	tokenString := strings.TrimSpace(strings.TrimPrefix(authorization, "ApiKey "))
	if tokenString == "" {
		return "", errors.New("missing ApiKey ")
	}
	log.Printf("returning token: %s", tokenString)
	return tokenString, nil

}
