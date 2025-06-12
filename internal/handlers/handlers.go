package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jordanhimawan/payroll-mgmt/internal/models"
	"github.com/jordanhimawan/payroll-mgmt/pkg/utils"
)

// App struct
type App struct {
	DB        *pgxpool.Pool
	JWTSecret []byte
}

// Handlers
func (app *App) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var user models.User
	query := `
		SELECT id, username, password_hash, role, salary, is_active, created_at, updated_at
		FROM users 
		WHERE username = $1 AND is_active = true
	`

	err := app.DB.QueryRow(context.Background(), query, req.Username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role,
		&user.Salary, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT
	claims := &models.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(app.JWTSecret)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := models.LoginResponse{
		Token: tokenString,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (app *App) CreateAttendancePeriodHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAttendancePeriodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		http.Error(w, "Invalid start date format", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		http.Error(w, "Invalid end date format", http.StatusBadRequest)
		return
	}

	if endDate.Before(startDate) {
		http.Error(w, "End date must be after start date", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)

	var period models.AttendancePeriod
	query := `
		INSERT INTO attendance_periods (name, start_date, end_date, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, start_date, end_date, is_active, payroll_processed, created_at, updated_at, created_by
	`

	err = app.DB.QueryRow(context.Background(), query, req.Name, startDate, endDate, userID).Scan(
		&period.ID, &period.Name, &period.StartDate, &period.EndDate,
		&period.IsActive, &period.PayrollProcessed, &period.CreatedAt,
		&period.UpdatedAt, &period.CreatedBy,
	)

	if err != nil {
		http.Error(w, "Failed to create attendance period", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(period)
}

func (app *App) SubmitAttendanceHandler(w http.ResponseWriter, r *http.Request) {
	var req models.SubmitAttendanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	attendanceDate, err := time.Parse("2006-01-02", req.AttendanceDate)
	if err != nil {
		http.Error(w, "Invalid attendance date format", http.StatusBadRequest)
		return
	}

	if utils.IsWeekend(attendanceDate) {
		http.Error(w, "Cannot submit attendance on weekends", http.StatusBadRequest)
		return
	}

	periodID, err := uuid.Parse(req.AttendancePeriodID)
	if err != nil {
		http.Error(w, "Invalid attendance period ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)
	ipAddress := utils.GetClientIP(r)
	now := time.Now()

	var attendance models.Attendance
	query := `
		INSERT INTO attendances (user_id, attendance_period_id, attendance_date, check_in_time, ip_address, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, attendance_date) 
		DO UPDATE SET 
			check_out_time = $4,
			updated_at = CURRENT_TIMESTAMP,
			updated_by = $6
		RETURNING id, user_id, attendance_period_id, attendance_date, check_in_time, 
				  check_out_time, is_present, ip_address, created_at, updated_at, created_by
	`

	err = app.DB.QueryRow(context.Background(), query, userID, periodID, attendanceDate, now, ipAddress, userID).Scan(
		&attendance.ID, &attendance.UserID, &attendance.AttendancePeriodID,
		&attendance.AttendanceDate, &attendance.CheckInTime, &attendance.CheckOutTime,
		&attendance.IsPresent, &attendance.IPAddress, &attendance.CreatedAt,
		&attendance.UpdatedAt, &attendance.CreatedBy,
	)

	if err != nil {
		http.Error(w, "Failed to submit attendance", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(attendance)
}

func (app *App) SubmitOvertimeHandler(w http.ResponseWriter, r *http.Request) {
	var req models.SubmitOvertimeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.HoursWorked <= 0 || req.HoursWorked > 3 {
		http.Error(w, "Overtime hours must be between 0 and 3", http.StatusBadRequest)
		return
	}

	overtimeDate, err := time.Parse("2006-01-02", req.OvertimeDate)
	if err != nil {
		http.Error(w, "Invalid overtime date format", http.StatusBadRequest)
		return
	}

	periodID, err := uuid.Parse(req.AttendancePeriodID)
	if err != nil {
		http.Error(w, "Invalid attendance period ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)
	ipAddress := utils.GetClientIP(r)

	var overtime models.Overtime
	query := `
		INSERT INTO overtimes (user_id, attendance_period_id, overtime_date, hours_worked, description, ip_address, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, overtime_date)
		DO UPDATE SET 
			hours_worked = $4,
			description = $5,
			updated_at = CURRENT_TIMESTAMP,
			updated_by = $7
		RETURNING id, user_id, attendance_period_id, overtime_date, hours_worked, 
				  description, ip_address, created_at, updated_at, created_by
	`

	err = app.DB.QueryRow(context.Background(), query, userID, periodID, overtimeDate, req.HoursWorked, req.Description, ipAddress, userID).Scan(
		&overtime.ID, &overtime.UserID, &overtime.AttendancePeriodID,
		&overtime.OvertimeDate, &overtime.HoursWorked, &overtime.Description,
		&overtime.IPAddress, &overtime.CreatedAt, &overtime.UpdatedAt, &overtime.CreatedBy,
	)

	if err != nil {
		http.Error(w, "Failed to submit overtime", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(overtime)
}

func (app *App) SubmitReimbursementHandler(w http.ResponseWriter, r *http.Request) {
	var req models.SubmitReimbursementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, "Amount must be positive", http.StatusBadRequest)
		return
	}

	if req.Description == "" {
		http.Error(w, "Description is required", http.StatusBadRequest)
		return
	}

	periodID, err := uuid.Parse(req.AttendancePeriodID)
	if err != nil {
		http.Error(w, "Invalid attendance period ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)
	ipAddress := utils.GetClientIP(r)

	var reimbursement models.Reimbursement
	query := `
		INSERT INTO reimbursements (user_id, attendance_period_id, amount, description, receipt_url, ip_address, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, attendance_period_id, amount, description, receipt_url, 
				  status, ip_address, created_at, updated_at, created_by
	`

	err = app.DB.QueryRow(context.Background(), query, userID, periodID, req.Amount, req.Description, req.ReceiptURL, ipAddress, userID).Scan(
		&reimbursement.ID, &reimbursement.UserID, &reimbursement.AttendancePeriodID,
		&reimbursement.Amount, &reimbursement.Description, &reimbursement.ReceiptURL,
		&reimbursement.Status, &reimbursement.IPAddress, &reimbursement.CreatedAt,
		&reimbursement.UpdatedAt, &reimbursement.CreatedBy,
	)

	if err != nil {
		http.Error(w, "Failed to submit reimbursement", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reimbursement)
}

func (app *App) GetAttendancePeriodsHandler(w http.ResponseWriter, r *http.Request) {
	var periods []models.AttendancePeriod
	query := `
		SELECT id, name, start_date, end_date, is_active, payroll_processed, 
			   payroll_processed_at, created_at, updated_at, created_by
		FROM attendance_periods 
		ORDER BY created_at DESC
	`

	rows, err := app.DB.Query(context.Background(), query)
	if err != nil {
		http.Error(w, "Failed to fetch attendance periods", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var period models.AttendancePeriod
		err := rows.Scan(
			&period.ID, &period.Name, &period.StartDate, &period.EndDate,
			&period.IsActive, &period.PayrollProcessed, &period.PayrollProcessedAt,
			&period.CreatedAt, &period.UpdatedAt, &period.CreatedBy,
		)
		if err != nil {
			http.Error(w, "Failed to scan attendance period", http.StatusInternalServerError)
			return
		}
		periods = append(periods, period)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(periods)
}
