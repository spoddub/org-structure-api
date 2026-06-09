package main

import (
	"log"
	"net/http"
	"os"

	_ "org-structure-api/docs"
	"org-structure-api/internal/analytics"
	"org-structure-api/internal/db"
	"org-structure-api/internal/handler"
	"org-structure-api/internal/repository"
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

	analyticsWriter, err := openAnalyticsWriter()
	if err != nil {
		log.Fatal(err)
	}

	if analyticsWriter != nil {
		log.Println("ClickHouse analytics connected")
	}

	departmentRepo := repository.NewDepartmentRepository(database)
	employeeRepo := repository.NewEmployeeRepository(database)

	h := handler.NewHandler(departmentRepo, employeeRepo, analyticsWriter)
	r := handler.NewRouter(h)

	addr := ":" + port

	log.Println("Starting server on", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func openAnalyticsWriter() (analytics.Writer, error) {
	addr := os.Getenv("CLICKHOUSE_ADDR")
	database := os.Getenv("CLICKHOUSE_DATABASE")
	user := os.Getenv("CLICKHOUSE_USER")
	password := os.Getenv("CLICKHOUSE_PASSWORD")

	if addr == "" || database == "" || user == "" {
		log.Println("ClickHouse analytics disabled")
		return nil, nil
	}

	return analytics.NewClickHouseWriter(addr, database, user, password)
}
