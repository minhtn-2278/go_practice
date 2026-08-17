package models

import "time"

type Employee struct {
	ID           int64     `json:"id"`
	Age          int       `json:"age"`
	Name         string    `json:"name"`
	Position     string    `json:"position"`
	Salary       int       `json:"salary"`
	DepartmentID int       `json:"department_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
