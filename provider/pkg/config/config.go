package config

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	p "github.com/pulumi/pulumi-go-provider"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
)

// LagoonConfig holds the provider configuration.
// All resources access this via infer.GetConfig[LagoonConfig](ctx).
type LagoonConfig struct {
	// APIUrl is the Lagoon GraphQL API endpoint.
	APIUrl string `pulumi:"apiUrl,optional"`

	// Token is a pre-configured JWT authentication token.
	Token string `pulumi:"token,optional" provider:"secret"`

	// JWTSecret is the Lagoon core JWTSECRET for generating admin tokens.
	JWTSecret string `pulumi:"jwtSecret,optional" provider:"secret"`

	// JWTAudience sets the `aud` claim for generated JWT tokens.
	// Defaults to "api.dev".
	JWTAudience string `pulumi:"jwtAudience,optional"`

	// Insecure disables SSL certificate verification.
	Insecure bool `pulumi:"insecure,optional"`

	// clientHolder holds the single shared Client instance for this provider
	// instance. It is a pointer to a holder struct (not the sync.Once/client
	// fields directly) so that every copy of LagoonConfig produced by
	// infer.GetConfig shares the same underlying Once and client slot; a
	// *sync.Once field alone shares the Once but each copy would still have
	// its own independent *client.Client field, so only the copy that wins
	// the Once race would ever see a non-nil client. It is initialized by
	// Configure and nil until then.
	clientHolder *clientHolder
}

// clientHolder is the shared, once-initialized Client instance referenced by
// all copies of a given LagoonConfig.
type clientHolder struct {
	once   sync.Once
	client *client.Client
}

// Annotate provides descriptions and defaults for config fields.
func (c *LagoonConfig) Annotate(a infer.Annotator) {
	a.Describe(&c.APIUrl, "The Lagoon GraphQL API endpoint URL.")
	a.SetDefault(&c.APIUrl, "https://api.lagoon.sh/graphql", "LAGOON_API_URL")

	a.Describe(&c.Token, "A pre-configured JWT authentication token for the Lagoon API.")
	a.SetDefault(&c.Token, nil, "LAGOON_TOKEN")

	a.Describe(&c.JWTSecret, "The Lagoon core JWTSECRET. Used to generate admin tokens on-the-fly.")
	a.SetDefault(&c.JWTSecret, nil, "LAGOON_JWT_SECRET")

	a.Describe(&c.JWTAudience, "The audience claim for generated JWT tokens. Defaults to 'api.dev'.")
	a.SetDefault(&c.JWTAudience, "api.dev", "LAGOON_JWT_AUDIENCE")

	a.Describe(&c.Insecure, "Disable SSL certificate verification when connecting to the Lagoon API.")
	a.SetDefault(&c.Insecure, false, "LAGOON_INSECURE")
}

// Configure validates the config and prepares it for use.
// Called by the Pulumi engine when the provider is configured.
func (c *LagoonConfig) Configure(ctx context.Context) error {
	// p.GetLogger returns a Logger value (not a pointer), and when the
	// context carries no logger key it falls back to a slog-backed sink
	// rather than returning a nil sink. Tests can therefore call
	// Configure with context.TODO() without a nil-deref on Debugf. See
	// pulumi-go-provider/logging.go.
	log := p.GetLogger(ctx)

	c.Token = strings.TrimSpace(c.Token)
	c.JWTSecret = strings.TrimSpace(c.JWTSecret)
	c.JWTAudience = strings.TrimSpace(c.JWTAudience)

	tokenFromSecret := false

	if c.Token == "" && c.JWTSecret != "" {
		log.Debugf("Generating admin token from jwtSecret (%d bytes, audience=%q)", len(c.JWTSecret), c.JWTAudience)
		token, err := c.generateAdminToken()
		if err != nil {
			return fmt.Errorf("failed to generate admin token from jwtSecret: %w", err)
		}
		c.Token = token
		tokenFromSecret = true
	}

	if c.Token == "" {
		if envToken := strings.TrimSpace(os.Getenv("LAGOON_TOKEN")); envToken != "" {
			log.Debugf("Using token from LAGOON_TOKEN environment variable")
			c.Token = envToken
		}
	}
	if c.Token == "" {
		if envSecret := strings.TrimSpace(os.Getenv("LAGOON_JWT_SECRET")); envSecret != "" {
			log.Debugf("Generating admin token from LAGOON_JWT_SECRET (%d bytes, audience=%q)", len(envSecret), c.JWTAudience)
			c.JWTSecret = envSecret
			token, err := generateAdminTokenFromSecret(envSecret, c.JWTAudience)
			if err != nil {
				return fmt.Errorf("failed to generate token from LAGOON_JWT_SECRET: %w", err)
			}
			c.Token = token
			tokenFromSecret = true
		}
	}

	if c.Token == "" {
		return fmt.Errorf("lagoon authentication required: set 'token' or 'jwtSecret' in provider config, " +
			"or use LAGOON_TOKEN/LAGOON_JWT_SECRET environment variables")
	}

	if !tokenFromSecret {
		c.JWTSecret = ""
	}

	c.clientHolder = &clientHolder{}

	return nil
}

