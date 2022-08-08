package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rezaDastrs/internal/config"
	"github.com/rezaDastrs/internal/handlers"
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
	mux.Get("/generals-quarters", handlers.Repo.Generals)
	mux.Get("/majors-suite", handlers.Repo.Majors)

	mux.Get("/search-availability", handlers.Repo.Availibility)
	mux.Post("/search-availability", handlers.Repo.PostAvailibility)
	mux.Get("/search-availability-json", handlers.Repo.AvailibilityJSON)

	mux.Get("/contact", handlers.Repo.Contact)

	mux.Get("/make-reservation", handlers.Repo.Reservation)
	mux.Post("/make-reservation", handlers.Repo.PostReservation)
	mux.Get("/reservation-summery", handlers.Repo.ReservationSummery)
	mux.Get("/choose-room/{id}", handlers.Repo.ChooseRoom)
	mux.Get("/book-room", handlers.Repo.BookRoom)

	mux.Get("/user/login", handlers.Repo.LoginPage)
	mux.Post("/user/login", handlers.Repo.PostLoginPage)
	mux.Get("/user/logout", handlers.Repo.Logout)

	//Admin
	mux.Route("/admin", func(mux chi.Router) {
		//use Auth middleware
		// mux.Use(Auth)

		mux.Get("/dashboard", handlers.Repo.AdminDashboard)

		mux.Get("/reservations-new", handlers.Repo.AdminNewReservations)
		mux.Get("/reservations-all", handlers.Repo.AdminAllReservations)
		mux.Get("/reservation-calendar", handlers.Repo.AdminCalendarReservation)
		mux.Post("/reservation-calendar", handlers.Repo.AdminPostCalendarReservation)
		mux.Get("/reservations/{src}/{id}", handlers.Repo.AdminShowReservation)
		mux.Post("/reservations/{src}/{id}", handlers.Repo.AdminPostShowReservation)

		mux.Get("/process-reservation/{src}/{id}", handlers.Repo.AdminProcessedReservation)
		mux.Get("/delete-reservation/{src}/{id}", handlers.Repo.AdminDeleteReservation)
	})

	//read static file
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	return mux
}
