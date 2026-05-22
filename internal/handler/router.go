package handler

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /departments/", h.createDepartment)
	mux.HandleFunc("POST /departments/{id}/employees/", h.createEmployee)
	mux.HandleFunc("GET /departments/{id}", h.getDepartmentByID)
	mux.HandleFunc("PATCH /departments/{id}", h.updateDepartment)
	mux.HandleFunc("DELETE /departments/{id}", h.deleteDepartment)
	mux.HandleFunc("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	return mux
}
