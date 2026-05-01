package config

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
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

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("expected MapClaims")
	}
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

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("expected MapClaims")
	}

	exp, err := claims.GetExpirationTime()
	if err != nil {
		t.Fatalf("failed to get exp: %v", err)
	}

	// Token should expire ~1 hour from now
	expectedExpiry := before.Add(1 * time.Hour)
	if exp.Before(before) || exp.After(expectedExpiry.Add(5*time.Second)) {
		t.Errorf("token expiry %v is not within expected range", exp)
	}
}

func TestConfigure_WithToken(t *testing.T) {
	cfg := &LagoonConfig{
		APIUrl: "https://api.test/graphql",
		Token:  "pre-set-token",
	}

	if err := cfg.Configure(context.TODO()); err != nil {
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

	if err := cfg.Configure(context.TODO()); err != nil {
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
	origToken, hadToken := os.LookupEnv("LAGOON_TOKEN")
	origSecret, hadSecret := os.LookupEnv("LAGOON_JWT_SECRET")
	os.Unsetenv("LAGOON_TOKEN")
	os.Unsetenv("LAGOON_JWT_SECRET")
	defer func() {
		if hadToken {
			os.Setenv("LAGOON_TOKEN", origToken)
		} else {
			os.Unsetenv("LAGOON_TOKEN")
		}
		if hadSecret {
			os.Setenv("LAGOON_JWT_SECRET", origSecret)
		} else {
			os.Unsetenv("LAGOON_JWT_SECRET")
		}
	}()

	cfg := &LagoonConfig{
		APIUrl: "https://api.test/graphql",
	}

	err := cfg.Configure(context.TODO())
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

	if err := cfg.Configure(context.TODO()); err != nil {
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

	if err := cfg.Configure(context.TODO()); err != nil {
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

	if err := cfg.Configure(context.TODO()); err != nil {
		t.Fatalf("Configure failed: %v", err)
	}

	if cfg.Token != "direct-token" {
		t.Errorf("expected direct token to take precedence, got %s", cfg.Token)
	}
	if cfg.JWTSecret != "" {
		t.Errorf("expected JWTSecret to be cleared when direct token takes precedence, got %s", cfg.JWTSecret)
	}
}

func TestConfigure_TrimsJWTSecretWhitespace(t *testing.T) {
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

	secret := "test-secret-for-trimming"

	cleanCfg := &LagoonConfig{
		APIUrl:      "https://api.test/graphql",
		JWTSecret:   secret,
		JWTAudience: "api.dev",
	}
	if err := cleanCfg.Configure(context.TODO()); err != nil {
		t.Fatalf("Configure (clean) failed: %v", err)
	}

	cases := []struct {
		name      string
		jwtSecret string
	}{
		{"trailing newline", secret + "\n"},
		{"trailing space", secret + " "},
		{"leading space", " " + secret},
		{"leading and trailing whitespace", "  " + secret + "\t\n"},
		{"trailing CRLF", secret + "\r\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &LagoonConfig{
				APIUrl:      "https://api.test/graphql",
				JWTSecret:   tc.jwtSecret,
				JWTAudience: "api.dev",
			}
			if err := cfg.Configure(context.TODO()); err != nil {
				t.Fatalf("Configure failed: %v", err)
			}
			cleanToken, err := jwt.Parse(cleanCfg.Token, func(t *jwt.Token) (any, error) {
				return []byte(secret), nil
			})
			if err != nil {
				t.Fatalf("failed to parse clean token: %v", err)
			}
			dirtyToken, err := jwt.Parse(cfg.Token, func(t *jwt.Token) (any, error) {
				return []byte(secret), nil
			})
			if err != nil {
				t.Fatalf("failed to parse dirty token: %v", err)
			}
			if !dirtyToken.Valid {
				t.Errorf("token generated from %q should be valid when verified with trimmed secret", tc.name)
			}
			cleanClaims, ok := cleanToken.Claims.(jwt.MapClaims)
			if !ok {
				t.Fatal("expected MapClaims from clean token")
			}
			dirtyClaims, ok := dirtyToken.Claims.(jwt.MapClaims)
			if !ok {
				t.Fatal("expected MapClaims from dirty token")
			}
			if cleanClaims["role"] != dirtyClaims["role"] || cleanClaims["aud"] != dirtyClaims["aud"] {
				t.Errorf("claims mismatch between clean and whitespace-padded secret tokens")
			}
		})
	}
}

func TestConfigure_TrimsTokenWhitespace(t *testing.T) {
	cfg := &LagoonConfig{
		APIUrl: "https://api.test/graphql",
		Token:  "  my-token\n",
	}
	if err := cfg.Configure(context.TODO()); err != nil {
		t.Fatalf("Configure failed: %v", err)
	}
	if cfg.Token != "my-token" {
		t.Errorf("expected trimmed token 'my-token', got %q", cfg.Token)
	}
}

func TestConfigure_TrimsEnvVarWhitespace(t *testing.T) {
	origToken := os.Getenv("LAGOON_TOKEN")
	origSecret := os.Getenv("LAGOON_JWT_SECRET")
	os.Setenv("LAGOON_TOKEN", "  env-token\n")
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
	if err := cfg.Configure(context.TODO()); err != nil {
		t.Fatalf("Configure failed: %v", err)
	}
	if cfg.Token != "env-token" {
		t.Errorf("expected trimmed env token 'env-token', got %q", cfg.Token)
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

	if err := cfg.Configure(context.TODO()); err != nil {
		t.Fatalf("Configure failed: %v", err)
	}

	// Verify the generated token is valid
	parts := strings.Split(cfg.Token, ".")
	if len(parts) != 3 {
		t.Errorf("expected JWT with 3 parts, got %d", len(parts))
	}
}

// ==================== Diff (CustomDiff) ====================

func TestDiff_NoChanges(t *testing.T) {
	c := &LagoonConfig{}
	resp, err := c.Diff(context.Background(), infer.DiffRequest[LagoonConfig, LagoonConfig]{
		Inputs: LagoonConfig{APIUrl: "https://api.test/graphql", Token: "tok", JWTAudience: "api.dev"},
		State:  LagoonConfig{APIUrl: "https://api.test/graphql", Token: "tok", JWTAudience: "api.dev"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.HasChanges {
		t.Error("expected no changes when config is identical")
	}
}

func TestDiff_WhitespaceDifferencesIgnored(t *testing.T) {
	c := &LagoonConfig{}
	resp, err := c.Diff(context.Background(), infer.DiffRequest[LagoonConfig, LagoonConfig]{
		Inputs: LagoonConfig{Token: "my-token\n", JWTSecret: " secret "},
		State:  LagoonConfig{Token: "my-token", JWTSecret: "secret"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.HasChanges {
		t.Error("expected no changes when only whitespace differs")
	}
}

func TestDiff_NeverReplace(t *testing.T) {
	c := &LagoonConfig{}
	resp, err := c.Diff(context.Background(), infer.DiffRequest[LagoonConfig, LagoonConfig]{
		Inputs: LagoonConfig{APIUrl: "https://new-api.test/graphql", Token: "new-tok", JWTSecret: "new-secret", JWTAudience: "api.prod", Insecure: true},
		State:  LagoonConfig{APIUrl: "https://old-api.test/graphql", Token: "old-tok", JWTSecret: "old-secret", JWTAudience: "api.dev", Insecure: false},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.HasChanges {
		t.Error("expected changes when all fields differ")
	}
	if resp.DeleteBeforeReplace {
		t.Error("DeleteBeforeReplace must always be false for provider config")
	}
	for field, d := range resp.DetailedDiff {
		if d.Kind != p.Update {
			t.Errorf("field %q has Kind=%v, want Update (never replace)", field, d.Kind)
		}
	}
	expectedFields := []string{"apiUrl", "token", "jwtSecret", "jwtAudience", "insecure"}
	for _, f := range expectedFields {
		if _, ok := resp.DetailedDiff[f]; !ok {
			t.Errorf("expected %q in DetailedDiff", f)
		}
	}
}

func TestDiff_AudienceDefaultEquivalence(t *testing.T) {
	c := &LagoonConfig{}
	resp, err := c.Diff(context.Background(), infer.DiffRequest[LagoonConfig, LagoonConfig]{
		Inputs: LagoonConfig{JWTAudience: ""},
		State:  LagoonConfig{JWTAudience: "api.dev"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.HasChanges {
		t.Error("empty string and 'api.dev' are runtime-equivalent; expected no changes")
	}
}

func TestDiff_PartialChange(t *testing.T) {
	c := &LagoonConfig{}
	resp, err := c.Diff(context.Background(), infer.DiffRequest[LagoonConfig, LagoonConfig]{
		Inputs: LagoonConfig{APIUrl: "https://api.test/graphql", Token: "new-tok"},
		State:  LagoonConfig{APIUrl: "https://api.test/graphql", Token: "old-tok"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.HasChanges {
		t.Error("expected changes")
	}
	if len(resp.DetailedDiff) != 1 {
		t.Errorf("expected 1 changed field, got %d", len(resp.DetailedDiff))
	}
	if _, ok := resp.DetailedDiff["token"]; !ok {
		t.Error("expected 'token' in DetailedDiff")
	}
}
