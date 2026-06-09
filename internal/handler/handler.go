package handler

import (
	"context"
	"log"

	"org-structure-api/internal/analytics"
	"org-structure-api/internal/repository"
)

type Handler struct {
	departmentRepo  *repository.DepartmentRepository
	employeeRepo    *repository.EmployeeRepository
	analyticsWriter analytics.Writer
}

func NewHandler(
	departmentRepo *repository.DepartmentRepository,
	employeeRepo *repository.EmployeeRepository,
	analyticsWriters ...analytics.Writer,
) *Handler {
	var analyticsWriter analytics.Writer
	if len(analyticsWriters) > 0 {
		analyticsWriter = analyticsWriters[0]
	}

	return &Handler{
		departmentRepo:  departmentRepo,
		employeeRepo:    employeeRepo,
		analyticsWriter: analyticsWriter,
	}
}

func (h *Handler) trackAnalytics(ctx context.Context, event analytics.Event) {
	if h.analyticsWriter == nil {
		return
	}

	if err := h.analyticsWriter.Track(ctx, event); err != nil {
		log.Println("error writing analytics event:", err)
	}
}
