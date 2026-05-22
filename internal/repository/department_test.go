package repository

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"org-structure-api/internal/model"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	replacer := strings.NewReplacer("/", "_", " ", "_")
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", replacer.Replace(t.Name()))

	database, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := database.DB()
	require.NoError(t, err)

	sqlDB.SetMaxOpenConns(1)

	t.Cleanup(func() {
		require.NoError(t, sqlDB.Close())
	})

	require.NoError(t, database.AutoMigrate(&model.Department{}, &model.Employee{}))

	return database
}

func TestDepartmentRepositoryCreateGetExistsAndListChildren(t *testing.T) {
	ctx := context.Background()

	database := setupRepositoryTestDB(t)
	repo := NewDepartmentRepository(database)

	root, err := repo.Create(ctx, "Engineering", nil)
	require.NoError(t, err)
	require.NotZero(t, root.ID)
	require.Equal(t, "Engineering", root.Name)
	require.Nil(t, root.ParentID)

	exists, err := repo.Exists(ctx, root.ID)
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = repo.Exists(ctx, 999)
	require.NoError(t, err)
	require.False(t, exists)

	child, err := repo.Create(ctx, "Backend", &root.ID)
	require.NoError(t, err)
	require.Equal(t, root.ID, *child.ParentID)

	got, err := repo.GetByID(ctx, root.ID)
	require.NoError(t, err)
	require.Equal(t, root.ID, got.ID)
	require.Equal(t, "Engineering", got.Name)

	children, err := repo.ListChildren(ctx, root.ID)
	require.NoError(t, err)
	require.Len(t, children, 1)
	require.Equal(t, child.ID, children[0].ID)
	require.Equal(t, "Backend", children[0].Name)
}

func TestDepartmentRepositoryExistsByNameAndParent(t *testing.T) {
	ctx := context.Background()

	database := setupRepositoryTestDB(t)
	repo := NewDepartmentRepository(database)

	root, err := repo.Create(ctx, "Engineering", nil)
	require.NoError(t, err)

	backend, err := repo.Create(ctx, "Backend", &root.ID)
	require.NoError(t, err)

	exists, err := repo.ExistsByNameAndParent(ctx, "Engineering", nil, nil)
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = repo.ExistsByNameAndParent(ctx, "engineering", nil, nil)
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = repo.ExistsByNameAndParent(ctx, "Backend", &root.ID, nil)
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = repo.ExistsByNameAndParent(ctx, "Backend", nil, nil)
	require.NoError(t, err)
	require.False(t, exists)

	exists, err = repo.ExistsByNameAndParent(ctx, "Backend", &root.ID, &backend.ID)
	require.NoError(t, err)
	require.False(t, exists)
}

func TestDepartmentRepositoryUpdate(t *testing.T) {
	ctx := context.Background()

	database := setupRepositoryTestDB(t)
	repo := NewDepartmentRepository(database)

	root, err := repo.Create(ctx, "Engineering", nil)
	require.NoError(t, err)

	child, err := repo.Create(ctx, "Backend", &root.ID)
	require.NoError(t, err)
	require.NotNil(t, child.ParentID)

	child.Name = "Backend Platform"
	child.ParentID = nil

	updated, err := repo.Update(ctx, child)
	require.NoError(t, err)

	require.Equal(t, child.ID, updated.ID)
	require.Equal(t, "Backend Platform", updated.Name)
	require.Nil(t, updated.ParentID)
}

func TestDepartmentRepositoryIsDescendant(t *testing.T) {
	ctx := context.Background()

	database := setupRepositoryTestDB(t)
	repo := NewDepartmentRepository(database)

	root, err := repo.Create(ctx, "Engineering", nil)
	require.NoError(t, err)

	child, err := repo.Create(ctx, "Backend", &root.ID)
	require.NoError(t, err)

	grandChild, err := repo.Create(ctx, "Platform", &child.ID)
	require.NoError(t, err)

	isDescendant, err := repo.IsDescendant(ctx, root.ID, grandChild.ID)
	require.NoError(t, err)
	require.True(t, isDescendant)

	isDescendant, err = repo.IsDescendant(ctx, root.ID, child.ID)
	require.NoError(t, err)
	require.True(t, isDescendant)

	isDescendant, err = repo.IsDescendant(ctx, child.ID, root.ID)
	require.NoError(t, err)
	require.False(t, isDescendant)
}

func TestDepartmentRepositoryDeleteCascade(t *testing.T) {
	ctx := context.Background()

	database := setupRepositoryTestDB(t)
	departmentRepo := NewDepartmentRepository(database)
	employeeRepo := NewEmployeeRepository(database)

	root, err := departmentRepo.Create(ctx, "Temp Root", nil)
	require.NoError(t, err)

	child, err := departmentRepo.Create(ctx, "Temp Child", &root.ID)
	require.NoError(t, err)

	hiredAt := time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC)

	_, err = employeeRepo.Create(ctx, root.ID, "Root Employee", "Tester", &hiredAt)
	require.NoError(t, err)

	_, err = employeeRepo.Create(ctx, child.ID, "Child Employee", "Tester", &hiredAt)
	require.NoError(t, err)

	err = departmentRepo.DeleteCascade(ctx, root.ID)
	require.NoError(t, err)

	exists, err := departmentRepo.Exists(ctx, root.ID)
	require.NoError(t, err)
	require.False(t, exists)

	exists, err = departmentRepo.Exists(ctx, child.ID)
	require.NoError(t, err)
	require.False(t, exists)

	employees, err := employeeRepo.ListByDepartmentID(ctx, root.ID)
	require.NoError(t, err)
	require.Empty(t, employees)

	employees, err = employeeRepo.ListByDepartmentID(ctx, child.ID)
	require.NoError(t, err)
	require.Empty(t, employees)
}

func TestDepartmentRepositoryDeleteReassign(t *testing.T) {
	ctx := context.Background()

	database := setupRepositoryTestDB(t)
	departmentRepo := NewDepartmentRepository(database)
	employeeRepo := NewEmployeeRepository(database)

	target, err := departmentRepo.Create(ctx, "Engineering", nil)
	require.NoError(t, err)

	source, err := departmentRepo.Create(ctx, "Temporary", nil)
	require.NoError(t, err)

	child, err := departmentRepo.Create(ctx, "Temporary Child", &source.ID)
	require.NoError(t, err)

	hiredAt := time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC)

	sourceEmployee, err := employeeRepo.Create(ctx, source.ID, "Source Employee", "Developer", &hiredAt)
	require.NoError(t, err)

	childEmployee, err := employeeRepo.Create(ctx, child.ID, "Child Employee", "Tester", &hiredAt)
	require.NoError(t, err)

	err = departmentRepo.DeleteReassign(ctx, source.ID, target.ID)
	require.NoError(t, err)

	exists, err := departmentRepo.Exists(ctx, source.ID)
	require.NoError(t, err)
	require.False(t, exists)

	exists, err = departmentRepo.Exists(ctx, child.ID)
	require.NoError(t, err)
	require.False(t, exists)

	employees, err := employeeRepo.ListByDepartmentID(ctx, target.ID)
	require.NoError(t, err)
	require.Len(t, employees, 2)

	employeeIDs := []int64{
		employees[0].ID,
		employees[1].ID,
	}

	require.ElementsMatch(t, []int64{sourceEmployee.ID, childEmployee.ID}, employeeIDs)

	for _, employee := range employees {
		require.Equal(t, target.ID, employee.DepartmentID)
	}
}
