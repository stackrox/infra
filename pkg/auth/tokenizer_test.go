package auth

import (
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4/jwt"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testSecret must be at least 32 bytes, the minimum key size for HS256.
const (
	testSecret    = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	foreignSecret = "0123456789012345678901234567890123456789012345678901234567890123"
)

// newTestSvcacct returns a fresh v1.ServiceAccount. It must not be copied by
// value, as protobuf message structs contain a sync.Mutex.
func newTestSvcacct() *v1.ServiceAccount {
	return &v1.ServiceAccount{
		Name:        "Test Service Account",
		Description: "Test service account",
		Email:       "test@redhat.com",
	}
}

// expiredClaims returns JWT claims that were issued an hour ago and expired a
// minute ago.
func expiredClaims() jwt.Claims {
	return jwt.Claims{
		Expiry:    jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		NotBefore: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Hour)),
	}
}

// notYetValidClaims returns JWT claims that were issued now and become valid in
// an hour.
func notYetValidClaims() jwt.Claims {
	return jwt.Claims{
		Expiry:    jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
		NotBefore: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
}

func TestStateTokenizerValidate(t *testing.T) {
	tokenizer := NewStateTokenizer(time.Hour, testSecret)

	validToken, err := tokenizer.Generate()
	require.NoError(t, err)

	expiredToken, err := signedToken([]byte(testSecret), expiredClaims())
	require.NoError(t, err)

	notYetValidToken, err := signedToken([]byte(testSecret), notYetValidClaims())
	require.NoError(t, err)

	// Signed with a different secret than the tokenizer was constructed with.
	foreignToken, err := signedToken([]byte(foreignSecret), expiredClaims())
	require.NoError(t, err)

	testCases := map[string]struct {
		token    string
		expError bool
	}{
		"valid":            {token: validToken, expError: false},
		"expired":          {token: expiredToken, expError: true},
		"not-yet-valid":    {token: notYetValidToken, expError: true},
		"foreign-secret":   {token: foreignToken, expError: true},
		"unsigned-garbage": {token: "not-a-jwt", expError: true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tokenizer.Validate(tc.token)
			if tc.expError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserTokenizerValidate(t *testing.T) {
	tokenizer := NewUserTokenizer(time.Hour, testSecret)

	user := &v1.User{
		Name:  "Test User",
		Email: "test@redhat.com",
	}

	validToken, err := tokenizer.Generate(user)
	require.NoError(t, err)

	expiredToken, err := signedToken([]byte(testSecret), userClaims{
		User:   user,
		Claims: expiredClaims(),
	})
	require.NoError(t, err)

	notYetValidToken, err := signedToken([]byte(testSecret), userClaims{
		User:   user,
		Claims: notYetValidClaims(),
	})
	require.NoError(t, err)

	foreignToken, err := signedToken([]byte(foreignSecret), userClaims{
		User:   user,
		Claims: expiredClaims(),
	})
	require.NoError(t, err)

	testCases := map[string]struct {
		token    string
		expError bool
	}{
		"valid":            {token: validToken, expError: false},
		"expired":          {token: expiredToken, expError: true},
		"not-yet-valid":    {token: notYetValidToken, expError: true},
		"foreign-secret":   {token: foreignToken, expError: true},
		"unsigned-garbage": {token: "not-a-jwt", expError: true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			validatedUser, err := tokenizer.Validate(tc.token)
			if tc.expError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, user, validatedUser)
			}
		})
	}
}

func TestServiceAccountTokenizerValidate(t *testing.T) {
	tokenizer := serviceAccountTokenizer{
		lifetime: time.Hour,
		secret:   []byte(testSecret),
	}

	svcacct := newTestSvcacct()
	validToken, err := tokenizer.Generate(svcacct)
	require.NoError(t, err)

	// Craft a token whose embedded service account has an ExpiresAt in the
	// past; this is the regression test for expired tokens still being
	// accepted by Validate.
	expiredSvcacct := newTestSvcacct()
	expiredSvcacct.ExpiresAt = time.Now().Add(-time.Minute).Unix()
	expiredSvcacct.NotBefore = time.Now().Add(-time.Hour).Unix()
	expiredSvcacct.IssuedAt = time.Now().Add(-time.Hour).Unix()
	expiredToken, err := signedToken([]byte(testSecret), serviceAccountValidator{expiredSvcacct})
	require.NoError(t, err)

	notYetValidSvcacct := newTestSvcacct()
	notYetValidSvcacct.ExpiresAt = time.Now().Add(2 * time.Hour).Unix()
	notYetValidSvcacct.NotBefore = time.Now().Add(time.Hour).Unix()
	notYetValidSvcacct.IssuedAt = time.Now().Unix()
	notYetValidToken, err := signedToken([]byte(testSecret), serviceAccountValidator{notYetValidSvcacct})
	require.NoError(t, err)

	foreignToken, err := signedToken([]byte(foreignSecret), serviceAccountValidator{expiredSvcacct})
	require.NoError(t, err)

	testCases := map[string]struct {
		token    string
		expError bool
	}{
		"valid":            {token: validToken, expError: false},
		"expired":          {token: expiredToken, expError: true},
		"not-yet-valid":    {token: notYetValidToken, expError: true},
		"foreign-secret":   {token: foreignToken, expError: true},
		"unsigned-garbage": {token: "not-a-jwt", expError: true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			validatedSvcacct, err := tokenizer.Validate(tc.token)
			if tc.expError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, newTestSvcacct().Name, validatedSvcacct.Name)
				assert.Equal(t, newTestSvcacct().Email, validatedSvcacct.Email)
			}
		})
	}
}
