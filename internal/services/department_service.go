package services

import (
	"context"
	"errors"
	"strings"

	"employee-management/internal/models"
)

var (
	ErrInvalidDepartment = errors.New("Invalid department")
)

type DepartmentRepository interface {
	Create(ctx context.Context, department models.Department) (models.Department, error)
	List(ctx context.Context, pagination models.Pagination) ([]models.Department, int64, error)
	ListEmployee(ctx context.Context, departmentID int64, pagination models.Pagination) ([]models.Employee, int64, error)
}

type DepartmentUseCase interface {
	Create(ctx context.Context, department models.Department) (models.Department, error)
	List(ctx context.Context, pagination models.Pagination) (models.PaginatedResult[models.Department], error)
	ListEmployee(ctx context.Context, departmentID int64, pagination models.Pagination) (models.PaginatedResult[models.Employee], error)
}

type DepartmentService struct {
	repository DepartmentRepository
}

func NewDepartmentService(repository DepartmentRepository) (*DepartmentService, error) {
	if repository == nil {
		return nil, errors.New("department repository is required")
	}

	return &DepartmentService{repository: repository}, nil
}

func (s *DepartmentService) Create(ctx context.Context, department models.Department) (models.Department, error) {
	if err := validateDepartmentContext(ctx); err != nil {
		return models.Department{}, err
	}
	if err := validateDepartment(department); err != nil {
		return models.Department{}, err
	}

	return s.repository.Create(ctx, normalizeDepartment(department))
}

func (s *DepartmentService) List(ctx context.Context, pagination models.Pagination) (models.PaginatedResult[models.Department], error) {
	if err := validateDepartmentContext(ctx); err != nil {
		return models.PaginatedResult[models.Department]{}, err
	}

	pagination, err := models.NewPagination(pagination.Page, pagination.Limit)
	if err != nil {
		return models.PaginatedResult[models.Department]{}, err
	}

	departments, total, err := s.repository.List(ctx, pagination)
	if err != nil {
		return models.PaginatedResult[models.Department]{}, err
	}

	return models.NewPaginatedResult(departments, pagination, total), nil
}

func (s *DepartmentService) ListEmployee(ctx context.Context, departmentID int64, pagination models.Pagination) (models.PaginatedResult[models.Employee], error) {
	if err := validateDepartmentContext(ctx); err != nil {
		return models.PaginatedResult[models.Employee]{}, err
	}
	if departmentID <= 0 {
		return models.PaginatedResult[models.Employee]{}, ErrInvalidDepartment
	}

	pagination, err := models.NewPagination(pagination.Page, pagination.Limit)
	if err != nil {
		return models.PaginatedResult[models.Employee]{}, err
	}

	employees, total, err := s.repository.ListEmployee(ctx, departmentID, pagination)
	if err != nil {
		return models.PaginatedResult[models.Employee]{}, err
	}

	return models.NewPaginatedResult(employees, pagination, total), nil
}

func validateDepartmentContext(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is required")
	}

	return ctx.Err()
}

func validateDepartment(department models.Department) error {
	if strings.TrimSpace(department.Name) == "" {
		return ErrInvalidDepartment
	}

	return nil
}

func normalizeDepartment(department models.Department) models.Department {
	department.Name = strings.TrimSpace(department.Name)
	return department
}
