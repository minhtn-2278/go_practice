package postgres

import (
	"context"
	"database/sql"
	"errors"

	"employee-management/internal/models"
	"employee-management/internal/services"
)

type EmployeeRepository struct {
	db *sql.DB
}

var _ services.EmployeeRepository = (*EmployeeRepository)(nil)

func NewEmployeeRepository(db *sql.DB) (*EmployeeRepository, error) {
	if db == nil {
		return nil, errors.New("database is required")
	}

	return &EmployeeRepository{db: db}, nil
}

func (r *EmployeeRepository) Create(ctx context.Context, employee models.Employee) (models.Employee, error) {
	if err := ctx.Err(); err != nil {
		return models.Employee{}, err
	}

	const query = `
		INSERT INTO employees (age, name, position, salary, department_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	statement, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return models.Employee{}, err
	}
	defer statement.Close()

	err = statement.QueryRowContext(
		ctx,
		query,
		employee.Age,
		employee.Name,
		employee.Position,
		employee.Salary,
		employee.DepartmentID,
	).Scan(&employee.ID, &employee.CreatedAt, &employee.UpdatedAt)
	if err != nil {
		return models.Employee{}, err
	}

	return employee, nil
}

func (r *EmployeeRepository) GetByID(ctx context.Context, id int64) (models.Employee, error) {
	if err := ctx.Err(); err != nil {
		return models.Employee{}, err
	}

	const query = `
		SELECT id, age, name, position, salary, department_id, created_at, updated_at
		FROM employees
		WHERE id = $1`

	statement, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return models.Employee{}, err
	}
	defer statement.Close()

	var employee models.Employee
	err = statement.QueryRowContext(ctx, id).Scan(
		&employee.ID,
		&employee.Age,
		&employee.Name,
		&employee.Position,
		&employee.Salary,
		&employee.DepartmentID,
		&employee.CreatedAt,
		&employee.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Employee{}, services.ErrEmployeeNotFound
	}
	if err != nil {
		return models.Employee{}, err
	}

	return employee, nil
}

func (r *EmployeeRepository) List(ctx context.Context, pagination models.Pagination) ([]models.Employee, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	const countQuery = `SELECT COUNT(*) FROM employees`
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	const listQuery = `
		SELECT id, age, name, position, salary, department_id, created_at, updated_at
		FROM employees
		ORDER BY id
		LIMIT $1 OFFSET $2`

	statement, err := r.db.PrepareContext(ctx, listQuery)
	if err != nil {
		return nil, 0, err
	}
	defer statement.Close()

	rows, err := statement.QueryContext(ctx, pagination.Limit, pagination.Offset())
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

func (r *EmployeeRepository) Update(ctx context.Context, employee models.Employee) (models.Employee, error) {
	if err := ctx.Err(); err != nil {
		return models.Employee{}, err
	}

	const query = `
		UPDATE employees
		SET age = $1,
			name = $2,
			position = $3,
			salary = $4,
			department_id = $5,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
		RETURNING id, created_at, updated_at`

	statement, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return models.Employee{}, err
	}
	defer statement.Close()

	err = statement.QueryRowContext(
		ctx,
		query,
		employee.Age,
		employee.Name,
		employee.Position,
		employee.Salary,
		employee.DepartmentID,
		employee.ID,
	).Scan(&employee.ID, &employee.CreatedAt, &employee.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Employee{}, services.ErrEmployeeNotFound
	}
	if err != nil {
		return models.Employee{}, err
	}

	return employee, nil
}

func (r *EmployeeRepository) Delete(ctx context.Context, id int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	const query = `DELETE FROM employees WHERE id = $1`
	statement, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer statement.Close()

	result, err := statement.ExecContext(ctx, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return services.ErrEmployeeNotFound
	}

	return nil
}
