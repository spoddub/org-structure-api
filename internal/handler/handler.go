package handler

import "org-structure-api/internal/repository"

type Handler struct {
	departmentRepo *repository.DepartmentRepository
	employeeRepo   *repository.EmployeeRepository
}

func NewHandler(departmentRepo *repository.DepartmentRepository, employeeRepo *repository.EmployeeRepository) *Handler {
	return &Handler{
		departmentRepo: departmentRepo,
		employeeRepo:   employeeRepo,
	}
}
