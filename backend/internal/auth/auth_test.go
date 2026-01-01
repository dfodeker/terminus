package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCheckPasswordHash(t *testing.T) {
	// First, we need to create some hashed passwords for testing
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "Correct password",
			password: password1,
			hash:     hash1,
			wantErr:  false,
		},
		{
			name:     "Incorrect password",
			password: "wrongPassword",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Password doesn't match different hash",
			password: password1,
			hash:     hash2,
			wantErr:  true,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Invalid hash",
			password: password1,
			hash:     "invalidhash",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	validToken, _ := MakeJWT(userID, "secret", time.Hour)

	tests := []struct {
		name        string
		tokenString string
		tokenSecret string
		wantUserID  uuid.UUID
		wantErr     bool
	}{
		{
			name:        "Valid token",
			tokenString: validToken,
			tokenSecret: "secret",
			wantUserID:  userID,
			wantErr:     false,
		},
		{
			name:        "Invalid token",
			tokenString: "invalid.token.string",
			tokenSecret: "secret",
			wantUserID:  uuid.Nil,
			wantErr:     true,
		},
		{
			name:        "Wrong secret",
			tokenString: validToken,
			tokenSecret: "wrong_secret",
			wantUserID:  uuid.Nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUserID, err := ValidateJWT(tt.tokenString, tt.tokenSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUserID != tt.wantUserID {
				t.Errorf("ValidateJWT() gotUserID = %v, want %v", gotUserID, tt.wantUserID)
			}
		})
	}
}

func TestMakeRefreshToken(t *testing.T) {
	//function performs a single action
	tests := []struct {
		name      string
		wantValue bool
		wantErr   bool
	}{
		{
			name:      "Return Token",
			wantValue: true,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRefreshToken, err := MakeRefreshToken()
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeRefreshToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRefreshToken == "" {
				t.Errorf("MakeRefreshToken() Expected Refresh Token, recieved %v", gotRefreshToken)
				return
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	userID := uuid.New()
	validToken, _ := MakeJWT(userID, "secret", time.Hour)
	invalidToken := "Invalid.Token"

	// var validHeader http.Header
	// validHeader.Set("Authorization", "Bearer "+validToken)
	validHeader := http.Header{
		"Authorization": []string{"Bearer " + validToken},
	}
	invalidHeader := http.Header{
		"Authorization": []string{"Bearer" + invalidToken},
	}
	justBearer := http.Header{
		"Authorization": []string{"Bearer "}, // no space, no token
	}
	nonJwt := http.Header{
		"Authorization": []string{"Bearer " + invalidToken}, // no space, no token
	}
	bearerNotPresent := http.Header{
		"Authorization": []string{invalidToken},
	}

	tests := []struct {
		name            string
		header          http.Header
		wantTokenString string
		wantErr         bool
	}{
		{
			name:            "Valid Bearer",
			header:          validHeader,
			wantTokenString: validToken,
			wantErr:         false,
		},
		{
			name:            "Invalid Bearer",
			header:          invalidHeader,
			wantTokenString: "",
			wantErr:         true,
		},
		{
			name:            "Bearer Present But No Token",
			header:          justBearer,
			wantTokenString: "",
			wantErr:         true,
		},
		{
			name:            "NonJWT",
			header:          nonJwt,
			wantTokenString: invalidToken,
			wantErr:         false,
		},
		{
			name:            "Bearer not Present",
			header:          bearerNotPresent,
			wantTokenString: "",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTokenString, err := GetBearerToken(tt.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBearerToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantTokenString != gotTokenString {

				t.Errorf("GetBearerToken() gotTokenString = %v, want %v", gotTokenString, tt.wantTokenString)
			}
		})
	}

}
