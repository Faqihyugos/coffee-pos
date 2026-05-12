package handler

import (
	"github.com/faqihyugos/coffee-pos/internal/entity"
	"github.com/faqihyugos/coffee-pos/internal/service"
	"github.com/faqihyugos/coffee-pos/pkg/response"
	"github.com/faqihyugos/coffee-pos/pkg/validator"
	"github.com/gin-gonic/gin"
)

type TableHandler struct {
	tableService *service.TableService
	validator    *validator.Validator
}

func NewTableHandler(tableService *service.TableService, v *validator.Validator) *TableHandler {
	return &TableHandler{
		tableService: tableService,
		validator:    v,
	}
}

func (h *TableHandler) FindAll(c *gin.Context) {
	tables, err := h.tableService.FindAll(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Gagal mengambil data meja")
		return
	}
	response.OK(c, "Berhasil", tables)
}

func (h *TableHandler) Create(c *gin.Context) {
	var req entity.CreateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format request tidak valid")
		return
	}

	if errs := h.validator.Validate(req); errs != nil {
		response.ValidationError(c, errs)
		return
	}

	table, err := h.tableService.Create(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, "Meja berhasil dibuat", table)
}

func (h *TableHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req entity.UpdateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format request tidak valid")
		return
	}

	if errs := h.validator.Validate(req); errs != nil {
		response.ValidationError(c, errs)
		return
	}

	table, err := h.tableService.Update(c.Request.Context(), id, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, "Meja berhasil diupdate", table)
}

func (h *TableHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.tableService.Delete(c.Request.Context(), id)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "Meja berhasil dihapus", nil)
}
