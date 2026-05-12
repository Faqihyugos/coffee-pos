package handler

import (
	"github.com/faqihyugos/coffee-pos/internal/entity"
	"github.com/faqihyugos/coffee-pos/internal/service"
	"github.com/faqihyugos/coffee-pos/pkg/response"
	"github.com/faqihyugos/coffee-pos/pkg/validator"
	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryService *service.CategoryService
	validator       *validator.Validator
}

func NewCategoryHandler(categoryService *service.CategoryService, v *validator.Validator) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
		validator:       v,
	}
}

func (h *CategoryHandler) FindAll(c *gin.Context) {
	categories, err := h.categoryService.FindAll(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Gagal mengambil data kategori")
		return
	}
	response.OK(c, "Berhasil", categories)
}

func (h *CategoryHandler) FindByID(c *gin.Context) {
	id := c.Param("id")
	category, err := h.categoryService.FindByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, "Berhasil", category)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var req entity.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format request tidak valid")
		return
	}

	if errs := h.validator.Validate(req); errs != nil {
		response.ValidationError(c, errs)
		return
	}

	category, err := h.categoryService.Create(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, "Kategori berhasil dibuat", category)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req entity.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format request tidak valid")
		return
	}

	if errs := h.validator.Validate(req); errs != nil {
		response.ValidationError(c, errs)
		return
	}

	category, err := h.categoryService.Update(c.Request.Context(), id, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, "Kategori berhasil diupdate", category)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.categoryService.Delete(c.Request.Context(), id)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "Kategori berhasil dihapus", nil)
}
