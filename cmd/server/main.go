package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jordanhimawan/payroll-mgmt/internal/config"
	postgres "github.com/jordanhimawan/payroll-mgmt/internal/database"
	"github.com/jordanhimawan/payroll-mgmt/internal/handlers"
	"github.com/jordanhimawan/payroll-mgmt/internal/services"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := postgres.NewPostgresConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	attendancePeriodRepo := postgres.NewAttendancePeriodRepository(db)
	attendanceRepo := postgres.NewAttendanceRepository(db)
	overtimeRepo := postgres.NewOvertimeRepository(db)
	reimbursementRepo := postgres.NewReimbursementRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	attendanceService := services.NewAttendanceService(attendanceRepo, attendancePeriodRepo)
	payrollService := services.NewPayrollService(userRepo, attendanceRepo, overtimeRepo, reimbursementRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	adminHandler := handlers.NewAdminHandler(attendancePeriodRepo)
	employeeHandler := handlers.NewEmployeeHandler(attendanceService, overtimeRepo, reimbursementRepo)
	commonHandler := handlers.NewCommonHandler(attendancePeriodRepo)

	// Initialize middleware
	authMiddleware := apiMiddleware.NewAuthMiddleware(cfg.JWTSecret)

	// Setup routes
	router := setupRoutes(authHandler, adminHandler, employeeHandler, commonHandler, authMiddleware)

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}

func setupRoutes(
	authHandler *handlers.AuthHandler,
	adminHandler *handlers.AdminHandler,
	employeeHandler *handlers.EmployeeHandler,
	commonHandler *handlers.CommonHandler,
	authMiddleware *appMiddleware.AuthMiddleware,
) chi.Router {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public routes
	r.Post("/api/v1/auth/login", authHandler.Login)

	// Protected routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(authMiddleware.Authenticate)

		// Admin routes
		r.Route("/admin", func(r chi.Router) {
			r.Use(authMiddleware.RequireAdmin)
			r.Post("/attendance-periods", adminHandler.CreateAttendancePeriod)
		})

		// Employee routes
		r.Route("/employee", func(r chi.Router) {
			r.Post("/attendance", employeeHandler.SubmitAttendance)
			r.Post("/overtime", employeeHandler.SubmitOvertime)
			r.Post("/reimbursement", employeeHandler.SubmitReimbursement)
		})

		// Common routes
		r.Get("/attendance-periods", commonHandler.GetAttendancePeriods)
	})

	return r
}
