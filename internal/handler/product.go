package handler

import (
	"strconv"

	"github.com/faqihyugos/coffee-pos/internal/entity"
	"github.com/faqihyugos/coffee-pos/internal/repository"
	"github.com/faqihyugos/coffee-pos/internal/service"
	"github.com/faqihyugos/coffee-pos/pkg/response"
	"github.com/faqihyugos/coffee-pos/pkg/validator"
	"github.com/gin-gonic/gin"
)

type PaginatedResponse struct {
	Items interface{} `json:"items"`
	Total int         `json:"total"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
}

type ProductHandler struct {
	productService *service.ProductService
	validator      *validator.Validator
}

func NewProductHandler(productService *service.ProductService, v *validator.Validator) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		validator:      v,
	}
}

func (h *ProductHandler) FindAll(c *gin.Context) {
	search := c.Query("search")
	categoryID := c.Query("category_id")
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	var isActive *bool
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if parsed, err := strconv.ParseBool(isActiveStr); err == nil {
			isActive = &parsed
		}
	}

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	filter := repository.ProductFilter{
		Search:     search,
		CategoryID: categoryID,
		IsActive:   isActive,
		Page:       page,
		Limit:      limit,
	}

	products, total, err := h.productService.FindAll(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c, "Gagal mengambil data produk")
		return
	}

	paginatedResponse := PaginatedResponse{
		Items: products,
		Total: total,
		Page:  filter.Page,
		Limit: filter.Limit,
	}

	response.OK(c, "Berhasil", paginatedResponse)
}

func (h *ProductHandler) FindByID(c *gin.Context) {
	id := c.Param("id")
	product, err := h.productService.FindByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, "Berhasil", product)
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req entity.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format request tidak valid")
		return
	}

	if errs := h.validator.Validate(req); errs != nil {
		response.ValidationError(c, errs)
		return
	}

	product, err := h.productService.Create(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, "Produk berhasil dibuat", product)
}

func (h *ProductHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req entity.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format request tidak valid")
		return
	}

	if errs := h.validator.Validate(req); errs != nil {
		response.ValidationError(c, errs)
		return
	}

	product, err := h.productService.Update(c.Request.Context(), id, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, "Produk berhasil diupdate", product)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.productService.Delete(c.Request.Context(), id)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "Produk berhasil dihapus", nil)
}
