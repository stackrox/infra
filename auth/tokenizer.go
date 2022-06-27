package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/dgrijalva/jwt-go"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/jeremywohl/flatten"
	"github.com/pkg/errors"
	"github.com/stackrox/infra/config"
	v1 "github.com/stackrox/infra/generated/api/v1"
)

// clockDriftLeeway is used to account for minor clock drift between our host,
// and OIDC.
//
// See this issue for context:
// https://github.com/dgrijalva/jwt-go/issues/314#issuecomment-494585527
const clockDriftLeeway = int64(10 * time.Second)

// createHumanUser synthesizes a v1.User struct from an oidcClaims struct.
func createHumanUser(profile oidcClaims) *v1.User {
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
	secret   []byte
	lifetime time.Duration
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

// oidcTokenizer facilitates the verification of user tokens generated by
// RedHat OIDC provider.
//
// These tokens are generated by RedHat OIDC provider after a successful user
// login, and are verified when the user returns from the external auth provider.
// Verification ensures that the user data was legitimately generated by RedHat
// OIDC provider.
//
// The token contains an expiration date and user profile data provided by
// RedHat OIDC provider.
type oidcTokenizer struct {
	verifier *oidc.IDTokenVerifier
}

// NewOidcTokenizer creates a new tokenizer that can verify OIDC provider
// generated ID Token.
func NewOidcTokenizer(verifier *oidc.IDTokenVerifier) *oidcTokenizer {
	return &oidcTokenizer{
		verifier: verifier,
	}
}

// oidcClaims facilitates the unmarshalling of JWTs containing OIDC user
// profile data.
type oidcClaims struct {
	jwt.StandardClaims
	FamilyName    string `json:"family_name"`
	GivenName     string `json:"given_name"`
	Name          string `json:"name"`
	Nickname      string `json:"nickname"`
	PictureURL    string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

// Valid imposes additional validity constraints on OIDC user profile data.
// Specifically, it enforces users to have verified email addresses and that
// those email addresses are from the allowed domains.
func (c oidcClaims) Valid() error {
	switch {
	case !c.EmailVerified:
		return errors.New("email address is not verified")
	case !(strings.HasSuffix(c.Email, "@stackrox.com") || strings.HasSuffix(c.Email, "@redhat.com")):
		return errors.Errorf("%q email address does not belong to StackRox or Red Hat", c.Email)
	default:
		c.StandardClaims.IssuedAt -= clockDriftLeeway
		valid := c.StandardClaims.Valid()
		c.StandardClaims.IssuedAt += clockDriftLeeway
		return valid
	}
}

// Validate validates a JWT Token and returns a synthesized v1.User struct.
func (t oidcTokenizer) Validate(ctx context.Context, rawIDToken string) (*v1.User, error) {
	token, errVerify := t.verifier.Verify(ctx, rawIDToken)
	if errVerify != nil {
		return nil, errVerify
	}

	var claims oidcClaims
	if err := token.Claims(&claims); err != nil {
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

type serviceAccountValidator v1.ServiceAccount

func (s serviceAccountValidator) Valid() error {
	switch {
	case s.Name == "":
		return errors.New("name was empty")
	case s.Description == "":
		return errors.New("description was empty")
	case !(strings.HasSuffix(s.Email, "@stackrox.com") || strings.HasSuffix(s.Email, "@redhat.com")):
		return errors.Errorf("%q is not a StackRox or Red Hat address", s.Email)
	default:
		return nil
	}
}

type serviceAccountTokenizer struct {
	secret []byte
}

// Generate generates a service account JWT containing a v1.ServiceAccount.
func (t serviceAccountTokenizer) Generate(svcacct v1.ServiceAccount) (string, error) {
	svc := serviceAccountValidator(svcacct)

	// Ensure that our service account is well-formed.
	if err := svc.Valid(); err != nil {
		return "", errors.Wrap(err, "invalid service account")
	}

	// Generate new token object, containing the wrapped data.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, svc)

	// Sign and get the complete encoded token as a string using the secret
	return token.SignedString(t.secret)
}

// Validate validates a service account JWT and returns the contained
// v1.ServiceAccount.
func (t serviceAccountTokenizer) Validate(token string) (v1.ServiceAccount, error) {
	var claims serviceAccountValidator
	if _, err := jwt.ParseWithClaims(token, &claims, func(_ *jwt.Token) (interface{}, error) {
		return t.secret, nil
	}); err != nil {
		return v1.ServiceAccount{}, err
	}

	return v1.ServiceAccount(claims), nil
}

// accessTokenizer facilitates the verification of access roles generated by
// OIDC provider.
type accessTokenizer struct {
	issuer string
	claims []config.ClaimOperation
}

// NewAccessTokenizer creates a new tokenizer that can verify OIDC provider
// generated Access Token.
func NewAccessTokenizer(issuer string, claims []config.ClaimOperation) *accessTokenizer {
	return &accessTokenizer{
		issuer: issuer,
		claims: claims,
	}
}

func decodeAccessToken(accessToken string) (map[string]interface{}, error) {
	s := strings.Split(accessToken, ".")
	if len(s) < 2 {
		return nil, errors.New("jws: invalid token received")
	}

	decoded, err := base64.RawURLEncoding.DecodeString(s[1])
	if err != nil {
		return nil, err
	}

	claims := map[string]interface{}{}
	err = json.NewDecoder(bytes.NewBuffer(decoded)).Decode(&claims)

	return claims, err
}

// Validate validates a JWT Token and returns a synthesized v1.User struct.
func (t accessTokenizer) Validate(ctx context.Context, accessToken string) error {
	claims, err := decodeAccessToken(accessToken)
	if err != nil {
		return err
	}

	flatClaims, errFlatten := flatten.Flatten(claims, "", flatten.DotStyle)
	if errFlatten != nil {
		return errFlatten
	}

	for _, claim := range t.claims {
		authClaimOp := ClaimOperation{claim}

		isValid, errCheck := authClaimOp.Check(flatClaims)
		if errCheck != nil {
			return errCheck
		}

		if !isValid {
			return errors.Errorf("Claim for key %q is not valid", claim.Key)
		}
	}

	return nil
}
