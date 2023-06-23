package main

import (
	"github.com/SeatSnobAri/seatflipsite/internal/handlers"
	"github.com/go-chi/cors"
	"net/http"
)

func routes() http.Handler {

	mux := chi.NewRouter()

	// default middleware
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	mux.Use(SessionLoad)
	mux.Use(RecoverPanic)
	mux.Use(NoSurf)
	mux.Use(CheckRemember)

	mux.Get("/", handlers.Repo.Home)

	// login
	mux.HandleFunc("/auth/google/login", handlers.Repo.OauthGoogleLogin)
	mux.HandleFunc("/auth/google/callback", handlers.Repo.OauthGoogleCallback)

	mux.Get("/user/logout", handlers.Repo.Logout)

	// Make account
	mux.Get("/user/sign-up", handlers.Repo.SignUp)
	mux.Post("/user/sign-up", handlers.Repo.PostSignUp)

	// extension stuff
	mux.Post("/broker", handlers.Repo.Broker)

	mux.Route("/pusher", func(mux chi.Router) {
		mux.Use(Auth)
		mux.Post("/auth", handlers.Repo.PusherAuth)
	})

	// admin routes
	mux.Route("/admin", func(mux chi.Router) {
		// all admin routes are protected
		mux.Use(Auth)
		mux.Get("/dashboard", handlers.Repo.AdminDashboard)
		mux.Get("/private-message", handlers.Repo.SendPrivateMessage)
		//mux.Get("/current-redis", handlers.Repo.endCurrentRedis)

	})

	// static files
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
