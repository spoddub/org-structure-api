-- +goose Up
CREATE TABLE departments
(
    department_id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    parent_id BIGINT REFERENCES departments(department_id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT departments_name_not_blank CHECK (btrim(name) != ''),
    CONSTRAINT departments_parent_not_self CHECK (parent_id IS NULL OR parent_id != department_id)
);

CREATE UNIQUE INDEX departments_unique_name_inside_parent_idx
    ON departments (parent_id, lower(name))
    WHERE parent_id IS NOT NULL;

CREATE UNIQUE INDEX departments_unique_root_name_idx
    ON departments (lower(name))
    WHERE parent_id IS NULL;

CREATE INDEX departments_parent_id_idx
    ON departments (parent_id);

CREATE TABLE employees
(
    employee_id BIGSERIAL PRIMARY KEY,
    department_id BIGINT NOT NULL REFERENCES departments(department_id) ON DELETE CASCADE,
    full_name VARCHAR(200) NOT NULL,
    position VARCHAR(200) NOT NULL,
    hired_at DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT employees_full_name_not_blank CHECK (btrim(full_name) != ''),
    CONSTRAINT employees_position_not_blank CHECK (btrim(position) != '')
);

CREATE INDEX employees_department_id_idx
    ON employees (department_id);

CREATE INDEX employees_full_name_idx
    ON employees (full_name);

-- +goose Down
DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS departments;