package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jordanhimawan/payroll-mgmt/internal/models"
)

type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
}

type AttendancePeriodRepository interface {
	Create(ctx context.Context, period *models.AttendancePeriod) error
	GetAll(ctx context.Context) ([]models.AttendancePeriod, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.AttendancePeriod, error)
	Update(ctx context.Context, period *models.AttendancePeriod) error
}

type AttendanceRepository interface {
	Create(ctx context.Context, attendance *models.Attendance) error
	Update(ctx context.Context, attendance *models.Attendance) error
	GetByUserAndDate(ctx context.Context, userID uuid.UUID, date time.Time) (*models.Attendance, error)
	GetByUserAndPeriod(ctx context.Context, userID, periodID uuid.UUID) ([]models.Attendance, error)
}

type OvertimeRepository interface {
	Create(ctx context.Context, overtime *models.Overtime) error
	Update(ctx context.Context, overtime *models.Overtime) error
	GetByUserAndDate(ctx context.Context, userID uuid.UUID, date time.Time) (*models.Overtime, error)
	GetByUserAndPeriod(ctx context.Context, userID, periodID uuid.UUID) ([]models.Overtime, error)
}

type ReimbursementRepository interface {
	Create(ctx context.Context, reimbursement *models.Reimbursement) error
	Update(ctx context.Context, reimbursement *models.Reimbursement) error
	GetByUserAndPeriod(ctx context.Context, userID, periodID uuid.UUID) ([]models.Reimbursement, error)
}
