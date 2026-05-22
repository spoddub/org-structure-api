## Org Structure API

`Org Structure API` is a REST API service written in Go using the standard `net/http` package.

It manages an organizational structure with departments and employees. Departments can be nested through `parent_id`, so the API can return a department tree with employees and child departments.

The service supports department creation, employee creation, department tree retrieval, department moving, cascade deletion and employee reassignment.

---

## Features

- Create departments
- Create employees inside departments
- Get department details with employees and child departments
- Control tree depth with the `depth` query parameter
- Optionally exclude employees with `include_employees=false`
- Update department name
- Move a department to another parent
- Move a department to root with `parent_id: null`
- Prevent self-parenting
- Prevent cycles in the department tree
- Delete department in `cascade` mode
- Delete department in `reassign` mode
- Unique department names inside the same parent
- PostgreSQL storage
- Database migrations with goose
- GORM repository layer
- Swagger API documentation
- Handler and repository tests
- Local and full Docker Compose run
- Makefile commands for common development tasks

---

## Tools used

| Tool | What it is used for |
| --- | --- |
| [Go](https://go.dev/) | Language and toolchain |
| [net/http](https://pkg.go.dev/net/http) | HTTP server and routing |
| [GORM](https://gorm.io/) | ORM and database access |
| [PostgreSQL](https://www.postgresql.org/) | Persistent storage |
| [goose](https://github.com/pressly/goose) | Database migrations |
| [swaggo/swag](https://github.com/swaggo/swag) | Swagger documentation generation |
| [http-swagger](https://github.com/swaggo/http-swagger) | Swagger UI for `net/http` |
| [Docker](https://www.docker.com/) | Containerized application runtime |
| [Docker Compose](https://docs.docker.com/compose/) | Local app and PostgreSQL environment |
| [Make](https://www.gnu.org/software/make/) | Common development commands |
| [golangci-lint](https://golangci-lint.run/) | Go linter |
| [testing](https://pkg.go.dev/testing) | Unit and handler tests |
| [testify](https://github.com/stretchr/testify) | Test assertions |
| [SQLite](https://www.sqlite.org/) | In-memory repository tests |

---

## API

### Swagger

Swagger UI is available after starting the app:

```text
http://localhost:8080/swagger/
```

Swagger JSON:

```text
http://localhost:8080/swagger/doc.json
```

Generate Swagger docs:

```bash
make swagger
```

---

### Departments

- `POST /departments/` - create a department
- `GET /departments/{id}` - get department details, employees and child departments
- `PATCH /departments/{id}` - update department name or parent
- `DELETE /departments/{id}` - delete department with `cascade` or `reassign` mode

### Employees

- `POST /departments/{id}/employees/` - create an employee inside a department

---

## Business rules

- Department `name` must not be empty.
- Department `name` length is limited to `200` characters.
- Department names must be unique inside the same `parent_id`.
- The same department name is allowed under different parents.
- `parent_id` can be `null`.
- `parent_id` is used to build the department tree.
- A department cannot be parent of itself.
- A department cannot be moved inside its own subtree.
- Employee `full_name` must not be empty.
- Employee `position` must not be empty.
- Employee `hired_at` is optional.
- Employee `hired_at` uses `YYYY-MM-DD` format.
- `DELETE mode=cascade` deletes the department, child departments and employees.
- `DELETE mode=reassign` moves employees to another department before deleting the department subtree.

---

## Date format

The API accepts employee hire dates in `YYYY-MM-DD` format.

Examples:

```text
2026-05-21
2025-01-10
```

Example request field:

```json
{
  "hired_at": "2026-05-21"
}
```

`hired_at` is optional and can be omitted or passed as `null`.

---

## Create department

Endpoint:

```text
POST /departments/
```

Create root department:

```bash
curl -X POST http://localhost:8080/departments/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Engineering"
  }'
```

Example response:

```json
{
  "id": 1,
  "name": "Engineering",
  "parent_id": null,
  "created_at": "2026-05-22T14:05:32.951922Z",
  "updated_at": "2026-05-22T14:05:32.951922Z"
}
```

Create child department:

```bash
curl -X POST http://localhost:8080/departments/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Backend",
    "parent_id": 1
  }'
```

If a department with the same name already exists inside the same parent, the API returns:

```text
409 Conflict
```

---

## Create employee

Endpoint:

```text
POST /departments/{id}/employees/
```

Example:

```bash
curl -X POST http://localhost:8080/departments/1/employees/ \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Sergei Poddubny",
    "position": "Backend Developer",
    "hired_at": "2026-05-21"
  }'
```

Example response:

```json
{
  "id": 1,
  "department_id": 1,
  "full_name": "Sergei Poddubny",
  "position": "Backend Developer",
  "hired_at": "2026-05-21T00:00:00Z",
  "created_at": "2026-05-22T14:05:51.755399Z",
  "updated_at": "2026-05-22T14:05:51.755399Z"
}
```

If the department does not exist, the API returns:

```text
404 Not Found
```

---

## Get department tree

Endpoint:

```text
GET /departments/{id}
```

Example:

```bash
curl http://localhost:8080/departments/1
```

Example response:

```json
{
  "department": {
    "id": 1,
    "name": "Engineering",
    "parent_id": null,
    "created_at": "2026-05-22T14:05:32.951922Z",
    "updated_at": "2026-05-22T14:05:32.951922Z"
  },
  "employees": [
    {
      "id": 1,
      "department_id": 1,
      "full_name": "Sergei Poddubny",
      "position": "Backend Developer",
      "hired_at": "2026-05-21T00:00:00Z",
      "created_at": "2026-05-22T14:05:51.755399Z",
      "updated_at": "2026-05-22T14:05:51.755399Z"
    }
  ],
  "children": [
    {
      "department": {
        "id": 2,
        "name": "Backend",
        "parent_id": 1,
        "created_at": "2026-05-22T14:05:46.002847Z",
        "updated_at": "2026-05-22T14:05:46.002847Z"
      },
      "employees": [],
      "children": []
    }
  ]
}
```

Optional query parameters:

- `depth` - tree depth, default `1`, max `5`
- `include_employees` - include employees in response, default `true`

Examples:

```bash
curl "http://localhost:8080/departments/1?depth=0"
```

```bash
curl "http://localhost:8080/departments/1?include_employees=false"
```

Invalid depth:

```bash
curl "http://localhost:8080/departments/1?depth=6"
```

Response:

```text
400 Bad Request
```

---

## Update department

Endpoint:

```text
PATCH /departments/{id}
```

Update name:

```bash
curl -X PATCH http://localhost:8080/departments/2 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Backend Platform"
  }'
```

Move department under another department:

```bash
curl -X PATCH http://localhost:8080/departments/2 \
  -H "Content-Type: application/json" \
  -d '{
    "parent_id": 1
  }'
```

Move department to root:

```bash
curl -X PATCH http://localhost:8080/departments/2 \
  -H "Content-Type: application/json" \
  -d '{
    "parent_id": null
  }'
```

Example response:

```json
{
  "id": 2,
  "name": "Backend Platform",
  "parent_id": null,
  "created_at": "2026-05-22T14:05:46.002847Z",
  "updated_at": "2026-05-22T14:06:23.098409Z"
}
```

Invalid self-parenting:

```bash
curl -X PATCH http://localhost:8080/departments/2 \
  -H "Content-Type: application/json" \
  -d '{
    "parent_id": 2
  }'
```

Response:

```text
400 Bad Request
```

If an update creates a duplicate name inside the same parent, the API returns:

```text
409 Conflict
```

---

## Delete department

Endpoint:

```text
DELETE /departments/{id}
```

The endpoint requires the `mode` query parameter.

### Cascade delete

Cascade mode deletes the department, all child departments and all employees inside this subtree.

```bash
curl -i -X DELETE "http://localhost:8080/departments/3?mode=cascade"
```

Successful response:

```text
204 No Content
```

### Reassign delete

Reassign mode moves employees from the deleted department subtree to another department before deleting the department subtree.

```bash
curl -i -X DELETE "http://localhost:8080/departments/5?mode=reassign&reassign_to_department_id=1"
```

Successful response:

```text
204 No Content
```

If `reassign_to_department_id` does not exist, the API returns:

```text
404 Not Found
```

If `reassign_to_department_id` is inside the deleted subtree, the API returns:

```text
409 Conflict
```

---

## Installation and local development

### Requirements

- Go 1.26.3+
- Docker
- Docker Compose
- make
- goose
- swag
- golangci-lint, optional for linting

Install goose:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Install swag:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

---

## Environment variables

For local development:

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/org_structure_api?sslmode=disable
PORT=8080
```

For Docker Compose, the app uses the PostgreSQL service name as host:

```env
DATABASE_URL=postgres://postgres:postgres@postgres:5432/org_structure_api?sslmode=disable
PORT=8080
```

---

## Full Docker run

Build and run the app together with PostgreSQL:

```bash
docker compose up --build
```

The Docker app container applies database migrations automatically before starting the server.

The app listens on:

```text
http://localhost:8080
```

Swagger UI:

```text
http://localhost:8080/swagger/
```

Stop Docker Compose services:

```bash
docker compose down
```

Remove containers and database volume:

```bash
docker compose down -v
```

---

## Local run with PostgreSQL from Docker

Start PostgreSQL only:

```bash
docker compose up -d postgres
```

Apply migrations:

```bash
make migrate-up
```

Run the app locally:

```bash
make run
```

The app listens on:

```text
http://localhost:8080
```

---

## Migrations

Apply migrations:

```bash
make migrate-up
```

Check migration status:

```bash
make migrate-status
```

Rollback the last migration:

```bash
make migrate-down
```

Manual goose command:

```bash
goose -dir migrations postgres "postgres://postgres:postgres@localhost:5432/org_structure_api?sslmode=disable" up
```

---

## Swagger

Generate Swagger documentation:

```bash
make swagger
```

Run the application:

```bash
docker compose up --build
```

Open Swagger UI:

```text
http://localhost:8080/swagger/
```

Swagger JSON:

```text
http://localhost:8080/swagger/doc.json
```

---

## Run tests

Run all tests:

```bash
make test
```

Or directly:

```bash
go test ./...
```

---

## Code quality

Run linter:

```bash
make lint
```

Run linter with automatic fixes:

```bash
make lint-fix
```

---

## Common commands

Generate Swagger docs:

```bash
make swagger
```

Run tests:

```bash
make test
```

Run linter:

```bash
make lint
```

Run linter with fixes:

```bash
make lint-fix
```

Apply migrations:

```bash
make migrate-up
```

Rollback migration:

```bash
make migrate-down
```

Check migration status:

```bash
make migrate-status
```

Run app locally:

```bash
make run
```

Build and run app with PostgreSQL:

```bash
docker compose up --build
```

Stop Docker Compose:

```bash
docker compose down
```

---

## Database schema

The service uses two main tables:

- `departments`
- `employees`

### departments

Main fields:

- `id`
- `name`
- `parent_id`
- `created_at`
- `updated_at`

Important constraints:

- `name` must not be empty
- `name` length is limited to `200`
- `parent_id` references `departments.id`
- `parent_id` can be `null`
- `parent_id` must not equal `id`
- department names are unique within the same `parent_id`

### employees

Main fields:

- `id`
- `department_id`
- `full_name`
- `position`
- `hired_at`
- `created_at`
- `updated_at`

Important constraints:

- `department_id` references `departments.id`
- `full_name` must not be empty
- `position` must not be empty
- `hired_at` is optional
