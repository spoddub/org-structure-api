package handler

import "net/http"

func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST/departments/", h.createDepartment)
	mux.HandleFunc("POST/departments/{id}/employees", h.createEmployee)

	return mux
}
