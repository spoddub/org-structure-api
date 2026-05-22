package main

import (
	"log"
	"net/http"
	_ "org-structure-api/docs"
	"org-structure-api/internal/db"
	"org-structure-api/internal/handler"
	"org-structure-api/internal/repository"
	"os"
)

// @title Org Structure API
// @version 1.0
// @description REST API for managing departments, employees, and organizational structure tree.
// @host localhost:8080
// @BasePath /
func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/org_structure_api?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	database, err := db.Open(databaseURL)
	if err != nil {
		log.Fatal(err)
	}

	departmentRepo := repository.NewDepartmentRepository(database)
	employeeRepo := repository.NewEmployeeRepository(database)

	h := handler.NewHandler(departmentRepo, employeeRepo)
	r := handler.NewRouter(h)

	addr := ":" + port

	log.Println("Starting server on", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
