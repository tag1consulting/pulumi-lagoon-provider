package config

import (
	"context"
	"fmt"
	"os"
	"strings"
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
	APIUrl string `pulumi:"apiUrl,optional" provider:"secret"`

	// Token is a pre-configured JWT authentication token.
	Token string `pulumi:"token,optional" provider:"secret"`

	// JWTSecret is the Lagoon core JWTSECRET for generating admin tokens.
	JWTSecret string `pulumi:"jwtSecret,optional" provider:"secret"`

	// JWTAudience sets the `aud` claim for generated JWT tokens.
	// Defaults to "api.dev".
	JWTAudience string `pulumi:"jwtAudience,optional"`

	// Insecure disables SSL certificate verification.
	Insecure bool `pulumi:"insecure,optional"`
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

	return nil
}

// NewClient creates a configured Lagoon API client from this config.
func (c *LagoonConfig) NewClient() *client.Client {
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
func (c *LagoonConfig) Diff(_ context.Context, req infer.DiffRequest[LagoonConfig, LagoonConfig]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	normalizeAudience := func(v string) string {
		v = strings.TrimSpace(v)
		if v == "" {
			return "api.dev"
		}
		return v
	}
	if strings.TrimSpace(req.Inputs.APIUrl) != strings.TrimSpace(req.State.APIUrl) {
		diff["apiUrl"] = p.PropertyDiff{Kind: p.Update}
	}
	if strings.TrimSpace(req.Inputs.Token) != strings.TrimSpace(req.State.Token) {
		diff["token"] = p.PropertyDiff{Kind: p.Update}
	}
	if strings.TrimSpace(req.Inputs.JWTSecret) != strings.TrimSpace(req.State.JWTSecret) {
		diff["jwtSecret"] = p.PropertyDiff{Kind: p.Update}
	}
	if normalizeAudience(req.Inputs.JWTAudience) != normalizeAudience(req.State.JWTAudience) {
		diff["jwtAudience"] = p.PropertyDiff{Kind: p.Update}
	}
	if req.Inputs.Insecure != req.State.Insecure {
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
