package handlers

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"employee-management/internal/models"
	"employee-management/internal/services"
	"employee-management/internal/utils"
)

type EmployeeHandler struct {
	service  services.EmployeeUseCase
	exportMu sync.Mutex
}

type updateEmployeeRequest struct {
	Age          *int    `json:"age"`
	Name         *string `json:"name"`
	Position     *string `json:"position"`
	Salary       *int    `json:"salary"`
	DepartmentID *int    `json:"department_id"`
}

const (
	employeeExportDir    = "exports"
	employeeJSONFilename = "employees.json"
	employeeCSVFilename  = "employees.csv"
)

type employeeExportResponse struct {
	Message  string `json:"message"`
	JSONFile string `json:"json_file"`
	CSVFile  string `json:"csv_file"`
}

func NewEmployeeHandler(service services.EmployeeUseCase) (*EmployeeHandler, error) {
	if service == nil {
		return nil, errors.New("employee service is required")
	}

	return &EmployeeHandler{service: service}, nil
}

func (h *EmployeeHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /employees", h.create)
	mux.HandleFunc("GET /employees", h.list)
	mux.HandleFunc("POST /employees/export", h.export)
	mux.HandleFunc("GET /employees/{id}", h.detail)
	mux.HandleFunc("PUT /employees/{id}", h.update)
	mux.HandleFunc("DELETE /employees/{id}", h.delete)
}

func (h *EmployeeHandler) create(w http.ResponseWriter, r *http.Request) {
	var employee models.Employee
	if err := utils.DecodeJSON(r, &employee); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	createdEmployee, err := h.service.Create(r.Context(), employee)
	if err != nil {
		writeEmployeeServiceError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, createdEmployee)
}

func (h *EmployeeHandler) list(w http.ResponseWriter, r *http.Request) {
	pagination, err := utils.ParsePagination(r.URL.Query())
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, models.ErrInvalidPagination.Error())
		return
	}

	keyword := r.URL.Query().Get("keyword")
	employees, err := h.service.List(r.Context(), &pagination, keyword)
	if err != nil {
		writeEmployeeServiceError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, employees)
}

func (h *EmployeeHandler) detail(w http.ResponseWriter, r *http.Request) {
	id, err := employeeID(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid employee ID")
		return
	}

	employee, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writeEmployeeServiceError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, employee)
}

func (h *EmployeeHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := employeeID(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid employee ID")
		return
	}

	employee, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writeEmployeeServiceError(w, err)
		return
	}

	var updateEmployeeParams updateEmployeeRequest
	if err := utils.DecodeJSON(r, &updateEmployeeParams); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if updateEmployeeParams.Age != nil {
		employee.Age = *updateEmployeeParams.Age
	}
	if updateEmployeeParams.Name != nil {
		employee.Name = *updateEmployeeParams.Name
	}
	if updateEmployeeParams.Position != nil {
		employee.Position = *updateEmployeeParams.Position
	}
	if updateEmployeeParams.Salary != nil {
		employee.Salary = *updateEmployeeParams.Salary
	}
	if updateEmployeeParams.DepartmentID != nil {
		employee.DepartmentID = *updateEmployeeParams.DepartmentID
	}

	updatedEmployee, err := h.service.Update(r.Context(), employee)
	if err != nil {
		writeEmployeeServiceError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, updatedEmployee)
}

func (h *EmployeeHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := employeeID(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid employee ID")
		return
	}

	err = h.service.Delete(r.Context(), id)
	if err != nil {
		writeEmployeeServiceError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

func (h *EmployeeHandler) export(w http.ResponseWriter, r *http.Request) {
	h.exportMu.Lock()
	defer h.exportMu.Unlock()

	result, err := h.service.List(r.Context(), nil, r.URL.Query().Get("keyword"))
	if err != nil {
		writeEmployeeServiceError(w, err)
		return
	}
	employees := result.Data

	var waitGroup sync.WaitGroup
	var mutex sync.Mutex
	exportErrors := make([]error, 0, 2)

	addExportError := func(err error) {
		if err == nil {
			return
		}

		mutex.Lock()
		exportErrors = append(exportErrors, err)
		mutex.Unlock()
	}

	waitGroup.Add(2)
	go func() {
		defer waitGroup.Done()
		addExportError(writeEmployeesJSON(employeeExportDir, employees))
	}()
	go func() {
		defer waitGroup.Done()
		addExportError(writeEmployeesCSV(employeeExportDir, employees))
	}()
	waitGroup.Wait()

	if len(exportErrors) > 0 {
		utils.WriteError(w, http.StatusInternalServerError, "Could not export employees")
		return
	}

	utils.WriteJSON(w, http.StatusOK, employeeExportResponse{
		Message:  "Employees exported successfully",
		JSONFile: employeeJSONFilename,
		CSVFile:  employeeCSVFilename,
	})
}

func writeEmployeesJSON(exportDir string, employees []models.Employee) error {
	if err := os.MkdirAll(exportDir, 0o755); err != nil {
		return fmt.Errorf("create export directory: %w", err)
	}

	path := filepath.Join(exportDir, employeeJSONFilename)
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create JSON export: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(employees); err != nil {
		return fmt.Errorf("encode JSON export: %w", err)
	}

	return nil
}

func writeEmployeesCSV(exportDir string, employees []models.Employee) error {
	if err := os.MkdirAll(exportDir, 0o755); err != nil {
		return fmt.Errorf("create export directory: %w", err)
	}

	path := filepath.Join(exportDir, employeeCSVFilename)
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create CSV export: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.Write([]string{
		"id",
		"age",
		"name",
		"position",
		"salary",
		"department_id",
		"created_at",
		"updated_at",
	}); err != nil {
		return fmt.Errorf("write CSV header: %w", err)
	}

	for _, employee := range employees {
		if err := writer.Write([]string{
			strconv.FormatInt(employee.ID, 10),
			strconv.Itoa(employee.Age),
			employee.Name,
			employee.Position,
			strconv.Itoa(employee.Salary),
			strconv.Itoa(employee.DepartmentID),
			employee.CreatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
			employee.UpdatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
		}); err != nil {
			return fmt.Errorf("write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("flush CSV export: %w", err)
	}

	return nil
}

func employeeID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func writeEmployeeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidEmployee):
		utils.WriteError(w, http.StatusBadRequest, services.ErrInvalidEmployee.Error())
	case errors.Is(err, services.ErrEmployeeNotFound):
		utils.WriteError(w, http.StatusNotFound, services.ErrEmployeeNotFound.Error())
	case errors.Is(err, services.ErrInvalidEmployeeAge):
		utils.WriteError(w, http.StatusBadRequest, services.ErrInvalidEmployeeAge.Error())
	case errors.Is(err, services.ErrInvalidEmployeeSalary):
		utils.WriteError(w, http.StatusBadRequest, services.ErrInvalidEmployeeSalary.Error())
	case errors.Is(err, services.ErrDepartmentNotFound):
		utils.WriteError(w, http.StatusNotFound, services.ErrDepartmentNotFound.Error())
	case errors.Is(err, models.ErrInvalidPagination):
		utils.WriteError(w, http.StatusBadRequest, models.ErrInvalidPagination.Error())
	default:
		utils.WriteError(w, http.StatusInternalServerError, "Internal server error")
	}
}
