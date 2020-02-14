package auth

import (
	"crypto/rsa"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	v1 "github.com/stackrox/infra/generated/api/v1"
)

// createHumanUser synthesizes a v1.User struct from an auth0Claims struct.
func createHumanUser(profile auth0Claims) *v1.User {
	return &v1.User{
		Name:    profile.Name,
		Email:   profile.Email,
		Picture: profile.PictureURL,
		Expiry:  &timestamp.Timestamp{Seconds: profile.ExpiresAt},
	}
}

// stateTokenizer facilitates the generation and verification of oauth2 state
// tokens.
//
// These tokens are generated when a user begins the login flow, and are
// verified when the user returns from the external auth provider. Verification
// ensures that the returning login flow originated here.
//
// The generated token does not contain data outside of an expiration date.
type stateTokenizer struct {
	lifetime time.Duration
	secret   []byte
}

// NewStateTokenizer creates a new stateTokenizer that can generate and verify
// tokens using the given lifetime and signed with the given secret.
func NewStateTokenizer(lifetime time.Duration, secret string) *stateTokenizer {
	return &stateTokenizer{
		lifetime: lifetime,
		secret:   []byte(secret),
	}
}

// Generate generates a state JWT.
func (t stateTokenizer) Generate() (string, error) {
	now := time.Now()
	claims := jwt.StandardClaims{
		ExpiresAt: now.Add(t.lifetime).Unix(),
		NotBefore: now.Unix(),
		IssuedAt:  now.Unix(),
	}

	// Generate new token object, containing the wrapped data.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string using the secret
	return token.SignedString(t.secret)
}

// Validate validates a state JWT.
func (t stateTokenizer) Validate(token string) error {
	_, err := jwt.Parse(token, func(_ *jwt.Token) (interface{}, error) {
		return t.secret, nil
	})
	return err
}

// auth0Tokenizer facilitates the verification of user tokens generated by
// Auth0.
//
// These tokens are generated by Auth0 after a successful user login, and are
// verified when the user returns from the external auth provider. Verification
// ensures that the user data was legitimately generated by Auth0.
//
// The token contains an expiration date and user profile data provided by
// Auth0.
type auth0Tokenizer struct {
	lifetime time.Duration
	key      *rsa.PublicKey
}

// NewAuth0Tokenizer creates a new auth0Tokenizer that can verify Auth0
// generated tokens signed with the given public key.
func NewAuth0Tokenizer(lifetime time.Duration, publicKey *rsa.PublicKey) *auth0Tokenizer {
	return &auth0Tokenizer{
		lifetime: lifetime,
		key:      publicKey,
	}
}

// auth0Claims facilitates the unmarshalling of JWTs containing Auth0 user
// profile data.
type auth0Claims struct {
	FamilyName    string `json:"family_name"`
	GivenName     string `json:"given_name"`
	Name          string `json:"name"`
	Nickname      string `json:"nickname"`
	PictureURL    string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	jwt.StandardClaims
}

// Valid imposes additional validity constraints on Auth0 user profile data.
// Specifically, it enforces users to have verified email addresses and that
// those email addresses are from the @stackrox.com domain.
func (c auth0Claims) Valid() error {
	switch {
	case !c.EmailVerified:
		return errors.New("email address is not verified")
	case !strings.HasSuffix(c.Email, "@stackrox.com"):
		return errors.New("email address does not belong to StackRox")
	default:
		return c.StandardClaims.Valid()
	}
}

// Validate validates an Auth0 JWT and returns a synthesized v1.User struct.
func (t auth0Tokenizer) Validate(token string) (*v1.User, error) {
	var claims auth0Claims
	if _, err := jwt.ParseWithClaims(token, &claims, func(_ *jwt.Token) (interface{}, error) {
		return t.key, nil
	}); err != nil {
		return nil, err
	}

	return createHumanUser(claims), nil
}

// userTokenizer facilitates the generation and verification of user tokens.
//
// These tokens are generated at the end of a successful user login, and are
// verified during every api call. Verification ensures that the user
// data/session hasn't expired.
//
// The generated token contains an expiration date and a v1.User struct.
type userTokenizer struct {
	lifetime time.Duration
	secret   []byte
}

// NewUserTokenizer creates a new userTokenizer that can generate and verify
// tokens using the given lifetime and signed with the given secret.
func NewUserTokenizer(lifetime time.Duration, secret string) *userTokenizer {
	return &userTokenizer{
		lifetime: lifetime,
		secret:   []byte(secret),
	}
}

// userClaims facilitates the arshalling/unmarshalling of JWTs containing v1
// .User data.
type userClaims struct {
	User v1.User `json:"user"`
	jwt.StandardClaims
}

// Generate generates a user JWT containing a v1.User struct.
func (t userTokenizer) Generate(user *v1.User) (string, error) {
	now := time.Now()
	claims := userClaims{
		User: *user,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: now.Add(t.lifetime).Unix(),
			NotBefore: now.Unix(),
			IssuedAt:  now.Unix(),
		},
	}

	// Generate new token object, containing the wrapped data.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string using the secret
	return token.SignedString(t.secret)
}

// Validate validates a user JWT and returns the contained v1.User struct.
func (t userTokenizer) Validate(token string) (*v1.User, error) {
	var claims userClaims
	if _, err := jwt.ParseWithClaims(token, &claims, func(_ *jwt.Token) (interface{}, error) {
		return t.secret, nil
	}); err != nil {
		return nil, err
	}
	return &claims.User, nil
}
