package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	password := "my-secret-password"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() returned an error: %v", err)
	}

	if hash == "" {
		t.Fatal("HashPassword() returned an empty hash")
	}

	if hash == password {
		t.Fatal("HashPassword() returned the original password")
	}
}

func TestHashPasswordGeneratesDifferentHashes(t *testing.T) {
	password := "my-secret-password"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("first HashPassword() failed: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("second HashPassword() failed: %v", err)
	}

	if hash1 == hash2 {
		t.Fatal("expected different hashes due to random salts")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "my-secret-password"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{
			name:     "correct password",
			password: password,
			want:     true,
		},
		{
			name:     "incorrect password",
			password: "wrong-password",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := CheckPasswordHash(tt.password, hash)
			if err != nil {
				t.Fatalf("CheckPasswordHash() returned an error: %v", err)
			}

			if match != tt.want {
				t.Errorf("CheckPasswordHash() = %v, want %v", match, tt.want)
			}
		})
	}
}

func TestCheckPasswordHashInvalidHash(t *testing.T) {
	_, err := CheckPasswordHash("password", "invalid-hash")

	if err == nil {
		t.Fatal("expected an error for an invalid hash")
	}
}

func TestMakeAndValidateJWT(t *testing.T) {
	id := uuid.New()
	tokenSecret := "secret"
	tokenString, err := MakeJWT(id, tokenSecret, time.Minute)
	if err != nil {
		t.Fatal("expected a token, got error", err)
	}

	tId, err := ValidateJWT(tokenString, tokenSecret)
	if err != nil {
		t.Fatal(err)
	}

	if tId != id {
		t.Fatalf("expected %v, got %v", id, tId)
	}
}

func TestJWTWithWrongSecret(t *testing.T) {
	id := uuid.New()
	tokenSecret := "secret"
	tokenString, err := MakeJWT(id, tokenSecret, time.Minute)
	if err != nil {
		t.Fatal("expected a token, got error", err)
	}

	_, err = ValidateJWT(tokenString, "invalid-secret") // wrong secret
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func TestJWTWithExpiredTime(t *testing.T) {
	id := uuid.New()
	tokenSecret := "secret"
	tokenString, err := MakeJWT(id, tokenSecret, -5*time.Minute) // expired 5 minutes ago
	if err != nil {
		t.Fatal("expected a token, got error", err)
	}

	_, err = ValidateJWT(tokenString, tokenSecret)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		wantToken  string
		wantErr    bool
	}{
		{
			name:       "missing header",
			authHeader: "",
			wantToken:  "",
			wantErr:    true,
		},
		{
			name:       "wrong scheme",
			authHeader: "Basic abc123",
			wantToken:  "",
			wantErr:    true,
		},
		{
			name:       "Bearer with no token",
			authHeader: "Bearer",
			wantToken:  "",
			wantErr:    true,
		},
		{
			name:       "extra parts",
			authHeader: "Bearer abc123 extra",
			wantToken:  "",
			wantErr:    true,
		},
		{
			name:       "valid header",
			authHeader: "Bearer abc123",
			wantToken:  "abc123",
			wantErr:    false,
		},
		{
			name:       "extra whitespace around valid header",
			authHeader: "  Bearer   abc123  ",
			wantToken:  "abc123",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := http.Header{}
			if tt.authHeader != "" {
				headers.Set("Authorization", tt.authHeader)
			}

			gotToken, err := GetBearerToken(headers)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetBearerToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotToken != tt.wantToken {
				t.Errorf("GetBearerToken() = %q, want %q", gotToken, tt.wantToken)
			}
		})
	}
}
