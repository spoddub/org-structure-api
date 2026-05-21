package main

import (
	"log"
	"net/http"
	"org-structure-api/internal/db"
	"org-structure-api/internal/handler"
	"org-structure-api/internal/repository"
	"os"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/org_structure_api?sslmode=disable"
	}

	database, err := db.Open(databaseURL)
	if err != nil {
		log.Fatal(err)
	}

	departmentRepo := repository.NewDepartmentRepository(database)
	employeeRepo := repository.NewEmployeeRepository(database)

	h := handler.NewHandler(departmentRepo, employeeRepo)
	r := handler.NewRouter(h)

	log.Println("Starting server on :8080")

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
