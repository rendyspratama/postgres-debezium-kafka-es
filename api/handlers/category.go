package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rendyspratama/digital-discovery/api/models"
	"github.com/rendyspratama/digital-discovery/api/repositories"
	"github.com/rendyspratama/digital-discovery/api/utils"
)

type CategoryHandler struct {
	repo repositories.CategoryRepository
}

func NewCategoryHandler(repo repositories.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{repo: repo}
}

// V1 Handlers

func (h *CategoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value("requestID").(string)
	categories, err := h.repo.GetAllCategories()
	if err != nil {
		utils.WriteErrorWithRequestID(w, http.StatusInternalServerError,
			fmt.Sprintf("Failed to fetch categories: %v", err), requestID)
		return
	}
	utils.WriteSuccessWithRequestID(w, categories, requestID)
}

func (h *CategoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value("requestID").(string)
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		utils.WriteErrorWithRequestID(w, http.StatusBadRequest,
			"Category ID is required", requestID)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.WriteErrorWithRequestID(w, http.StatusBadRequest,
			"Invalid category ID format", requestID)
		return
	}

	category, err := h.repo.GetCategoryByID(id)
	if err != nil {
		utils.WriteErrorWithRequestID(w, http.StatusInternalServerError,
			"Failed to fetch category", requestID)
		return
	}
	if category == nil {
		utils.WriteErrorWithRequestID(w, http.StatusNotFound,
			"Category not found", requestID)
		return
	}
	utils.WriteSuccessWithRequestID(w, category, requestID)
}

func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value("requestID").(string)
	var category models.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		utils.WriteErrorWithRequestID(w, http.StatusBadRequest,
			fmt.Sprintln("Invalid request body", err), requestID)
		return
	}

	if err := category.Validate(); err != nil {
		utils.WriteErrorWithRequestID(w, http.StatusBadRequest,
			err.Error(), requestID)
		return
	}

	if err := h.repo.CreateCategory(&category); err != nil {
		utils.WriteErrorWithRequestID(w, http.StatusInternalServerError,
			"Failed to create category", requestID)
		return
	}

	utils.WriteSuccessWithRequestID(w, category, requestID)
}

func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		utils.WriteError(w, http.StatusBadRequest, "Category ID is required")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid category ID format")
		return
	}

	var category models.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := category.Validate(); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	category.ID = id
	if err := h.repo.UpdateCategory(&category); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to update category")
		return
	}

	utils.WriteSuccess(w, category)
}

func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		utils.WriteError(w, http.StatusBadRequest, "Category ID is required")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid category ID format")
		return
	}

	if err := h.repo.DeleteCategory(id); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to delete category")
		return
	}

	utils.WriteSuccess(w, map[string]string{"message": "Category deleted successfully"})
}

// V2 Handlers

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination struct {
		Total       int  `json:"total"`
		Page        int  `json:"page"`
		PerPage     int  `json:"per_page"`
		TotalPages  int  `json:"total_pages"`
		HasNextPage bool `json:"has_next_page"`
	} `json:"pagination"`
}

func (h *CategoryHandler) GetCategoriesV2(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}

	// Get categories with pagination
	categories, total, err := h.repo.GetCategoriesWithPagination(page, perPage)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch categories")
		return
	}

	// Calculate pagination metadata
	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	// Prepare response
	response := PaginatedResponse{
		Data: categories,
	}
	response.Pagination.Total = total
	response.Pagination.Page = page
	response.Pagination.PerPage = perPage
	response.Pagination.TotalPages = totalPages
	response.Pagination.HasNextPage = page < totalPages

	utils.WriteSuccess(w, response)
}
