package repository

import (
	"context"
	"org-structure-api/internal/model"

	"gorm.io/gorm"
)

type DepartmentRepository struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) *DepartmentRepository {
	return &DepartmentRepository{db: db}
}

func (r *DepartmentRepository) Create(ctx context.Context, name string, parentID *int64) (*model.Department, error) {
	department := &model.Department{
		Name:     name,
		ParentID: parentID,
	}

	if err := r.db.WithContext(ctx).Create(department).Error; err != nil {
		return nil, err
	}

	return department, nil
}

func (r *DepartmentRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).Model(model.Department{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
