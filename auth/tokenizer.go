package auth

import (
	"context"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"github.com/stackrox/infra/auth/claimrule"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/logging"
	"golang.org/x/oauth2"
)

const (
	// clockDriftLeeway is used to account for minor clock drift between our host and OIDC.
	clockDriftLeeway = 10 * time.Second

	emailSuffixRedHat = "@redhat.com"
)

var excludedEmails = map[string]bool{"infra@stackrox.com": true}

// createHumanUser synthesizes a v1.User struct from an oidcClaims struct.
func createHumanUser(profile oidcClaims) *v1.User {
	return &v1.User{
		Name:    profile.Name,
		Email:   profile.Email,
		Picture: profile.PictureURL,
		Expiry:  &timestamp.Timestamp{Seconds: int64(*profile.Expiry)},
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
	nowDate := jwt.NewNumericDate(now)
	claims := jwt.Claims{
		Expiry:    jwt.NewNumericDate(now.Add(t.lifetime)),
		NotBefore: nowDate,
		IssuedAt:  nowDate,
	}
	return signedToken(t.secret, claims)
}

// Validate validates a state JWT.
func (t stateTokenizer) Validate(token string) error {
	_, err := jwt.ParseSigned(token, []jose.SignatureAlgorithm{jose.HS256})
	return err
}

// oidcTokenizer facilitates the verification of user tokens generated by an
// OIDC provider.
//
// These tokens are generated by the OIDC provider after successful user login.
// They are verified when the user returns from the external auth provider.
// Verification ensures that the user data was legitimately generated by the
// OIDC provider.
//
// The token contains an expiration date and user profile data provided by the
// OIDC provider.
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
	jwt.Claims
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
	_, isExcluded := excludedEmails[c.Email]
	if isExcluded {
		errMsg := "email address is excluded"
		log.AuditLog(logging.INFO, "oidc-claim-validation", errMsg, "email", c.Email)
		return errors.New(errMsg)
	}
	switch {
	case !c.EmailVerified:
		errMsg := "email address is not verified"
		log.AuditLog(logging.INFO, "oidc-claim-validation", errMsg, "email", c.Email)
		return errors.New(errMsg)
	case !strings.HasSuffix(c.Email, emailSuffixRedHat):
		errMsg := "email address does not belong to Red Hat"
		log.AuditLog(logging.INFO, "oidc-claim-validation", errMsg, "email", c.Email)
		return errors.Errorf(errMsg)
	default:
		// Use an empty jwt.Expected to skip non-time-related validation and use time.Now()
		// for the validation.
		return c.ValidateWithLeeway(jwt.Expected{}, clockDriftLeeway)
	}
}

// Validate validates a JWT Token and returns a synthesized v1.User struct.
func (t oidcTokenizer) Validate(ctx context.Context, rawToken *oauth2.Token) (*v1.User, error) {
	rawIDToken := rawToken.Extra("id_token").(string)

	idToken, errVerify := t.verifier.Verify(ctx, rawIDToken)
	if errVerify != nil {
		return nil, errVerify
	}

	var claims oidcClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}

	if rawAccessToken := rawToken.Extra("access_token"); idToken.AccessTokenHash != "" && rawAccessToken != nil {
		if err := idToken.VerifyAccessToken(rawAccessToken.(string)); err != nil {
			return nil, err
		}
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
	jwt.Claims
	User v1.User `json:"user"`
}

// Generate generates a user JWT containing a v1.User struct.
func (t userTokenizer) Generate(user *v1.User) (string, error) {
	now := time.Now()
	nowDate := jwt.NewNumericDate(now)
	claims := userClaims{
		User: *user,
		Claims: jwt.Claims{
			Expiry:    jwt.NewNumericDate(now.Add(t.lifetime)),
			NotBefore: nowDate,
			IssuedAt:  nowDate,
		},
	}
	return signedToken(t.secret, claims)
}

