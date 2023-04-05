// Package auth facilitates an OAuth login/logout flow.
package auth

import (
	"fmt"
	"log"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	v1 "github.com/stackrox/infra/generated/proto/api/v1"
	"golang.org/x/oauth2"
)

const (
	tokenCookieNew     = "token=%s"
	tokenCookieExpired = "token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT"
)

// OidcAuth facilitates an Oauth2 login flow via http handlers.
type OidcAuth struct {
	endpoint   string
	provider   *oidc.Provider
	jwtState   *stateTokenizer
	jwtAccess  *accessTokenizer
	jwtOidc    *oidcTokenizer
	jwtUser    *userTokenizer
	conf       *oauth2.Config
	jwtSvcAcct serviceAccountTokenizer
}

// ValidateUser validates a user JWT and returns the contained v1.User struct.
func (a OidcAuth) ValidateUser(token string) (*v1.User, error) {
	return a.jwtUser.Validate(token)
}

// GenerateServiceAccountToken generates a service account JWT containing a
// v1.User struct.
func (a OidcAuth) GenerateServiceAccountToken(svcacct v1.ServiceAccount) (string, error) {
	return a.jwtSvcAcct.Generate(svcacct)
}

// ValidateServiceAccountToken validates a service account JWT and returns the
// contained v1.ServiceAccount struct.
func (a OidcAuth) ValidateServiceAccountToken(token string) (v1.ServiceAccount, error) {
	return a.jwtSvcAcct.Validate(token)
}

// loginHandler handles the login part of an OIDC flow.
//
// A state token is generated and sent along with the redirect to OIDC provider.
func (a OidcAuth) loginHandler(w http.ResponseWriter, r *http.Request) {
	// Generate a new state token.
	stateToken, err := a.jwtState.Generate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	oauth2.SetAuthURLParam("response_mode", "query")
	redirectURL := a.conf.AuthCodeURL(stateToken)

	// Redirect to authorization server so that the user can login externally.
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// callbackHandler handles the callback part of an Oauth2 flow.
//
// After returning from OIDC provider, the state token is verified. A user
// profile is then obtained from OIDC provider that includes details about the
// newly logged-in user. This user information is then stored in a cookie.
func (a OidcAuth) callbackHandler(w http.ResponseWriter, r *http.Request) {
	// Get the value of the "state" HTTP GET param, and validate that it is
	// legitimate.
	stateToken := r.URL.Query().Get("state")
	err := a.jwtState.Validate(stateToken)
	if err != nil {
		log.Printf("[ERROR] Failed to validate state token: %v", err)
		http.Redirect(w, r, "/logout", http.StatusTemporaryRedirect)
		return
	}

	// Get the value of the "code" HTTP GET param, and exchange it for a token.
	code := r.URL.Query().Get("code")
	rawToken, err := a.conf.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("[ERROR] Failed to exchange code: %v", err)
		http.Redirect(w, r, "/logout", http.StatusTemporaryRedirect)
		return
	}

	if err := a.jwtAccess.Validate(r.Context(), rawToken); err != nil {
		log.Printf("[ERROR] Failed to validate Access token: %v", err)
		http.Redirect(w, r, "/logout", http.StatusTemporaryRedirect)
		return
	}

	// Validate that the OIDC idToken profile token is legitimate, and extract a
	// user struct from it.
	user, err := a.jwtOidc.Validate(r.Context(), rawToken)
	if err != nil {
		log.Printf("[ERROR] Failed to validate ID token: %v", err)
		http.Redirect(w, r, "/logout", http.StatusTemporaryRedirect)
		return
	}

	// Generate token containing a user struct.
	userToken, err := a.jwtUser.Generate(user)
	if err != nil {
		log.Printf("[ERROR] Failed to generate user token: %v", err)
		http.Redirect(w, r, "/logout", http.StatusTemporaryRedirect)
		return
	}

	// Persist the user token as a cookie in the user's browser and redirect to
	// a logged in page
	w.Header().Set("set-cookie", fmt.Sprintf(tokenCookieNew, userToken))
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// logoutHandler handles the logout.
//
// The user token cookie is destroyed, and the user is redirected to logout page.
func (a OidcAuth) logoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("set-cookie", tokenCookieExpired)
	http.Redirect(w, r, fmt.Sprintf("https://%s/logout-page.html", a.endpoint), http.StatusTemporaryRedirect)
}

// Handle adds several standard OAuth routes handlers to the given http mux.
func (a OidcAuth) Handle(mux *http.ServeMux) {
	mux.Handle("/callback", http.HandlerFunc(a.callbackHandler))
	mux.Handle("/login", http.HandlerFunc(a.loginHandler))
	mux.Handle("/logout", http.HandlerFunc(a.logoutHandler))
}

// Authorized wraps the given http.Handler in an authorization check. The given
// handler is only called if the user is authorized, otherwise a 404 status
// code is returned.
func (a OidcAuth) Authorized(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the "token" cookie from the request.
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		// Validate the token JWT.
		if _, err := a.jwtUser.Validate(cookie.Value); err != nil {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		// User is authorized, invoke original handler.
		handler.ServeHTTP(w, r)
	})
}

// AuthorizedFunc wraps the given http.HandlerFunc in an authorization check.
// The given handler is only called if the user is authorized, otherwise a 404
// status code is returned.
func (a OidcAuth) AuthorizedFunc(handler http.HandlerFunc) http.Handler { //nolint:interfacer
	return a.Authorized(handler)
}

// Endpoint returns the OAuth service endpoint (host with optional port)
// string. Used for generating redirects.
func (a OidcAuth) Endpoint() string {
	return a.endpoint
}
