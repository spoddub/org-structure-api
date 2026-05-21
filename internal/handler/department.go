package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"
)

type CreateDepartmentParams struct {
	Name     string `json:"name"`
	ParentID *int64 `json:"parent_id"`
}

func (h *Handler) createDepartment(w http.ResponseWriter, r *http.Request) {
	var params CreateDepartmentParams

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	params.Name = strings.TrimSpace(params.Name)

	if err := checkDepartmentName(params.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if params.ParentID != nil {
		exists, err := h.departmentRepo.Exists(r.Context(), *params.ParentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !exists {
			http.Error(w, "parent department not found", http.StatusNotFound)
			return
		}
	}

	department, err := h.departmentRepo.Create(r.Context(), params.Name, params.ParentID)
	if err != nil {
		http.Error(w, "error creating department", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, department)
}

func checkDepartmentName(name string) error {
	if name == "" {
		return errors.New("name is empty")
	}

	if utf8.RuneCountInString(name) > 200 {
		return errors.New("name is too long")
	}

	return nil
}
