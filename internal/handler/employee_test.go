package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"org-structure-api/internal/model"

	"github.com/stretchr/testify/require"
)

func stringPtr(value string) *string {
	return &value
}

func TestCheckFullName(t *testing.T) {
	require.NoError(t, checkFullName("Sergei Poddubny"))

	require.Error(t, checkFullName(""))
	require.Error(t, checkFullName(strings.Repeat("a", 201)))
}

func TestCheckPosition(t *testing.T) {
	require.NoError(t, checkPosition("Backend Developer"))

	require.Error(t, checkPosition(""))
	require.Error(t, checkPosition(strings.Repeat("a", 201)))
}

func TestCheckHiredDate(t *testing.T) {
	t.Run("nil hired_at", func(t *testing.T) {
		value, ok, err := checkHiredDate(nil)

		require.NoError(t, err)
		require.False(t, ok)
		require.True(t, value.IsZero())
	})

	t.Run("empty hired_at", func(t *testing.T) {
		value, ok, err := checkHiredDate(stringPtr(""))

		require.NoError(t, err)
		require.False(t, ok)
		require.True(t, value.IsZero())
	})

	t.Run("valid hired_at", func(t *testing.T) {
		value, ok, err := checkHiredDate(stringPtr("2026-05-21"))

		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC), value)
	})

	t.Run("invalid hired_at", func(t *testing.T) {
		_, _, err := checkHiredDate(stringPtr("21-05-2026"))

		require.Error(t, err)
	})
}

func TestCreateEmployee(t *testing.T) {
	router := setupTestRouter(t)

	department := createDepartmentViaHTTP(t, router, "Engineering", nil)

	payload := []byte(`{
		"full_name":"Sergei Poddubny",
		"position":"Backend Developer",
		"hired_at":"2026-05-21"
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/departments/%d/employees/", department.ID),
		bytes.NewReader(payload),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var employee model.Employee
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&employee))

	require.NotZero(t, employee.ID)
	require.Equal(t, department.ID, employee.DepartmentID)
	require.Equal(t, "Sergei Poddubny", employee.FullName)
	require.Equal(t, "Backend Developer", employee.Position)
	require.NotNil(t, employee.HiredAt)
}

func TestCreateEmployeeWithUnknownDepartment(t *testing.T) {
	router := setupTestRouter(t)

	payload := []byte(`{
		"full_name":"Sergei Poddubny",
		"position":"Backend Developer",
		"hired_at":"2026-05-21"
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/departments/999/employees/",
		bytes.NewReader(payload),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}
