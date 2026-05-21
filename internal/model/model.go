package model

import "time"

type Department struct {
	ID        int64     `gorm:"primaryKey;column:id" json:"id"`
	Name      string    `gorm:"column:name" json:"name"`
	ParentID  *int64    `gorm:"column:parent_id" json:"parent_id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`

	Children  []Department `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Employees []Employee   `gorm:"foreignKey:DepartmentID" json:"employees,omitempty"`
}

func (Department) TableName() string {
	return "departments"
}

type Employee struct {
	ID           int64      `gorm:"primaryKey;column:id" json:"id"`
	DepartmentID int64      `gorm:"column:department_id" json:"department_id"`
	FullName     string     `gorm:"column:full_name" json:"full_name"`
	Position     string     `gorm:"column:position" json:"position"`
	HiredAt      *time.Time `gorm:"column:hired_at" json:"hired_at"`
	CreatedAt    time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (Employee) TableName() string {
	return "employees"
}