// Validate validates a user JWT and returns the contained v1.User struct.
func (t userTokenizer) Validate(token string) (*v1.User, error) {
	parsedToken, err := jwt.ParseSigned(token, []jose.SignatureAlgorithm{jose.HS256})
	if err != nil {
		return nil, err
	}

	var claims userClaims
	err = parsedToken.Claims(t.secret, &claims)
	if err != nil {
		return nil, err
	}

	return &claims.User, nil
}

type serviceAccountValidator v1.ServiceAccount

func (s serviceAccountValidator) Valid() error {
	_, isExcluded := excludedEmails[s.Email]
	if isExcluded {
		return errors.New("email address is excluded")
	}

	now := time.Now().Unix()

	switch {
	case s.ExpiresAt < now:
		return errors.New("token expired")
	case s.NotBefore > now:
		return errors.New("token not yet valid")
	case s.IssuedAt > now:
		return errors.New("token issued in the future")
	case s.Name == "":
		return errors.New("name was empty")
	case s.Description == "":
		return errors.New("description was empty")
	case !strings.HasSuffix(s.Email, emailSuffixRedHat):
		return errors.Errorf("%q is not a Red Hat address", s.Email)
	default:
		return nil
	}
}

type serviceAccountTokenizer struct {
	secret   []byte
	lifetime time.Duration
}

// Generate generates a service account JWT containing a v1.ServiceAccount.
func (t serviceAccountTokenizer) Generate(svcacct v1.ServiceAccount) (string, error) {
	// Set issuing and expiration times on new ServiceAccount.
	now := time.Now()
	svcacct.ExpiresAt = now.Add(t.lifetime).Unix()
	svcacct.NotBefore = now.Unix()
	svcacct.IssuedAt = now.Unix()

	svc := serviceAccountValidator(svcacct)

	// Ensure that our service account is well-formed.
	if err := svc.Valid(); err != nil {
		log.AuditLog(logging.INFO, "service-account-validation", err.Error(), "email", svc.Email)
		return "", errors.Wrap(err, "invalid service account")
	}

	return signedToken(t.secret, svc)
}

// Validate validates a service account JWT and returns the contained
// v1.ServiceAccount.
func (t serviceAccountTokenizer) Validate(token string) (v1.ServiceAccount, error) {
	parsedToken, err := jwt.ParseSigned(token, []jose.SignatureAlgorithm{jose.HS256})
	if err != nil {
		return v1.ServiceAccount{}, err
	}

	var claims serviceAccountValidator
	err = parsedToken.Claims(t.secret, &claims)
	if err != nil {
		return v1.ServiceAccount{}, err
	}

	return v1.ServiceAccount(claims), nil
}

// accessTokenizer facilitates the verification of access roles generated by
// OIDC provider.
type accessTokenizer struct {
	claimRules *claimrule.ClaimRules
}

// NewAccessTokenizer creates a new tokenizer that can verify OIDC provider
// generated Access Token.
func NewAccessTokenizer(claimRules *claimrule.ClaimRules) *accessTokenizer {
	return &accessTokenizer{
		claimRules: claimRules,
	}
}

// Validate validates all specified claim rules for Access Token.
func (t accessTokenizer) Validate(_ context.Context, rawToken *oauth2.Token) error {
	if t.claimRules.IsEmpty() {
		return nil
	}

	rawAccessToken := rawToken.Extra("access_token")
	if rawAccessToken == nil {
		return errors.Errorf("access token is not available")
	}

	return t.claimRules.Validate(rawAccessToken.(string))
}

func signedToken(key []byte, claims any) (string, error) {
	sigKey := jose.SigningKey{
		Algorithm: jose.HS256,
		Key:       key,
	}
	// See https://pkg.go.dev/github.com/go-jose/go-jose/v4/jwt#example-Signed for an example.
	sig, err := jose.NewSigner(sigKey, (&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		return "", err
	}

	return jwt.Signed(sig).Claims(claims).Serialize()
}
