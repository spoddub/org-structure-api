package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"org-structure-api/internal/model"
	"org-structure-api/internal/repository"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestRouter(t *testing.T) http.Handler {
	t.Helper()

	database := setupHandlerTestDB(t)

	departmentRepo := repository.NewDepartmentRepository(database)
	employeeRepo := repository.NewEmployeeRepository(database)

	h := NewHandler(departmentRepo, employeeRepo)

	return NewRouter(h)
}

func setupHandlerTestDB(t *testing.T) *gorm.DB {
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

func createDepartmentViaHTTP(
	t *testing.T,
	router http.Handler,
	name string,
	parentID *int64,
) model.Department {
	t.Helper()

	body := map[string]any{
		"name": name,
	}

	if parentID != nil {
		body["parent_id"] = *parentID
	}

	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/departments/", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var department model.Department
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&department))

	return department
}

func TestCheckDepth(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    int
		expectedErr bool
	}{
		{
			name:     "empty means default depth",
			input:    "",
			expected: 1,
		},
		{
			name:     "zero depth",
			input:    "0",
			expected: 0,
		},
		{
			name:     "max depth",
			input:    "5",
			expected: 5,
		},
		{
			name:        "too big depth",
			input:       "6",
			expectedErr: true,
		},
		{
			name:        "negative depth",
			input:       "-1",
			expectedErr: true,
		},
		{
			name:        "invalid depth",
			input:       "abc",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := checkDepth(tt.input)

			if tt.expectedErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestCheckIncludeEmployees(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    bool
		expectedErr bool
	}{
		{
			name:     "empty means true",
			input:    "",
			expected: true,
		},
		{
			name:     "true",
			input:    "true",
			expected: true,
		},
		{
			name:     "false",
			input:    "false",
			expected: false,
		},
		{
			name:     "case insensitive",
			input:    " TRUE ",
			expected: true,
		},
		{
			name:        "invalid value",
			input:       "yes",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := checkIncludeEmployees(tt.input)

			if tt.expectedErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestCheckDepartmentName(t *testing.T) {
	require.NoError(t, checkDepartmentName("Engineering"))

	require.Error(t, checkDepartmentName(""))
	require.Error(t, checkDepartmentName(strings.Repeat("a", 201)))
}

func TestCreateAndGetDepartment(t *testing.T) {
	router := setupTestRouter(t)

	root := createDepartmentViaHTTP(t, router, "Engineering", nil)
	child := createDepartmentViaHTTP(t, router, "Backend", &root.ID)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/departments/%d", root.ID), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var response DepartmentNodeResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))

	require.Equal(t, root.ID, response.Department.ID)
	require.Equal(t, "Engineering", response.Department.Name)

	require.Empty(t, response.Employees)
	require.Len(t, response.Children, 1)
	require.Equal(t, child.ID, response.Children[0].Department.ID)
	require.Equal(t, "Backend", response.Children[0].Department.Name)
}

func TestGetDepartmentWithDepthZero(t *testing.T) {
	router := setupTestRouter(t)

	root := createDepartmentViaHTTP(t, router, "Engineering", nil)
	createDepartmentViaHTTP(t, router, "Backend", &root.ID)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/departments/%d?depth=0", root.ID), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var response DepartmentNodeResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))

	require.Equal(t, root.ID, response.Department.ID)
	require.Empty(t, response.Children)
}

func TestUpdateDepartment(t *testing.T) {
	router := setupTestRouter(t)

	root := createDepartmentViaHTTP(t, router, "Engineering", nil)
	child := createDepartmentViaHTTP(t, router, "Backend", &root.ID)

	payload := []byte(`{"name":"Backend Platform","parent_id":null}`)

	req := httptest.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("/departments/%d", child.ID),
		bytes.NewReader(payload),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var updated model.Department
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&updated))

	require.Equal(t, child.ID, updated.ID)
	require.Equal(t, "Backend Platform", updated.Name)
	require.Nil(t, updated.ParentID)
}

func TestUpdateDepartmentRejectsSelfParent(t *testing.T) {
	router := setupTestRouter(t)

	department := createDepartmentViaHTTP(t, router, "Engineering", nil)

	payload := []byte(fmt.Sprintf(`{"parent_id":%d}`, department.ID))

	req := httptest.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("/departments/%d", department.ID),
		bytes.NewReader(payload),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}
