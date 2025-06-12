package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	PasswordHash string     `json:"-" db:"password_hash"`
	Role         string     `json:"role" db:"role"`
	Salary       *float64   `json:"salary,omitempty" db:"salary"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy    *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy    *uuid.UUID `json:"updated_by,omitempty" db:"updated_by"`
}

type AttendancePeriod struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	Name               string     `json:"name" db:"name"`
	StartDate          time.Time  `json:"start_date" db:"start_date"`
	EndDate            time.Time  `json:"end_date" db:"end_date"`
	IsActive           bool       `json:"is_active" db:"is_active"`
	PayrollProcessed   bool       `json:"payroll_processed" db:"payroll_processed"`
	PayrollProcessedAt *time.Time `json:"payroll_processed_at,omitempty" db:"payroll_processed_at"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy          uuid.UUID  `json:"created_by" db:"created_by"`
	UpdatedBy          *uuid.UUID `json:"updated_by,omitempty" db:"updated_by"`
}

type Attendance struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	UserID             uuid.UUID  `json:"user_id" db:"user_id"`
	AttendancePeriodID uuid.UUID  `json:"attendance_period_id" db:"attendance_period_id"`
	AttendanceDate     time.Time  `json:"attendance_date" db:"attendance_date"`
	CheckInTime        *time.Time `json:"check_in_time,omitempty" db:"check_in_time"`
	CheckOutTime       *time.Time `json:"check_out_time,omitempty" db:"check_out_time"`
	IsPresent          bool       `json:"is_present" db:"is_present"`
	IPAddress          string     `json:"ip_address,omitempty" db:"ip_address"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy          uuid.UUID  `json:"created_by" db:"created_by"`
	UpdatedBy          *uuid.UUID `json:"updated_by,omitempty" db:"updated_by"`
}

type Overtime struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	UserID             uuid.UUID  `json:"user_id" db:"user_id"`
	AttendancePeriodID uuid.UUID  `json:"attendance_period_id" db:"attendance_period_id"`
	OvertimeDate       time.Time  `json:"overtime_date" db:"overtime_date"`
	HoursWorked        float64    `json:"hours_worked" db:"hours_worked"`
	Description        string     `json:"description,omitempty" db:"description"`
	IPAddress          string     `json:"ip_address,omitempty" db:"ip_address"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy          uuid.UUID  `json:"created_by" db:"created_by"`
	UpdatedBy          *uuid.UUID `json:"updated_by,omitempty" db:"updated_by"`
}

type Reimbursement struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	UserID             uuid.UUID  `json:"user_id" db:"user_id"`
	AttendancePeriodID uuid.UUID  `json:"attendance_period_id" db:"attendance_period_id"`
	Amount             float64    `json:"amount" db:"amount"`
	Description        string     `json:"description" db:"description"`
	ReceiptURL         string     `json:"receipt_url,omitempty" db:"receipt_url"`
	Status             string     `json:"status" db:"status"`
	IPAddress          string     `json:"ip_address,omitempty" db:"ip_address"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy          uuid.UUID  `json:"created_by" db:"created_by"`
	UpdatedBy          *uuid.UUID `json:"updated_by,omitempty" db:"updated_by"`
}

// Request DTOs
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type CreateAttendancePeriodRequest struct {
	Name      string `json:"name" validate:"required"`
	StartDate string `json:"start_date" validate:"required"` // YYYY-MM-DD format
	EndDate   string `json:"end_date" validate:"required"`   // YYYY-MM-DD format
}

type SubmitAttendanceRequest struct {
	AttendancePeriodID string `json:"attendance_period_id" validate:"required,uuid"`
	AttendanceDate     string `json:"attendance_date" validate:"required"` // YYYY-MM-DD format
}

type SubmitOvertimeRequest struct {
	AttendancePeriodID string  `json:"attendance_period_id" validate:"required,uuid"`
	OvertimeDate       string  `json:"overtime_date" validate:"required"` // YYYY-MM-DD format
	HoursWorked        float64 `json:"hours_worked" validate:"required,min=0.1,max=3"`
	Description        string  `json:"description,omitempty"`
}

type SubmitReimbursementRequest struct {
	AttendancePeriodID string  `json:"attendance_period_id" validate:"required,uuid"`
	Amount             float64 `json:"amount" validate:"required,min=0.01"`
	Description        string  `json:"description" validate:"required"`
	ReceiptURL         string  `json:"receipt_url,omitempty"`
}

// Response DTOs
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
