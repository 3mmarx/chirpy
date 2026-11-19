package auth

import (
	"testing"
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