// NewClient returns the shared Lagoon API client for this config.
// The client is created once per provider configure call and reused for all
// subsequent resource operations, preserving API-version detection cache
// and token refresh state across operations. All copies of LagoonConfig
// produced by infer.GetConfig after Configure share the same clientHolder
// pointer, so concurrent resource operations correctly observe the one
// cached client instead of racing to populate independent copies of it.
func (c *LagoonConfig) NewClient() *client.Client {
	if c.clientHolder != nil {
		h := c.clientHolder
		h.once.Do(func() {
			h.client = c.newClientUncached()
		})
		return h.client
	}
	return c.newClientUncached()
}

func (c *LagoonConfig) newClientUncached() *client.Client {
	opts := []client.ClientOption{}

	if c.Insecure {
		opts = append(opts, client.WithInsecureSSL())
	}

	// Only enable token refresh when the token was derived from a JWT secret
	// (not when an explicit token was provided via config or LAGOON_TOKEN)
	if c.JWTSecret != "" {
		audience := c.JWTAudience
		secret := c.JWTSecret
		opts = append(opts, client.WithTokenFunc(func() (string, error) {
			return generateAdminTokenFromSecret(secret, audience)
		}))
	}

	return client.NewClient(c.APIUrl, c.Token, opts...)
}

// Diff compares old and new provider configuration. No config change requires a
// provider replace — changing apiUrl, token, jwtSecret, or insecure only changes
// how the provider authenticates or connects, not which resources it can manage.
// Returning Update (never UpdateReplace) prevents the cascading replace of every
// resource associated with this provider instance.
//
// The type parameters are *LagoonConfig, not LagoonConfig: infer.Config is
// called with a pointer (infer.Config(&LagoonConfig{})), which Go's generics
// resolve to T = *LagoonConfig. The framework's internal dispatch then checks
// whether *LagoonConfig implements CustomDiff[*LagoonConfig, *LagoonConfig],
// not CustomDiff[LagoonConfig, LagoonConfig]. A value-typed signature here
// silently fails that check and falls through to infer's default diffing,
// which treats every changed field (including the JWT-derived token, which
// changes on every Configure call by design) as forcing a full provider
// replace. See pulumi-lagoon-provider#267.
func (c *LagoonConfig) Diff(_ context.Context, req infer.DiffRequest[*LagoonConfig, *LagoonConfig]) (infer.DiffResponse, error) {
	inputs, state := req.Inputs, req.State
	if inputs == nil {
		inputs = &LagoonConfig{}
	}
	if state == nil {
		state = &LagoonConfig{}
	}

	diff := map[string]p.PropertyDiff{}
	normalizeAudience := func(v string) string {
		v = strings.TrimSpace(v)
		if v == "" {
			return "api.dev"
		}
		return v
	}
	if strings.TrimSpace(inputs.APIUrl) != strings.TrimSpace(state.APIUrl) {
		diff["apiUrl"] = p.PropertyDiff{Kind: p.Update}
	}
	// Only diff `token` when it is the explicit credential on both sides.
	// When jwtSecret is present, Configure derives a fresh token each time;
	// diffing that derived value produces noise because JWTs change every run.
	inputSecret := strings.TrimSpace(inputs.JWTSecret)
	stateSecret := strings.TrimSpace(state.JWTSecret)
	if inputSecret == "" && stateSecret == "" {
		if strings.TrimSpace(inputs.Token) != strings.TrimSpace(state.Token) {
			diff["token"] = p.PropertyDiff{Kind: p.Update}
		}
	}
	if inputSecret != stateSecret {
		diff["jwtSecret"] = p.PropertyDiff{Kind: p.Update}
	}
	if normalizeAudience(inputs.JWTAudience) != normalizeAudience(state.JWTAudience) {
		diff["jwtAudience"] = p.PropertyDiff{Kind: p.Update}
	}
	if inputs.Insecure != state.Insecure {
		diff["insecure"] = p.PropertyDiff{Kind: p.Update}
	}
	return p.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff}, nil
}

// generateAdminToken creates an admin JWT from the configured secret.
func (c *LagoonConfig) generateAdminToken() (string, error) {
	return generateAdminTokenFromSecret(c.JWTSecret, c.JWTAudience)
}

// generateAdminTokenFromSecret creates an admin JWT token.
func generateAdminTokenFromSecret(jwtSecret, audience string) (string, error) {
	if audience == "" {
		audience = "api.dev"
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"role": "admin",
		"iss":  "lagoon-api",
		"sub":  "lagoonadmin",
		"aud":  audience,
		"iat":  now.Unix(),
		"exp":  now.Add(1 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}
