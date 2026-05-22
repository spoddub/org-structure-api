package repository

import (
	"context"
	"org-structure-api/internal/model"
	"time"

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

func (r *DepartmentRepository) ExistsByNameAndParent(
	ctx context.Context,
	name string,
	parentID *int64,
	excludeID *int64,
) (bool, error) {
	var count int64

	query := r.db.WithContext(ctx).
		Model(model.Department{}).
		Where("lower(name) = lower(?)", name)

	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}

	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *DepartmentRepository) GetByID(ctx context.Context, id int64) (*model.Department, error) {
	var department model.Department

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&department).Error
	if err != nil {
		return nil, err
	}

	return &department, nil
}

func (r *DepartmentRepository) ListChildren(ctx context.Context, parentID int64) ([]model.Department, error) {
	var departments []model.Department

	err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Find(&departments).Error
	if err != nil {
		return nil, err
	}

	return departments, nil
}

func (r *DepartmentRepository) Update(ctx context.Context, department *model.Department) (*model.Department, error) {
	parentID := any(nil)
	if department.ParentID != nil {
		parentID = *department.ParentID
	}

	updates := map[string]any{
		"name":       department.Name,
		"parent_id":  parentID,
		"updated_at": time.Now(),
	}

	err := r.db.WithContext(ctx).Model(model.Department{}).Where("id = ?", department.ID).Updates(updates).Error
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, department.ID)
}

func (r *DepartmentRepository) IsDescendant(ctx context.Context, departmentID int64, possibleDescendant int64) (bool, error) {
	var exists bool

	err := r.db.WithContext(ctx).Raw(`
		WITH RECURSIVE subtree(id) AS (
			SELECT id
			FROM departments
			WHERE parent_id = ?

			UNION ALL

			SELECT departments.id
			FROM departments
			JOIN subtree ON departments.parent_id = subtree.id
		)
		SELECT EXISTS (
			SELECT 1
			FROM subtree
			WHERE id = ?
		)
	`, departmentID, possibleDescendant).Scan(&exists).Error
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *DepartmentRepository) DeleteCascade(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
			WITH RECURSIVE subtree(id) AS (
				SELECT id
				FROM departments
				WHERE id = ?

				UNION ALL

				SELECT departments.id
				FROM departments
				JOIN subtree ON departments.parent_id = subtree.id
			)
			DELETE FROM employees
			WHERE department_id IN (
				SELECT id
				FROM subtree
			)
		`, id).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
			WITH RECURSIVE subtree(id) AS (
				SELECT id
				FROM departments
				WHERE id = ?

				UNION ALL

				SELECT departments.id
				FROM departments
				JOIN subtree ON departments.parent_id = subtree.id
			)
			DELETE FROM departments
			WHERE id IN (
				SELECT id
				FROM subtree
			)
		`, id).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *DepartmentRepository) DeleteReassign(ctx context.Context, id int64, reassignToID int64) error {
	now := time.Now()

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Exec(`
			WITH RECURSIVE subtree(id) AS (
				SELECT id
				FROM departments
				WHERE id = ?

				UNION ALL

				SELECT departments.id
				FROM departments
				JOIN subtree ON departments.parent_id = subtree.id
			)
			UPDATE employees
			SET department_id = ?, updated_at = ?
			WHERE department_id IN (
				SELECT id
				FROM subtree
			)
		`, id, reassignToID, now).Error
		if err != nil {
			return err
		}

		err = tx.Exec(`
			WITH RECURSIVE subtree(id) AS (
				SELECT id
				FROM departments
				WHERE id = ?

				UNION ALL

				SELECT departments.id
				FROM departments
				JOIN subtree ON departments.parent_id = subtree.id
			)
			DELETE FROM departments
			WHERE id IN (
				SELECT id
				FROM subtree
			)
		`, id).Error
		if err != nil {
			return err
		}

		return nil
	})
}
