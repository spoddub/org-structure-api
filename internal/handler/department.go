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

type CreateDepartmentResponse struct {
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

	response := CreateDepartmentResponse{
		Name:     params.Name,
		ParentID: params.ParentID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
