package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEmployeeRepositoryCreateAndListByDepartmentID(t *testing.T) {
	ctx := context.Background()

	database := setupRepositoryTestDB(t)

	departmentRepo := NewDepartmentRepository(database)
	employeeRepo := NewEmployeeRepository(database)

	engineering, err := departmentRepo.Create(ctx, "Engineering", nil)
	require.NoError(t, err)

	product, err := departmentRepo.Create(ctx, "Product", nil)
	require.NoError(t, err)

	hiredAt := time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC)

	employee, err := employeeRepo.Create(
		ctx,
		engineering.ID,
		"Sergei Poddubny",
		"Backend Developer",
		&hiredAt,
	)
	require.NoError(t, err)

	_, err = employeeRepo.Create(
		ctx,
		product.ID,
		"John Doe",
		"Product Manager",
		nil,
	)
	require.NoError(t, err)

	employees, err := employeeRepo.ListByDepartmentID(ctx, engineering.ID)
	require.NoError(t, err)

	require.Len(t, employees, 1)
	require.Equal(t, employee.ID, employees[0].ID)
	require.Equal(t, engineering.ID, employees[0].DepartmentID)
	require.Equal(t, "Sergei Poddubny", employees[0].FullName)
	require.Equal(t, "Backend Developer", employees[0].Position)
	require.NotNil(t, employees[0].HiredAt)
}

func TestEmployeeRepositoryListByDepartmentIDReturnsEmptySlice(t *testing.T) {
	ctx := context.Background()

	database := setupRepositoryTestDB(t)

	departmentRepo := NewDepartmentRepository(database)
	employeeRepo := NewEmployeeRepository(database)

	department, err := departmentRepo.Create(ctx, "Engineering", nil)
	require.NoError(t, err)

	employees, err := employeeRepo.ListByDepartmentID(ctx, department.ID)
	require.NoError(t, err)

	require.Empty(t, employees)
}
