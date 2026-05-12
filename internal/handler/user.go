package handler

import (
	"github.com/faqihyugos/coffee-pos/internal/entity"
	"github.com/faqihyugos/coffee-pos/internal/service"
	"github.com/faqihyugos/coffee-pos/pkg/response"
	"github.com/faqihyugos/coffee-pos/pkg/validator"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
	validator   *validator.Validator
}

func NewUserHandler(userService *service.UserService, v *validator.Validator) *UserHandler {
	return &UserHandler{
		userService: userService,
		validator:   v,
	}
}

func (h *UserHandler) FindAll(c *gin.Context) {
	cashiers, err := h.userService.FindAllCashiers(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Gagal mengambil data cashier")
		return
	}
	response.OK(c, "Berhasil", cashiers)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req entity.CreateCashierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format request tidak valid")
		return
	}

	if errs := h.validator.Validate(req); errs != nil {
		response.ValidationError(c, errs)
		return
	}

	result, err := h.userService.CreateCashier(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, "Cashier berhasil dibuat", result)
}

func (h *UserHandler) Update(c *gin.Context) {
	var req entity.UpdateCashierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format request tidak valid")
		return
	}

	if errs := h.validator.Validate(req); errs != nil {
		response.ValidationError(c, errs)
		return
	}

	result, err := h.userService.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, "Data cashier berhasil diupdate", result)
}

func (h *UserHandler) ToggleStatus(c *gin.Context) {
	result, err := h.userService.ToggleStatus(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "Status cashier berhasil diubah", result)
}
