package postgres

import (
	"context"
	"database/sql"
	"errors"

	"employee-management/internal/models"
	"employee-management/internal/services"
)

type DepartmentRepository struct {
	db *sql.DB
}

func NewDepartmentRepository(db *sql.DB) (*DepartmentRepository, error) {
	if db == nil {
		return nil, errors.New("database is required")
	}

	return &DepartmentRepository{db: db}, nil
}

func (r *DepartmentRepository) ExistsByIDTx(ctx context.Context, tx *sql.Tx, id int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	const query = `
		SELECT id
		FROM departments
		WHERE id = $1`
	statement, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer statement.Close()

	var departmentID int64
	if err := statement.QueryRowContext(ctx, id).Scan(&departmentID); errors.Is(err, sql.ErrNoRows) {
		return services.ErrDepartmentNotFound
	} else if err != nil {
		return err
	}

	return nil
}

func (r *DepartmentRepository) Create(ctx context.Context, department models.Department) (models.Department, error) {
	if err := ctx.Err(); err != nil {
		return models.Department{}, err
	}

	const query = `
		INSERT INTO departments (name)
		VALUES ($1)
		RETURNING id, created_at, updated_at`

	statement, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return models.Department{}, err
	}
	defer statement.Close()

	err = statement.QueryRowContext(ctx, department.Name).Scan(
		&department.ID,
		&department.CreatedAt,
		&department.UpdatedAt,
	)
	if err != nil {
		return models.Department{}, err
	}

	return department, nil
}

func (r *DepartmentRepository) List(ctx context.Context, pagination models.Pagination) ([]models.Department, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	const countQuery = `SELECT COUNT(*) FROM departments`
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	const query = `
		SELECT id, name, created_at, updated_at
		FROM departments
		ORDER BY id
		LIMIT $1 OFFSET $2`

	statement, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer statement.Close()

	rows, err := statement.QueryContext(ctx, pagination.Limit, pagination.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	departments := make([]models.Department, 0)
	for rows.Next() {
		var department models.Department
		if err := rows.Scan(
			&department.ID,
			&department.Name,
			&department.CreatedAt,
			&department.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		departments = append(departments, department)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return departments, total, nil
}

func (r *DepartmentRepository) ListEmployee(ctx context.Context, departmentID int64, pagination models.Pagination) ([]models.Employee, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	const countQuery = `SELECT COUNT(*) FROM employees WHERE department_id = $1`
	var total int64
	countStatement, err := r.db.PrepareContext(ctx, countQuery)
	if err != nil {
		return nil, 0, err
	}
	defer countStatement.Close()

	if err := countStatement.QueryRowContext(ctx, departmentID).Scan(&total); err != nil {
		return nil, 0, err
	}

	const query = `
		SELECT id, age, name, position, salary, department_id, created_at, updated_at
		FROM employees
		WHERE department_id = $1
		ORDER BY id
		LIMIT $2 OFFSET $3`

	statement, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer statement.Close()

	rows, err := statement.QueryContext(
		ctx,
		query,
		departmentID,
		pagination.Limit,
		pagination.Offset(),
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	employees := make([]models.Employee, 0)
	for rows.Next() {
		var employee models.Employee
		if err := rows.Scan(
			&employee.ID,
			&employee.Age,
			&employee.Name,
			&employee.Position,
			&employee.Salary,
			&employee.DepartmentID,
			&employee.CreatedAt,
			&employee.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		employees = append(employees, employee)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return employees, total, nil
}
