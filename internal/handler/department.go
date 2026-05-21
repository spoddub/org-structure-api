package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"org-structure-api/internal/model"
	"strconv"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
)

type CreateDepartmentParams struct {
	Name     string `json:"name"`
	ParentID *int64 `json:"parent_id"`
}

type DepartmentNodeResponse struct {
	Department model.Department         `json:"department"`
	Employees  []*model.Employee        `json:"employees"`
	Children   []DepartmentNodeResponse `json:"children"`
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

func (h *Handler) getDepartmentByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid department ID", http.StatusBadRequest)
		return
	}

	depthStr := r.URL.Query().Get("depth")

	depth, err := checkDepth(depthStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	includeEmployeesStr := r.URL.Query().Get("include_employees")

	includeEmployees, err := checkIncludeEmployees(includeEmployeesStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	department, err := h.departmentRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "department not found", http.StatusNotFound)
			return
		}

		http.Error(w, "error getting department", http.StatusInternalServerError)
		return
	}

	response, err := h.buildDepartmentNode(r.Context(), *department, depth, includeEmployees)
	if err != nil {
		http.Error(w, "error building tree", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, response)

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

func checkDepth(depthStr string) (int, error) {
	if depthStr == "" {
		return 1, nil
	}

	depth, err := strconv.Atoi(depthStr)
	if err != nil {
		return 0, errors.New("invalid department depth")
	}

	if depth < 0 {
		return 0, errors.New("invalid department depth")
	}

	if depth > 5 {
		return 0, errors.New("depth must be between <= 5")
	}

	return depth, nil
}

func checkIncludeEmployees(include string) (bool, error) {
	include = strings.TrimSpace(strings.ToLower(include))
	switch include {
	case "":
		return true, nil
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, errors.New("invalid include_employees")
	}
}

func (h *Handler) buildDepartmentNode(
	ctx context.Context,
	department model.Department,
	depth int,
	includeEmployees bool,
) (DepartmentNodeResponse, error) {
	node := DepartmentNodeResponse{
		Department: department,
		Employees:  make([]*model.Employee, 0),
		Children:   make([]DepartmentNodeResponse, 0),
	}

	if includeEmployees {
		employees, err := h.employeeRepo.ListByDepartmentID(ctx, department.ID)
		if err != nil {
			return DepartmentNodeResponse{}, err
		}

		node.Employees = employees
	}

	if depth == 0 {
		return node, nil
	}

	childrenDepartments, err := h.departmentRepo.ListChildren(ctx, department.ID)
	if err != nil {
		return DepartmentNodeResponse{}, err
	}

	for _, childDepartment := range childrenDepartments {
		childNode, err := h.buildDepartmentNode(ctx, childDepartment, depth-1, includeEmployees)
		if err != nil {
			return DepartmentNodeResponse{}, err
		}

		node.Children = append(node.Children, childNode)
	}

	return node, nil
}
