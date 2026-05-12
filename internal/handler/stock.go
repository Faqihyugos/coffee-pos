package handler

import (
	"github.com/faqihyugos/coffee-pos/internal/entity"
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
