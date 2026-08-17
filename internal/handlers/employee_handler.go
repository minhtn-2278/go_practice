package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"employee-management/internal/models"
	"employee-management/internal/services"
	"employee-management/internal/utils"
)

type EmployeeHandler struct {
	service services.EmployeeUseCase
}

type updateEmployeeRequest struct {
	Age          *int    `json:"age"`
	Name         *string `json:"name"`
	Position     *string `json:"position"`
	Salary       *int    `json:"salary"`
	DepartmentID *int    `json:"department_id"`
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

	employees, err := h.service.List(r.Context(), pagination)
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
	case errors.Is(err, models.ErrInvalidPagination):
		utils.WriteError(w, http.StatusBadRequest, models.ErrInvalidPagination.Error())
	default:
		utils.WriteError(w, http.StatusInternalServerError, "Internal server error")
	}
}
