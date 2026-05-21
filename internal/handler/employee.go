package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type CreateEmployeeParams struct {
	FullName string  `json:"full_name"`
	Position string  `json:"position"`
	HiredAt  *string `json:"hired_at"`
}

type CreateEmployeeResponse struct {
	DepartmentID int64      `json:"department_id"`
	FullName     string     `json:"full_name"`
	Position     string     `json:"position"`
	HiredAt      *time.Time `json:"hired_at"`
}

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

	hiredAt, err := checkHiredDate(params.HiredAt)
	if err != nil {
		http.Error(w, "invalid hired date, expected format YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	response := CreateEmployeeResponse{
		DepartmentID: departmentID,
		FullName:     params.FullName,
		Position:     params.Position,
		HiredAt:      hiredAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "error creating employee", http.StatusInternalServerError)
		return
	}
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

func checkHiredDate(hiredAt *string) (*time.Time, error) {
	if hiredAt == nil {
		return nil, nil
	}

	trimmedTime := strings.TrimSpace(*hiredAt)
	if trimmedTime == "" {
		return nil, nil
	}

	t, err := time.Parse("2006-01-02", trimmedTime)
	if err != nil {
		return nil, err
	}

	return &t, nil
}
