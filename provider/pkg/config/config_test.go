package config

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
)

// requireUnsetEnv removes the named environment variables for the duration
// of the test and restores their original values on test cleanup. Errors
// from os.Unsetenv and os.Setenv are surfaced via t.Fatalf / t.Errorf
// rather than silently dropped, so a failure to mutate the environment
// does not corrupt subsequent tests.
//
// Use this in place of bare os.Unsetenv + defer save/restore for tests
// that need a variable absent. For tests that only need a variable set
// to a specific value, prefer t.Setenv directly (it already auto-restores
// on cleanup).
func requireUnsetEnv(t *testing.T, keys ...string) {
	t.Helper()
	for _, k := range keys {
		if v, ok := os.LookupEnv(k); ok {
			key, val := k, v
			t.Cleanup(func() {
				if err := os.Setenv(key, val); err != nil {
					t.Errorf("restore %s: %v", key, err)
				}
			})
		}
		if err := os.Unsetenv(k); err != nil {
			t.Fatalf("unset %s: %v", k, err)
		}
	}
}

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
	requireUnsetEnv(t, "LAGOON_TOKEN", "LAGOON_JWT_SECRET")

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
	requireUnsetEnv(t, "LAGOON_JWT_SECRET")
	t.Setenv("LAGOON_TOKEN", "env-token")

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
	requireUnsetEnv(t, "LAGOON_TOKEN")
	t.Setenv("LAGOON_JWT_SECRET", "env-jwt-secret")

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
	requireUnsetEnv(t, "LAGOON_TOKEN", "LAGOON_JWT_SECRET")

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
	requireUnsetEnv(t, "LAGOON_JWT_SECRET")
	t.Setenv("LAGOON_TOKEN", "  env-token\n")

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
	requireUnsetEnv(t, "LAGOON_TOKEN", "LAGOON_JWT_SECRET")

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
	resp, err := c.Diff(context.Background(), infer.DiffRequest[*LagoonConfig, *LagoonConfig]{
		Inputs: &LagoonConfig{APIUrl: "https://api.test/graphql", Token: "tok", JWTAudience: "api.dev"},
		State:  &LagoonConfig{APIUrl: "https://api.test/graphql", Token: "tok", JWTAudience: "api.dev"},
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
	resp, err := c.Diff(context.Background(), infer.DiffRequest[*LagoonConfig, *LagoonConfig]{
		Inputs: &LagoonConfig{Token: "my-token\n", JWTSecret: " secret "},
		State:  &LagoonConfig{Token: "my-token", JWTSecret: "secret"},
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
	resp, err := c.Diff(context.Background(), infer.DiffRequest[*LagoonConfig, *LagoonConfig]{
		Inputs: &LagoonConfig{APIUrl: "https://new-api.test/graphql", Token: "new-tok", JWTSecret: "new-secret", JWTAudience: "api.prod", Insecure: true},
		State:  &LagoonConfig{APIUrl: "https://old-api.test/graphql", Token: "old-tok", JWTSecret: "old-secret", JWTAudience: "api.dev", Insecure: false},
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
	// "token" is intentionally absent: when jwtSecret is present on both sides,
	// Diff skips the token comparison because the token is derived and changes
	// every run. Only jwtSecret itself is compared.
	expectedFields := []string{"apiUrl", "jwtSecret", "jwtAudience", "insecure"}
	for _, f := range expectedFields {
		if _, ok := resp.DetailedDiff[f]; !ok {
			t.Errorf("expected %q in DetailedDiff", f)
		}
	}
	if _, ok := resp.DetailedDiff["token"]; ok {
		t.Error("expected 'token' to be absent from DetailedDiff when jwtSecret is present")
	}
}

func TestDiff_AudienceDefaultEquivalence(t *testing.T) {
	c := &LagoonConfig{}
	resp, err := c.Diff(context.Background(), infer.DiffRequest[*LagoonConfig, *LagoonConfig]{
		Inputs: &LagoonConfig{JWTAudience: ""},
		State:  &LagoonConfig{JWTAudience: "api.dev"},
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
	resp, err := c.Diff(context.Background(), infer.DiffRequest[*LagoonConfig, *LagoonConfig]{
		Inputs: &LagoonConfig{APIUrl: "https://api.test/graphql", Token: "new-tok"},
		State:  &LagoonConfig{APIUrl: "https://api.test/graphql", Token: "old-tok"},
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

func TestNewClient_ReusesClient(t *testing.T) {
	for _, k := range []string{"LAGOON_TOKEN", "LAGOON_JWT_SECRET", "LAGOON_API_URL", "LAGOON_JWT_AUDIENCE", "LAGOON_INSECURE"} {
		t.Setenv(k, "")
	}
	t.Setenv("LAGOON_TOKEN", "test-token")
	cfg := &LagoonConfig{APIUrl: "https://api.example.com/graphql"}
	if err := cfg.Configure(context.TODO()); err != nil {
		t.Fatalf("Configure failed: %v", err)
	}
	c1 := cfg.NewClient()
	c2 := cfg.NewClient()
	if c1 != c2 {
		t.Error("expected NewClient to return the same client instance on repeated calls")
	}
}

func TestNewClient_NilHolder_CreatesNewClient(t *testing.T) {
	// Without Configure (no clientHolder), NewClient still works — it just
	// creates a fresh client each time rather than reusing one.
	cfg := &LagoonConfig{APIUrl: "https://api.example.com/graphql", Token: "tok"}
	c1 := cfg.NewClient()
	c2 := cfg.NewClient()
	if c1 == nil || c2 == nil {
		t.Error("expected non-nil clients")
	}
	// Not the same pointer — no caching without Configure
	if c1 == c2 {
		t.Error("expected distinct clients when clientHolder is nil")
	}
}

// TestNewClient_ConcurrentStructCopies_ShareSameClient reproduces the scenario
// that crashes DeployTarget.Create (and any other resource) under concurrent
// Pulumi resource operations: infer.GetConfig returns independent value
// copies of LagoonConfig for each operation. Regression test for the bug
// where cachedClient was a plain *client.Client field — shared clientOnce
// meant sync.Once.Do ran exactly once total, but only the copy that won the
// race ever had its cachedClient field populated; every other copy's
// NewClient() returned nil, and callers dereferencing that nil *Client
// panicked. See GitHub issue #265.
func TestNewClient_ConcurrentStructCopies_ShareSameClient(t *testing.T) {
	for _, k := range []string{"LAGOON_TOKEN", "LAGOON_JWT_SECRET", "LAGOON_API_URL", "LAGOON_JWT_AUDIENCE", "LAGOON_INSECURE"} {
		t.Setenv(k, "")
	}
	t.Setenv("LAGOON_TOKEN", "test-token")
	cfg := &LagoonConfig{APIUrl: "https://api.example.com/graphql"}
	if err := cfg.Configure(context.TODO()); err != nil {
		t.Fatalf("Configure failed: %v", err)
	}

	const goroutines = 20
	clients := make([]*client.Client, goroutines)
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// infer.GetConfig returns a value copy of the config struct on
			// every call; simulate that here by copying cfg by value before
			// calling NewClient, exactly as concurrent resource Create/Update
			// calls would each observe their own copy.
			copyOfCfg := *cfg
			clients[i] = copyOfCfg.NewClient()
		}(i)
	}
	wg.Wait()

	for i, c := range clients {
		if c == nil {
			t.Fatalf("goroutine %d got a nil client", i)
		}
		if c != clients[0] {
			t.Errorf("goroutine %d got a different client instance than goroutine 0; expected all struct copies to share the same cached client", i)
		}
	}
}

func TestDiff_JWTSecret_TokenNotDiffed(t *testing.T) {
	// When jwtSecret is the auth source, the derived token changes every run.
	// Diff must not report a token change to avoid spurious provider updates.
	c := &LagoonConfig{}
	resp, err := c.Diff(context.Background(), infer.DiffRequest[*LagoonConfig, *LagoonConfig]{
		Inputs: &LagoonConfig{JWTSecret: "my-secret", Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.new"},
		State:  &LagoonConfig{JWTSecret: "my-secret", Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.old"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.HasChanges {
		t.Errorf("expected no changes when only derived token differs (same jwtSecret); got diff: %v", resp.DetailedDiff)
	}
	if _, ok := resp.DetailedDiff["token"]; ok {
		t.Error("expected 'token' to be absent from DetailedDiff when jwtSecret is the auth source")
	}
}

func TestDiff_JWTSecretChanged_Detected(t *testing.T) {
	// Changing jwtSecret itself must still be detected.
	c := &LagoonConfig{}
	resp, err := c.Diff(context.Background(), infer.DiffRequest[*LagoonConfig, *LagoonConfig]{
		Inputs: &LagoonConfig{JWTSecret: "new-secret"},
		State:  &LagoonConfig{JWTSecret: "old-secret"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.HasChanges {
		t.Error("expected changes when jwtSecret changes")
	}
	if _, ok := resp.DetailedDiff["jwtSecret"]; !ok {
		t.Error("expected 'jwtSecret' in DetailedDiff")
	}
	if _, ok := resp.DetailedDiff["token"]; ok {
		t.Error("expected 'token' to be absent from DetailedDiff even when jwtSecret changes")
	}
}

func TestDiff_ExplicitToken_NoDiffWhenUnchanged(t *testing.T) {
	// When no jwtSecret, explicit token changes are still detected normally.
	c := &LagoonConfig{}
	resp, err := c.Diff(context.Background(), infer.DiffRequest[*LagoonConfig, *LagoonConfig]{
		Inputs: &LagoonConfig{Token: "tok-v2"},
		State:  &LagoonConfig{Token: "tok-v1"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.HasChanges {
		t.Error("expected changes when explicit token changes without jwtSecret")
	}
	if _, ok := resp.DetailedDiff["token"]; !ok {
		t.Error("expected 'token' in DetailedDiff for explicit token change")
	}
}

// ==================== Diff dispatch through the real framework (issue #267) ====================

// dummyArgs/dummyState/dummyResource exist only so infer.NewProviderBuilder has
// at least one resource to validate against (infer.Config alone is not enough
// to build a provider). They are not exercised by the assertions below.
type dummyArgs struct{}
type dummyState struct{ dummyArgs }

type dummyResource struct{}

func (dummyResource) Create(_ context.Context, req infer.CreateRequest[dummyArgs]) (infer.CreateResponse[dummyState], error) {
	return infer.CreateResponse[dummyState]{ID: "dummy", Output: dummyState{req.Inputs}}, nil
}

// TestDiffConfig_DispatchesToCustomDiff is a regression test for issue #267: the
// provider is wired exactly as production does it (infer.Config(&LagoonConfig{})
// registers the pointer type, not the value type), and the resulting DiffConfig
// RPC entry point is called directly - the same path the Pulumi engine calls on
// every refresh/preview/up. Before the fix, LagoonConfig.Diff was declared with
// infer.DiffRequest[LagoonConfig, LagoonConfig] (value type parameters), which
// does not satisfy the CustomDiff[*LagoonConfig, *LagoonConfig] interface the
// framework actually checks for when the config is registered as a pointer.
// That mismatch silently fell through to infer's default diffing, which treats
// any changed field (including the JWT-derived token, expected to change every
// run) as forcing a full provider replace. Calling c.Diff(...) directly, as the
// other tests in this file do, cannot catch this: it bypasses the framework's
// interface-satisfaction check entirely and always succeeds. Only a real
// DiffConfig RPC call through a built provider exercises the actual bug.
func TestDiffConfig_DispatchesToCustomDiff(t *testing.T) {
	// Without the fix, DiffConfig falls through to infer's default diffing
	// path, which calls GetSchema and needs provider.RunInfo from the
	// context - context this minimal test harness doesn't provide, so the
	// fallback path panics instead of just returning a wrong-but-clean
	// result. Recover so a regression here reports as a normal test
	// failure rather than crashing the whole test binary.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("DiffConfig panicked: %v (this happens when LagoonConfig.Diff is not dispatched by the "+
				"framework and infer's default diffing fallback runs instead - see issue #267)", r)
		}
	}()

	prov, err := infer.NewProviderBuilder().
		WithResources(infer.Resource(&dummyResource{})).
		WithConfig(infer.Config(&LagoonConfig{})).
		Build()
	if err != nil {
		t.Fatalf("failed to build provider: %v", err)
	}
	if prov.DiffConfig == nil {
		t.Fatal("provider.DiffConfig is nil; infer.Config did not wire up DiffConfig")
	}

	sameSecret := property.New("same-secret")
	req := p.DiffRequest{
		Urn: "urn:pulumi:test::test::pulumi:providers:lagoon::lagoon-provider",
		State: property.NewMap(map[string]property.Value{
			"apiUrl":    property.New("https://api.test/graphql"),
			"jwtSecret": sameSecret,
			"token":     property.New("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.old"),
		}),
		Inputs: property.NewMap(map[string]property.Value{
			"apiUrl":    property.New("https://api.test/graphql"),
			"jwtSecret": sameSecret,
			"token":     property.New("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.new"),
		}),
	}

	resp, err := prov.DiffConfig(context.Background(), req)
	if err != nil {
		t.Fatalf("DiffConfig returned an error: %v", err)
	}
	for field, d := range resp.DetailedDiff {
		if d.Kind == p.UpdateReplace || d.Kind == p.AddReplace || d.Kind == p.DeleteReplace {
			t.Errorf("DiffConfig marked field %q as %v for a token-only change with identical jwtSecret on "+
				"both sides; this means LagoonConfig.Diff was not dispatched by the framework and infer's "+
				"default force-replace-on-any-change fallback ran instead (issue #267). Full DetailedDiff: %+v",
				field, d.Kind, resp.DetailedDiff)
		}
	}
	if _, ok := resp.DetailedDiff["token"]; ok {
		t.Errorf("expected 'token' to be absent from DetailedDiff when jwtSecret is unchanged; got: %+v", resp.DetailedDiff)
	}
}
