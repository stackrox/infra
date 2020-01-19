package server

import (
	"net/http"

	"github.com/stackrox/infra/auth"
	"github.com/stackrox/infra/config"
)

func buildRoutes(gwMux http.Handler, cfg *config.Config) *http.ServeMux {
	a, err := auth.NewOAuth(cfg)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(cfg.Storage.StaticDir)))
	mux.Handle("/callback", http.HandlerFunc(a.CallbackHandler))
	mux.Handle("/login", http.HandlerFunc(a.LoginHandler))
	mux.Handle("/logout", http.HandlerFunc(a.LogoutHandler))
	mux.Handle("/v1/", gwMux)

	return mux
}
