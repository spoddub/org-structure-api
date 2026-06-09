package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"org-structure-api/internal/analytics"
)

type CreateEmployeeParams struct {
	FullName string  `json:"full_name"`
	Position string  `json:"position"`
	HiredAt  *string `json:"hired_at"`
}

// createEmployee godoc
// @Summary Create employee
// @Description Create a new employee in an existing department.
// @Tags employees
// @Accept json
// @Produce json
// @Param id path int true "Department ID"
// @Param request body CreateEmployeeParams true "Employee payload"
// @Success 201 {object} model.Employee
// @Failure 400 {string} string "invalid request"
// @Failure 404 {string} string "department not found"
// @Failure 500 {string} string "internal server error"
// @Router /departments/{id}/employees/ [post]
func (h *Handler) createEmployee(w http.ResponseWriter, r *http.Request) {
	departmentIDStr := r.PathValue("id")
	departmentID, err := strconv.ParseInt(departmentIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid department id", http.StatusBadRequest)
		return
	}

	var params CreateEmployeeParams

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	params.FullName = strings.TrimSpace(params.FullName)
	if err := checkFullName(params.FullName); err != nil {
		http.Error(w, "invalid full name", http.StatusBadRequest)
		return
	}

	params.Position = strings.TrimSpace(params.Position)
	if err := checkPosition(params.Position); err != nil {
		http.Error(w, "invalid position", http.StatusBadRequest)
		return
	}

	hiredAtValue, hasHired, err := checkHiredDate(params.HiredAt)
	if err != nil {
		http.Error(w, "invalid hired date, expected format YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	var hiredAt *time.Time
	if hasHired {
		hiredAt = &hiredAtValue
	}

	exists, err := h.departmentRepo.Exists(r.Context(), departmentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "department not found", http.StatusNotFound)
		return
	}

	employee, err := h.employeeRepo.Create(r.Context(), departmentID, params.FullName, params.Position, hiredAt)
	if err != nil {
		http.Error(w, "error creating employee", http.StatusInternalServerError)
		return
	}

	departmentIDValue := uint64(departmentID)

	h.trackAnalytics(r.Context(), analytics.Event{
		Time:         time.Now(),
		Type:         "employee_created",
		EntityType:   "employee",
		EntityID:     uint64(employee.ID),
		DepartmentID: &departmentIDValue,
		Metadata:     "{}",
	})

	writeJSON(w, http.StatusCreated, employee)
}

func checkFullName(name string) error {
	if name == "" {
		return errors.New("full_name is empty")
	}

	if utf8.RuneCountInString(name) > 200 {
		return errors.New("full_name is too long")
	}

	return nil
}

func checkPosition(position string) error {
	if position == "" {
		return errors.New("position is empty")
	}

	if utf8.RuneCountInString(position) > 200 {
		return errors.New("position is too long")
	}

	return nil
}

func checkHiredDate(hiredAt *string) (time.Time, bool, error) {
	if hiredAt == nil {
		return time.Time{}, false, nil
	}

	trimmedTime := strings.TrimSpace(*hiredAt)
	if trimmedTime == "" {
		return time.Time{}, false, nil
	}

	t, err := time.Parse("2006-01-02", trimmedTime)
	if err != nil {
		return time.Time{}, false, err
	}

	return t, true, nil
}
