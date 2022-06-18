package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rezaDastrs/pkg/config"
	"github.com/rezaDastrs/pkg/handlers"
)

// CsrfToken  Handler
func routes(app *config.AppConfig) http.Handler {
	//with using Chi package
	mux := chi.NewRouter()

	//recover
	mux.Use(middleware.Recoverer)

	mux.Use(Nosurf)

	mux.Use(SessionLoad)

	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)

	return mux
}
