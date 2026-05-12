package handler

import (
	"strconv"

	"github.com/faqihyugos/coffee-pos/internal/entity"
	"github.com/faqihyugos/coffee-pos/internal/service"
	"github.com/faqihyugos/coffee-pos/pkg/response"
	"github.com/faqihyugos/coffee-pos/pkg/validator"
	"github.com/gin-gonic/gin"
)

type PromoHandler struct {
	promoService *service.PromoService
	validator    *validator.Validator
}

func NewPromoHandler(promoService *service.PromoService, v *validator.Validator) *PromoHandler {
	return &PromoHandler{
		promoService: promoService,
		validator:    v,
	}
}

func (h *PromoHandler) FindAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	promos, total, err := h.promoService.FindAll(c.Request.Context(), page, limit)
	if err != nil {
		response.InternalError(c, "Gagal mengambil data promo")
		return
	}

	paginatedResponse := PaginatedResponse{
		Items: promos,
		Total: total,
		Page:  page,
		Limit: limit,
	}

	response.OK(c, "Berhasil", paginatedResponse)
}

func (h *PromoHandler) Create(c *gin.Context) {
	var req entity.CreatePromoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format request tidak valid")
		return
	}

	if errs := h.validator.Validate(req); errs != nil {
		response.ValidationError(c, errs)
		return
	}

	promo, err := h.promoService.Create(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, "Promo berhasil dibuat", promo)
}

func (h *PromoHandler) Update(c *gin.Context) {
	var req entity.UpdatePromoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format request tidak valid")
		return
	}

	if errs := h.validator.Validate(req); errs != nil {
		response.ValidationError(c, errs)
		return
	}

	id := c.Param("id")
	promo, err := h.promoService.Update(c.Request.Context(), id, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, "Promo berhasil diupdate", promo)
}

func (h *PromoHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.promoService.Delete(c.Request.Context(), id)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, "Promo berhasil dihapus", nil)
}
