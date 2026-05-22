package handler

import "net/http"

func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /departments/", h.createDepartment)
	mux.HandleFunc("POST /departments/{id}/employees/", h.createEmployee)
	mux.HandleFunc("GET /departments/{id}", h.getDepartmentByID)
	mux.HandleFunc("PATCH /departments/{id}", h.updateDepartment)
	mux.HandleFunc("DELETE /departments/{id}", h.deleteDepartment)

	return mux
}
