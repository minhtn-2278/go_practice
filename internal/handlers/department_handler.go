package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"employee-management/internal/models"
	"employee-management/internal/services"
	"employee-management/internal/utils"
)

type DepartmentHandler struct {
	service services.DepartmentUseCase
}

func NewDepartmentHandler(service services.DepartmentUseCase) (*DepartmentHandler, error) {
	if service == nil {
		return nil, errors.New("department service is required")
	}

	return &DepartmentHandler{service: service}, nil
}

func (h *DepartmentHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /departments", h.create)
	mux.HandleFunc("GET /departments", h.list)
	mux.HandleFunc("GET /departments/{id}/employees", h.listEmployee)
}

func (h *DepartmentHandler) create(w http.ResponseWriter, r *http.Request) {
	var department models.Department
	if err := utils.DecodeJSON(r, &department); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	createdDepartment, err := h.service.Create(r.Context(), department)
	if err != nil {
		writeDepartmentServiceError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, createdDepartment)
}

func (h *DepartmentHandler) list(w http.ResponseWriter, r *http.Request) {
	pagination, err := utils.ParsePagination(r.URL.Query())
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, models.ErrInvalidPagination.Error())
		return
	}

	departments, err := h.service.List(r.Context(), pagination)
	if err != nil {
		writeDepartmentServiceError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, departments)
}

func (h *DepartmentHandler) listEmployee(w http.ResponseWriter, r *http.Request) {
	departmentID, err := departmentID(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid department ID")
		return
	}

	pagination, err := utils.ParsePagination(r.URL.Query())
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, models.ErrInvalidPagination.Error())
		return
	}

	employees, err := h.service.ListEmployee(r.Context(), departmentID, pagination)
	if err != nil {
		writeDepartmentServiceError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, employees)
}

func departmentID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func writeDepartmentServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidDepartment):
		utils.WriteError(w, http.StatusBadRequest, services.ErrInvalidDepartment.Error())
	case errors.Is(err, models.ErrInvalidPagination):
		utils.WriteError(w, http.StatusBadRequest, models.ErrInvalidPagination.Error())
	default:
		utils.WriteError(w, http.StatusInternalServerError, "Internal server error")
	}
}
