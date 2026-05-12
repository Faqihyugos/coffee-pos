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

type StockHandler struct {
	stockService *service.StockService
	validator    *validator.Validator
}

func NewStockHandler(stockService *service.StockService, v *validator.Validator) *StockHandler {
	return &StockHandler{
		stockService: stockService,
		validator:    v,
	}
}

func (h *StockHandler) GetStock(c *gin.Context) {
	productID := c.Param("id")
	product, err := h.stockService.GetStock(c.Request.Context(), productID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, "Berhasil", gin.H{
		"stock":      product.Stock,
		"product_id": product.ID,
		"name":       product.Name,
	})
}

func (h *StockHandler) Adjust(c *gin.Context) {
	productID := c.Param("id")
	userID := c.GetString("user_id")

	var req entity.StockAdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format request tidak valid")
		return
	}

	if errs := h.validator.Validate(req); errs != nil {
		response.ValidationError(c, errs)
		return
	}

	err := h.stockService.Adjust(c.Request.Context(), productID, userID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, "Stok berhasil diupdate", nil)
}

func (h *StockHandler) GetMovements(c *gin.Context) {
	productID := c.Param("id")

	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	typeFilter := c.Query("type")

	filter := repository.StockFilter{
		Page:  page,
		Limit: limit,
		Type:  typeFilter,
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	movements, total, err := h.stockService.GetMovements(c.Request.Context(), productID, filter)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	paginatedResponse := PaginatedResponse{
		Items: movements,
		Total: total,
		Page:  filter.Page,
		Limit: filter.Limit,
	}

	response.OK(c, "Berhasil", paginatedResponse)
}
