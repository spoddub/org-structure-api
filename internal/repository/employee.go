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

func (r *EmployeeRepository) ListByDepartmentID(ctx context.Context, departmentID int64) ([]*model.Employee, error) {
	var employees []*model.Employee

	err := r.db.WithContext(ctx).
		Where("department_id = ?", departmentID).
		Order("created_at ASC").
		Find(&employees).
		Error
	if err != nil {
		return nil, err
	}

	return employees, nil
}
