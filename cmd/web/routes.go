package main

import (
	"net/http"

	"github.com/Atul-Ranjan12/tourism/internal/config"
	"github.com/Atul-Ranjan12/tourism/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	// Set up Multiplexer configuration
	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	// Login Page Routes
	mux.Get("/login", handlers.Repo.ShowLogin)
	mux.Post("/login", handlers.Repo.PostLogin)

	// SignUp Page Routes
	mux.Get("/signup", handlers.Repo.ShowSignUp)
	mux.Post("/signup", handlers.Repo.PostSignUp)

	// User logout
	mux.Get("/logout", handlers.Repo.Logout)

	// Reservation Section
	// 1. Bus Reservation
	mux.Get("/make-bus-reservation", handlers.Repo.MakeBusReservation)
	mux.Post("/make-bus-reservation", handlers.Repo.PostMakeBusReservation)

	// 2. Hotel Room Reservation
	mux.Get("/make-hotel-reservation", handlers.Repo.ShowMakeHotelReservation)
	mux.Post("/make-hotel-reservation", handlers.Repo.PostShowMakeHotelReservation)

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// TODO: Set up Routes for Logged In Users
	mux.Route("/merchant", func(mux chi.Router) {
		// Check if user is authenticated
		mux.Use(Auth)

		mux.Get("/{src}/dashboard", handlers.Repo.ShowAdminDashboard)

		// Verification Code :: Stage 1
		mux.Get("/{src}/verification", handlers.Repo.ShowAdminVerification)
		mux.Post("/{src}/verification", handlers.Repo.PostShowAdminVerification)

		// Address Verification :: Stage 2
		mux.Get("/{src}/verification-address", handlers.Repo.ShowAdminAddress)
		mux.Post("/{src}/verification-address", handlers.Repo.PostShowAdminAddress)

		// Document Verificaiotn :: Stage 3
		mux.Get("/{src}/verification-documents", handlers.Repo.ShowDocumentsVerification)
		mux.Post("/{src}/verification-documents", handlers.Repo.PostShowDocumentsVerification)

		// Merchant Dashboard Items
		// Merchant Add item Seciotn
		mux.Get("/{src}/merchant-add-items", handlers.Repo.AdminAddMerchantItems)

		// Merchant Add Bus Section:
		mux.Get("/{src}/add-bus", handlers.Repo.AdminAddBus)
		mux.Post("/{src}/add-bus", handlers.Repo.PostAdminAddBus)

		// Merchant SHow and Edit Bus section
		mux.Get("/{src}/add-bus/{id}", handlers.Repo.AdminShowBus)
		mux.Post("/{src}/add-bus/{id}", handlers.Repo.PostAdminUpdateBus)

		// Delete the bus
		mux.Get("/{src}/add-bus/delete/{id}", handlers.Repo.PostAdminDeleteBus)

		// Merchant Add Hotel Resrvation Section
		mux.Get("/{src}/add-hotel", handlers.Repo.AdminAddHotel)
		mux.Post("/{src}/add-hotel", handlers.Repo.PostAdminAddHotel)

		// Mercant Show and Edit the Bus Section
		mux.Get("/{src}/add-hotel/{id}", handlers.Repo.AdminShowOneHotel)
		mux.Post("/{src}/add-hotel/{id}", handlers.Repo.PostAdminShowOneHotel)

		// Delete the reservation
		mux.Get("/{src}/add-hotel/delete/{id}", handlers.Repo.DeleteBus)

		// Show the bus Reservation :: UnProcessed
		mux.Get("/{src}/merchant-show-reservations", handlers.Repo.ShowAllReservations)

		// Show One Single Bus Reservation
		mux.Get("/{src}/merchant-show-reservations/{id}", handlers.Repo.ShowOneReservation)
		mux.Post("/{src}/merchant-show-reservations/{id}", handlers.Repo.PostShowOneReservation)

		// Link to Process the Bus Reservation
		mux.Get("/{src}/merchant-show-reservations/{id}/process", handlers.Repo.ProcessBusReservation)

		// Show Bus Reservaitons UnProcessed
		mux.Get("/{src}/merchant-show-reservations-processed", handlers.Repo.ShowReservationsProcessed)

		// Function to delete the bus reservations
		mux.Get("/{src}/delete-reservation/{id}", handlers.Repo.DeleteBusReservation)

		// Function to get the reservation calender
		mux.Get("/{src}/reservation-calender", handlers.Repo.ShowReservationCalender)

	})

	return mux
}
