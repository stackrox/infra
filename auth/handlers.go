// Package auth facilitates an OAuth login/logout flow.
package auth

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	v1 "github.com/stackrox/infra/generated/api/v1"
	"golang.org/x/oauth2"
)

const (
	tokenCookieNew     = "token=%s"
	tokenCookieExpired = "token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT"
)

// OAuth facilitates an Oauth2 login flow via http handlers.
type OAuth struct {
	endpoint   string
	tenant     string
	jwtState   *stateTokenizer
	jwtAuth0   *auth0Tokenizer
	jwtUser    *userTokenizer
	jwtSvcAcct serviceAccountTokenizer
	conf       *oauth2.Config
}

// ValidateUser validates a user JWT and returns the contained v1.User struct.
func (a OAuth) ValidateUser(token string) (*v1.User, error) {
	return a.jwtUser.Validate(token)
}

// GenerateServiceAccountToken generates a service account JWT containing a
// v1.User struct.
func (a OAuth) GenerateServiceAccountToken(svcacct v1.ServiceAccount) (string, error) {
	return a.jwtSvcAcct.Generate(svcacct)
}

// ValidateServiceAccountToken validates a service account JWT and returns the
// contained v1.ServiceAccount struct.
func (a OAuth) ValidateServiceAccountToken(token string) (v1.ServiceAccount, error) {
	return a.jwtSvcAcct.Validate(token)
}

// loginHandler handles the login part of an Oauth2 flow.
//
// A state token is generated and sent along with the redirect to Auth0.
func (a OAuth) loginHandler(w http.ResponseWriter, r *http.Request) {
	// Generate a new state token.
	stateToken, err := a.jwtState.Generate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to Auth0 so that the user can login externally.
	audience := oauth2.SetAuthURLParam("audience", fmt.Sprintf("https://%s/userinfo", a.tenant))
	url := a.conf.AuthCodeURL(stateToken, audience)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// callbackHandler handles the callback part of an Oauth2 flow.
//
// After returning from Auth0, the state token is verified. A user profile is
// then obtained from Auth0 that includes details about the newly logged-in
// user. This user information is then stored in a cookie.
func (a OAuth) callbackHandler(w http.ResponseWriter, r *http.Request) {
	// Get the value of the "state" HTTP GET param, and validate that it is
	// legitimate.
	stateToken := r.URL.Query().Get("state")
	err := a.jwtState.Validate(stateToken)
	if err != nil {
		log.Printf("failed to validate state token: %v", err)
		http.Redirect(w, r, "/logout", http.StatusTemporaryRedirect)
		return
	}

	// Get the value of the "code" HTTP GET param, and exchange it for a user token.
	code := r.URL.Query().Get("code")
	oauthToken, err := a.conf.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("failed to exchange code: %v", err)
		http.Redirect(w, r, "/logout", http.StatusTemporaryRedirect)
		return
	}

	// Validate that the Auth0 user profile token is legitimate, and extract a
	// user struct from it.
	auth0Token := oauthToken.Extra("id_token").(string)
	user, err := a.jwtAuth0.Validate(auth0Token)
	if err != nil {
		log.Printf("failed to validate auth0 token: %v", err)
		http.Redirect(w, r, "/logout", http.StatusTemporaryRedirect)
		return
	}

	// Generate token containing a user struct.
	userToken, err := a.jwtUser.Generate(user)
	if err != nil {
		log.Printf("failed to generate user token: %v", err)
		http.Redirect(w, r, "/logout", http.StatusTemporaryRedirect)
		return
	}

	// Persist the user token as a cookie in the user's browser and redirect to
	// a logged in page
	w.Header().Set("set-cookie", fmt.Sprintf(tokenCookieNew, userToken))
	http.Redirect(w, r, "/v1/whoami", http.StatusTemporaryRedirect)
}

// logoutHandler handles the logout part of an Oauth2 flow.
//
// The user token cookie is destroyed, and the user is redirected to Auth0 for logout.
func (a OAuth) logoutHandler(w http.ResponseWriter, r *http.Request) {
	URL, _ := url.Parse(fmt.Sprintf("https://%s/v2/logout", a.tenant))
	parameters := url.Values{}
	parameters.Add("returnTo", fmt.Sprintf("https://%s/login", a.endpoint))
	parameters.Add("client_id", a.conf.ClientID)
	URL.RawQuery = parameters.Encode()

	w.Header().Set("set-cookie", tokenCookieExpired)
	http.Redirect(w, r, URL.String(), http.StatusTemporaryRedirect)
}

// Handle adds several standard OAuth routes handlers to the given http mux.
func (a OAuth) Handle(mux *http.ServeMux) {
	mux.Handle("/callback", http.HandlerFunc(a.callbackHandler))
	mux.Handle("/login", http.HandlerFunc(a.loginHandler))
	mux.Handle("/logout", http.HandlerFunc(a.logoutHandler))
}
