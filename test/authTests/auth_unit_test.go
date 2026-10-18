package authTests

import (
	"testing"

	userModels "dalista/dalista-be/app/repository/models/userModels"
	"dalista/dalista-be/app/utils"
	"dalista/dalista-be/config"

	"github.com/google/uuid"
)

func configureJWT(t *testing.T) {
	t.Helper()
	t.Setenv("JWT_KEY", "test-jwt-secret-key-minimum-32-characters!!")
	t.Setenv("JWT_EXPIRY_HOURS", "8")
	config.ReloadFromEnv()
}

func TestAccessTokenRoundTrip(t *testing.T) {
	configureJWT(t)
	user := &userModels.User{ID: uuid.New(), Role: userModels.RoleUser}

	token, _, err := utils.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	sub, err := utils.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("ParseAccessToken: %v", err)
	}
	if sub != user.ID.String() {
		t.Fatalf("sub = %q, want %q", sub, user.ID.String())
	}
}

func TestTokenTypeIsEnforced(t *testing.T) {
	configureJWT(t)
	user := &userModels.User{ID: uuid.New(), Role: userModels.RoleUser}

	access, _, _ := utils.GenerateAccessToken(user)
	refresh, _, _ := utils.GenerateRefreshToken(user)

	// A refresh token must not validate as an access token, and vice versa.
	if _, err := utils.ParseAccessToken(refresh); err == nil {
		t.Fatal("refresh token accepted as access token")
	}
	if _, err := utils.ParseRefreshToken(access); err == nil {
		t.Fatal("access token accepted as refresh token")
	}
}

func TestPasswordHashingAndPolicy(t *testing.T) {
	hash, err := utils.HashPassword("hunter2go")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !utils.VerifyPassword("hunter2go", hash) {
		t.Fatal("VerifyPassword rejected the correct password")
	}
	if utils.VerifyPassword("wrong", hash) {
		t.Fatal("VerifyPassword accepted a wrong password")
	}

	cases := map[string]bool{
		"hunter2go":  true,  // 9 chars, letter + digit
		"short1":     false, // too short
		"allletters": false, // no digit
		"12345678":   false, // no letter
	}
	for in, want := range cases {
		if got := utils.PasswordMeetsPolicy(in); got != want {
			t.Errorf("PasswordMeetsPolicy(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestGenerateOTP(t *testing.T) {
	code, err := utils.GenerateOTP()
	if err != nil {
		t.Fatalf("GenerateOTP: %v", err)
	}
	if len(code) != utils.OTPLength {
		t.Fatalf("OTP length = %d, want %d", len(code), utils.OTPLength)
	}
	for _, r := range code {
		if r < '0' || r > '9' {
			t.Fatalf("OTP %q contains a non-digit", code)
		}
	}
	// Hash is stable + hex-encoded SHA-256 (64 chars).
	if h := utils.HashCode(code); len(h) != 64 || h != utils.HashCode(code) {
		t.Fatalf("HashCode not stable/!=64 chars: %q", h)
	}
}
