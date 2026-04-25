package auth

import (
	"net/http"
	"time"

	"github.com/bonggalshn/budget-be/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
)

func Routes(handler *auth.Handler, middleware *auth.Middleware) http.Handler {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(10, time.Minute))

		r.Post("/login", handler.Login)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate)

		r.Get("/me", handler.GetMe)
		r.Post("/logout", handler.Logout)
	})

	return r
}