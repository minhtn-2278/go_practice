package services

import (
	"context"
	"errors"
	"strings"

	"employee-management/internal/models"
)

var (
	ErrInvalidEmployee       = errors.New("Invalid employee")
	ErrInvalidEmployeeAge    = errors.New("Invalid employee age, must be greater than 0")
	ErrInvalidEmployeeSalary = errors.New("Invalid employee salary, must be greater than 0")
	ErrDepartmentNotFound    = errors.New("Department not found")
	ErrEmployeeNotFound      = errors.New("Employee not found")
)

type EmployeeRepository interface {
	Create(ctx context.Context, employee models.Employee) (models.Employee, error)
	GetByID(ctx context.Context, id int64) (models.Employee, error)
	List(ctx context.Context, pagination models.Pagination, keyword string) ([]models.Employee, int64, error)
	Update(ctx context.Context, employee models.Employee) (models.Employee, error)
	Delete(ctx context.Context, id int64) error
}

type EmployeeUseCase interface {
	Create(ctx context.Context, employee models.Employee) (models.Employee, error)
	GetByID(ctx context.Context, id int64) (models.Employee, error)
	List(ctx context.Context, pagination models.Pagination, keyword string) (models.PaginatedResult[models.Employee], error)
	Update(ctx context.Context, employee models.Employee) (models.Employee, error)
	Delete(ctx context.Context, id int64) error
}

type EmployeeService struct {
	repository EmployeeRepository
}

func NewEmployeeService(repository EmployeeRepository) (*EmployeeService, error) {
	if repository == nil {
		return nil, errors.New("employee repository is required")
	}

	return &EmployeeService{repository: repository}, nil
}

func (s *EmployeeService) Create(ctx context.Context, employee models.Employee) (models.Employee, error) {
	if err := validateContext(ctx); err != nil {
		return models.Employee{}, err
	}
	if err := validateEmployee(employee); err != nil {
		return models.Employee{}, err
	}

	return s.repository.Create(ctx, normalizeEmployee(employee))
}

func (s *EmployeeService) GetByID(ctx context.Context, id int64) (models.Employee, error) {
	if err := validateContext(ctx); err != nil {
		return models.Employee{}, err
	}
	if id <= 0 {
		return models.Employee{}, ErrInvalidEmployee
	}

	return s.repository.GetByID(ctx, id)
}

func (s *EmployeeService) List(ctx context.Context, pagination models.Pagination, keyword string) (models.PaginatedResult[models.Employee], error) {
	if err := validateContext(ctx); err != nil {
		return models.PaginatedResult[models.Employee]{}, err
	}

	pagination, err := models.NewPagination(pagination.Page, pagination.Limit)
	if err != nil {
		return models.PaginatedResult[models.Employee]{}, err
	}

	employees, total, err := s.repository.List(ctx, pagination, strings.TrimSpace(keyword))
	if err != nil {
		return models.PaginatedResult[models.Employee]{}, err
	}

	return models.NewPaginatedResult(employees, pagination, total), nil
}

func (s *EmployeeService) Update(ctx context.Context, employee models.Employee) (models.Employee, error) {
	if err := validateContext(ctx); err != nil {
		return models.Employee{}, err
	}
	if employee.ID <= 0 {
		return models.Employee{}, ErrInvalidEmployee
	}
	if err := validateEmployee(employee); err != nil {
		return models.Employee{}, err
	}

	return s.repository.Update(ctx, normalizeEmployee(employee))
}

func (s *EmployeeService) Delete(ctx context.Context, id int64) error {
	if err := validateContext(ctx); err != nil {
		return err
	}
	if id <= 0 {
		return ErrInvalidEmployee
	}

	return s.repository.Delete(ctx, id)
}

func validateContext(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is required")
	}

	return ctx.Err()
}

func validateEmployee(employee models.Employee) error {
	if employee.Age <= 0 {
		return ErrInvalidEmployeeAge
	}

	if employee.Salary <= 0 {
		return ErrInvalidEmployeeSalary
	}

	return nil
}

func normalizeEmployee(employee models.Employee) models.Employee {
	employee.Name = strings.TrimSpace(employee.Name)
	employee.Position = strings.TrimSpace(employee.Position)

	return employee
}
