package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"org-structure-api/internal/model"

	"gorm.io/gorm"
)

type CreateDepartmentParams struct {
	Name     string `json:"name"`
	ParentID *int64 `json:"parent_id"`
}

type UpdateDepartmentParams struct {
	Name     *string       `json:"name"`
	ParentID OptionalInt64 `json:"parent_id"`
}

type DepartmentNodeResponse struct {
	Department model.Department         `json:"department"`
	Employees  []*model.Employee        `json:"employees"`
	Children   []DepartmentNodeResponse `json:"children"`
}

type OptionalInt64 struct {
	Set   bool
	Value *int64
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
			http.Error(w, "error checking parent department", http.StatusInternalServerError)
			return
		}

		if !exists {
			http.Error(w, "parent department not found", http.StatusNotFound)
			return
		}
	}

	nameExists, err := h.departmentRepo.ExistsByNameAndParent(r.Context(), params.Name, params.ParentID, nil)
	if err != nil {
		http.Error(w, "error checking department name", http.StatusInternalServerError)
		return
	}

	if nameExists {
		http.Error(w, "department with this name already exists", http.StatusConflict)
		return
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

func (h *Handler) updateDepartment(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid department id", http.StatusBadRequest)
		return
	}

	var params UpdateDepartmentParams

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
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

	if params.Name != nil {
		name := strings.TrimSpace(*params.Name)

		if err := checkDepartmentName(name); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		department.Name = name
	}

	if params.ParentID.Set {
		if params.ParentID.Value == nil {
			department.ParentID = nil
		} else {
			newParentID := *params.ParentID.Value

			if newParentID == id {
				http.Error(w, "department cannot be parent of itself", http.StatusBadRequest)
				return
			}

			exists, err := h.departmentRepo.Exists(r.Context(), newParentID)
			if err != nil {
				http.Error(w, "error checking parent department", http.StatusInternalServerError)
				return
			}

			if !exists {
				http.Error(w, "parent department not found", http.StatusNotFound)
				return
			}

			isDescendant, err := h.departmentRepo.IsDescendant(r.Context(), id, newParentID)
			if err != nil {
				http.Error(w, "error checking department tree", http.StatusInternalServerError)
				return
			}

			if isDescendant {
				http.Error(w, "cannot move department inside its own subtree", http.StatusConflict)
				return
			}

			department.ParentID = &newParentID
		}
	}

	nameExists, err := h.departmentRepo.ExistsByNameAndParent(
		r.Context(),
		department.Name,
		department.ParentID,
		&id,
	)
	if err != nil {
		http.Error(w, "error checking department name", http.StatusInternalServerError)
		return
	}

	if nameExists {
		http.Error(w, "department with this name already exists", http.StatusConflict)
		return
	}

	updatedDepartment, err := h.departmentRepo.Update(r.Context(), department)
	if err != nil {
		http.Error(w, "error updating department", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, updatedDepartment)
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

func (h *Handler) deleteDepartment(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid department id", http.StatusBadRequest)
		return
	}

	mode := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("mode")))
	if mode == "" {
		http.Error(w, "mode is required", http.StatusBadRequest)
		return
	}

	exists, err := h.departmentRepo.Exists(r.Context(), id)
	if err != nil {
		http.Error(w, "error checking department", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "department not found", http.StatusNotFound)
		return
	}

	switch mode {
	case "cascade":
		if err := h.departmentRepo.DeleteCascade(r.Context(), id); err != nil {
			http.Error(w, "error deleting department", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	case "reassign":
		reassignToStr := r.URL.Query().Get("reassign_to_department_id")
		if reassignToStr == "" {
			http.Error(w, "reassign_to_department_id is required", http.StatusBadRequest)
			return
		}

		reassignToID, err := strconv.ParseInt(reassignToStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid reassign_to_department_id", http.StatusBadRequest)
			return
		}

		if reassignToID == id {
			http.Error(w, "cannot reassign employees to deleted department", http.StatusBadRequest)
			return
		}

		targetExists, err := h.departmentRepo.Exists(r.Context(), reassignToID)
		if err != nil {
			http.Error(w, "error checking target department", http.StatusInternalServerError)
			return
		}

		if !targetExists {
			http.Error(w, "target department not found", http.StatusNotFound)
			return
		}

		isDescendant, err := h.departmentRepo.IsDescendant(r.Context(), id, reassignToID)
		if err != nil {
			http.Error(w, "error checking department tree", http.StatusInternalServerError)
			return
		}

		if isDescendant {
			http.Error(w, "cannot reassign employees to department inside deleted subtree", http.StatusConflict)
			return
		}

		if err := h.departmentRepo.DeleteReassign(r.Context(), id, reassignToID); err != nil {
			http.Error(w, "error deleting department", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "invalid delete mode", http.StatusBadRequest)
	}
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
		return 0, errors.New("depth must be <= 5")
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

func (o *OptionalInt64) UnmarshalJSON(data []byte) error {
	o.Set = true

	raw := strings.TrimSpace(string(data))
	if raw == "null" {
		o.Value = nil
		return nil
	}

	var value int64
	if err := json.Unmarshal(data, &value); err != nil {
		return errors.New("parent_id must be integer or null")
	}

	o.Value = &value

	return nil
}
