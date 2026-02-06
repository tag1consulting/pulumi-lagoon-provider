package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAdminTokenFromSecret(t *testing.T) {
	secret := "test-secret-key-for-jwt"
	audience := "api.test"

	tokenStr, err := generateAdminTokenFromSecret(secret, audience)
	if err != nil {
		t.Fatalf("generateAdminTokenFromSecret failed: %v", err)
	}

	if tokenStr == "" {
		t.Fatal("expected non-empty token")
	}

	// Parse and verify the token
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			t.Fatalf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("expected MapClaims")
	}

	if claims["role"] != "admin" {
		t.Errorf("expected role=admin, got %v", claims["role"])
	}
	if claims["iss"] != "lagoon-api" {
		t.Errorf("expected iss=lagoon-api, got %v", claims["iss"])
	}
	if claims["sub"] != "lagoonadmin" {
		t.Errorf("expected sub=lagoonadmin, got %v", claims["sub"])
	}
	if claims["aud"] != "api.test" {
		t.Errorf("expected aud=api.test, got %v", claims["aud"])
	}
}

func TestGenerateAdminTokenFromSecret_DefaultAudience(t *testing.T) {
	secret := "test-secret"
	tokenStr, err := generateAdminTokenFromSecret(secret, "")
	if err != nil {
		t.Fatalf("generateAdminTokenFromSecret failed: %v", err)
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	claims := token.Claims.(jwt.MapClaims)
	if claims["aud"] != "api.dev" {
		t.Errorf("expected aud=api.dev (default), got %v", claims["aud"])
	}
}

func TestGenerateAdminTokenFromSecret_Expiry(t *testing.T) {
	secret := "test-secret"
	before := time.Now()

	tokenStr, err := generateAdminTokenFromSecret(secret, "api.dev")
	if err != nil {
		t.Fatalf("generateAdminTokenFromSecret failed: %v", err)
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	claims := token.Claims.(jwt.MapClaims)

	exp, err := claims.GetExpirationTime()
	if err != nil {
		t.Fatalf("failed to get exp: %v", err)
	}

	// Token should expire ~1 hour from now
	expectedExpiry := before.Add(1 * time.Hour)
	if exp.Time.Before(before) || exp.Time.After(expectedExpiry.Add(5*time.Second)) {
		t.Errorf("token expiry %v is not within expected range", exp.Time)
	}
}

func TestConfigure_WithToken(t *testing.T) {
	cfg := &LagoonConfig{
		APIUrl: "https://api.test/graphql",
		Token:  "pre-set-token",
	}

	if err := cfg.Configure(nil); err != nil {
		t.Fatalf("Configure failed: %v", err)
	}

	if cfg.Token != "pre-set-token" {
		t.Errorf("expected token to remain 'pre-set-token', got %s", cfg.Token)
	}
}

func TestConfigure_WithJWTSecret(t *testing.T) {
	cfg := &LagoonConfig{
		APIUrl:      "https://api.test/graphql",
		JWTSecret:   "my-secret",
		JWTAudience: "api.test",
	}

	if err := cfg.Configure(nil); err != nil {
		t.Fatalf("Configure failed: %v", err)
	}

	if cfg.Token == "" {
		t.Error("expected token to be generated from JWT secret")
	}

	// Verify it's a valid JWT
	if !strings.Contains(cfg.Token, ".") {
		t.Error("expected token to be a JWT (contains dots)")
	}
}

func TestConfigure_NoAuth(t *testing.T) {
	// Clear env vars that might interfere
	origToken := os.Getenv("LAGOON_TOKEN")
	origSecret := os.Getenv("LAGOON_JWT_SECRET")
	os.Unsetenv("LAGOON_TOKEN")
	os.Unsetenv("LAGOON_JWT_SECRET")
	defer func() {
		if origToken != "" {
			os.Setenv("LAGOON_TOKEN", origToken)
		}
		if origSecret != "" {
			os.Setenv("LAGOON_JWT_SECRET", origSecret)
		}
	}()

	cfg := &LagoonConfig{
		APIUrl: "https://api.test/graphql",
	}

	err := cfg.Configure(nil)
	if err == nil {
		t.Fatal("expected error when no authentication is configured")
	}

	if !strings.Contains(err.Error(), "lagoon authentication required") {
		t.Errorf("expected authentication error, got: %v", err)
	}
}

func TestConfigure_EnvVarToken(t *testing.T) {
	origToken := os.Getenv("LAGOON_TOKEN")
	origSecret := os.Getenv("LAGOON_JWT_SECRET")
	os.Setenv("LAGOON_TOKEN", "env-token")
	os.Unsetenv("LAGOON_JWT_SECRET")
	defer func() {
		if origToken != "" {
			os.Setenv("LAGOON_TOKEN", origToken)
		} else {
			os.Unsetenv("LAGOON_TOKEN")
		}
		if origSecret != "" {
			os.Setenv("LAGOON_JWT_SECRET", origSecret)
		}
	}()

	cfg := &LagoonConfig{
		APIUrl: "https://api.test/graphql",
	}

	if err := cfg.Configure(nil); err != nil {
		t.Fatalf("Configure failed: %v", err)
	}

	if cfg.Token != "env-token" {
		t.Errorf("expected token from env var, got %s", cfg.Token)
	}
}

func TestConfigure_EnvVarJWTSecret(t *testing.T) {
	origToken := os.Getenv("LAGOON_TOKEN")
	origSecret := os.Getenv("LAGOON_JWT_SECRET")
	os.Unsetenv("LAGOON_TOKEN")
	os.Setenv("LAGOON_JWT_SECRET", "env-jwt-secret")
	defer func() {
		if origToken != "" {
			os.Setenv("LAGOON_TOKEN", origToken)
		} else {
			os.Unsetenv("LAGOON_TOKEN")
		}
		if origSecret != "" {
			os.Setenv("LAGOON_JWT_SECRET", origSecret)
		} else {
			os.Unsetenv("LAGOON_JWT_SECRET")
		}
	}()

	cfg := &LagoonConfig{
		APIUrl:      "https://api.test/graphql",
		JWTAudience: "api.dev",
	}

	if err := cfg.Configure(nil); err != nil {
		t.Fatalf("Configure failed: %v", err)
	}

	if cfg.Token == "" {
		t.Error("expected token to be generated from LAGOON_JWT_SECRET env var")
	}
}

func TestNewClient_Basic(t *testing.T) {
	cfg := &LagoonConfig{
		APIUrl: "https://api.test/graphql",
		Token:  "test-token",
	}

	client := cfg.NewClient()
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_WithInsecure(t *testing.T) {
	cfg := &LagoonConfig{
		APIUrl:   "https://api.test/graphql",
		Token:    "test-token",
		Insecure: true,
	}

	client := cfg.NewClient()
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_WithJWTSecret(t *testing.T) {
	cfg := &LagoonConfig{
		APIUrl:      "https://api.test/graphql",
		Token:       "test-token",
		JWTSecret:   "my-secret",
		JWTAudience: "api.test",
	}

	c := cfg.NewClient()
	if c == nil {
		t.Fatal("expected non-nil client")
	}
	// We can't directly check tokenFunc since it's unexported,
	// but we can verify the client was created successfully with JWTSecret
}

func TestConfigure_TokenPrecedence(t *testing.T) {
	// Direct token takes precedence over JWT secret
	cfg := &LagoonConfig{
		APIUrl:    "https://api.test/graphql",
		Token:     "direct-token",
		JWTSecret: "some-secret",
	}

	if err := cfg.Configure(nil); err != nil {
		t.Fatalf("Configure failed: %v", err)
	}

	if cfg.Token != "direct-token" {
		t.Errorf("expected direct token to take precedence, got %s", cfg.Token)
	}
}

func TestConfigure_JWTSecretGeneratesToken(t *testing.T) {
	origToken := os.Getenv("LAGOON_TOKEN")
	origSecret := os.Getenv("LAGOON_JWT_SECRET")
	os.Unsetenv("LAGOON_TOKEN")
	os.Unsetenv("LAGOON_JWT_SECRET")
	defer func() {
		if origToken != "" {
			os.Setenv("LAGOON_TOKEN", origToken)
		}
		if origSecret != "" {
			os.Setenv("LAGOON_JWT_SECRET", origSecret)
		}
	}()

	cfg := &LagoonConfig{
		APIUrl:      "https://api.test/graphql",
		JWTSecret:   "test-secret",
		JWTAudience: "api.dev",
	}

	if err := cfg.Configure(nil); err != nil {
		t.Fatalf("Configure failed: %v", err)
	}

	// Verify the generated token is valid
	parts := strings.Split(cfg.Token, ".")
	if len(parts) != 3 {
		t.Errorf("expected JWT with 3 parts, got %d", len(parts))
	}
}
