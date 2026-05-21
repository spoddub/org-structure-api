package repository

import (
	"context"
	"org-structure-api/internal/model"
	"time"

	"gorm.io/gorm"
)

type EmployeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}

func (r *EmployeeRepository) Create(ctx context.Context, departmentID int64, fullName string, position string, hiredAt *time.Time) (*model.Employee, error) {
	employee := &model.Employee{
		DepartmentID: departmentID,
		FullName:     fullName,
		Position:     position,
		HiredAt:      hiredAt,
	}

	if err := r.db.WithContext(ctx).Create(employee).Error; err != nil {
		return nil, err
	}

	return employee, nil
}
