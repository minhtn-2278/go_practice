package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"employee-management/internal/models"
	"employee-management/internal/services"
)

type EmployeeRepository struct {
	db                   *sql.DB
	departmentRepository *DepartmentRepository
}

var _ services.EmployeeRepository = (*EmployeeRepository)(nil)

func NewEmployeeRepository(db *sql.DB, departmentRepository *DepartmentRepository) (*EmployeeRepository, error) {
	if db == nil {
		return nil, errors.New("database is required")
	}
	if departmentRepository == nil {
		return nil, errors.New("department repository is required")
	}

	return &EmployeeRepository{
		db:                   db,
		departmentRepository: departmentRepository,
	}, nil
}

func (r *EmployeeRepository) Create(ctx context.Context, employee models.Employee) (storedEmployee models.Employee, returnErr error) {
	if err := ctx.Err(); err != nil {
		return models.Employee{}, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Employee{}, err
	}

	committed := false
	defer func() {
		if committed {
			return
		}

		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) && returnErr == nil {
			returnErr = rollbackErr
		}
	}()

	if err := r.departmentRepository.ExistsByIDTx(ctx, tx, int64(employee.DepartmentID)); err != nil {
		return models.Employee{}, err
	}

	const createEmployeeQuery = `
		INSERT INTO employees (age, name, position, salary, department_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	statement, err := tx.PrepareContext(ctx, createEmployeeQuery)
	if err != nil {
		return models.Employee{}, err
	}
	defer statement.Close()

	err = statement.QueryRowContext(
		ctx,
		employee.Age,
		employee.Name,
		employee.Position,
		employee.Salary,
		employee.DepartmentID,
	).Scan(&employee.ID, &employee.CreatedAt, &employee.UpdatedAt)
	if err != nil {
		return models.Employee{}, err
	}

	if err = tx.Commit(); err != nil {
		return models.Employee{}, err
	}
	committed = true

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

func (r *EmployeeRepository) List(ctx context.Context, pagination models.Pagination, keyword string) ([]models.Employee, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	keyword = strings.TrimSpace(keyword)
	filterArgs := make([]any, 0, 1)
	whereClause := ""
	if keyword != "" {
		filterArgs = append(filterArgs, "%"+keyword+"%")
		whereClause = " WHERE name ILIKE $1 OR position ILIKE $1"
	}

	countQuery := "SELECT COUNT(*) FROM employees" + whereClause
	listQuery := "SELECT id, age, name, position, salary, department_id, created_at, updated_at FROM employees" +
		whereClause +
		" ORDER BY id LIMIT $" + strconv.Itoa(len(filterArgs)+1) +
		" OFFSET $" + strconv.Itoa(len(filterArgs)+2)
	listArgs := append([]any(nil), filterArgs...)
	listArgs = append(listArgs, pagination.Limit, pagination.Offset())

	countStatement, err := r.db.PrepareContext(ctx, countQuery)
	if err != nil {
		return nil, 0, err
	}
	defer countStatement.Close()

	var total int64
	if err := countStatement.QueryRowContext(ctx, filterArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	statement, err := r.db.PrepareContext(ctx, listQuery)
	if err != nil {
		return nil, 0, err
	}
	defer statement.Close()

	rows, err := statement.QueryContext(ctx, listArgs...)
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

func (r *EmployeeRepository) Update(ctx context.Context, employee models.Employee) (storedEmployee models.Employee, returnErr error) {
	if err := ctx.Err(); err != nil {
		return models.Employee{}, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Employee{}, err
	}

	committed := false
	defer func() {
		if committed {
			return
		}

		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) && returnErr == nil {
			returnErr = rollbackErr
		}
	}()

	if err := r.departmentRepository.ExistsByIDTx(ctx, tx, int64(employee.DepartmentID)); err != nil {
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

	statement, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return models.Employee{}, err
	}
	defer statement.Close()

	err = statement.QueryRowContext(
		ctx,
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

	if err = tx.Commit(); err != nil {
		return models.Employee{}, err
	}
	committed = true

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
