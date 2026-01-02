package handlers

import (
	"cflow/internal/models"
	"cflow/internal/services"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type PaymentHandler struct {
	svc      *services.PaymentService
	validate *validator.Validate
}

func NewPaymentHandler(svc *services.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		svc:      svc,
		validate: validator.New(),
	}
}

func (h *PaymentHandler) CreatePayment(c echo.Context) error {
	ctx := c.Request().Context()

	var paymentRequest models.CreatePaymentRequest
	if err := c.Bind(&paymentRequest); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.validate.Struct(&paymentRequest); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := h.svc.CreatePayment(ctx, &paymentRequest)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, paymentRequest)
}

func (h *PaymentHandler) GetPaymentByID(c echo.Context) error {
	ctx := c.Request().Context()

	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "id is required")
	}

	uuid, err := uuid.Parse(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	payment, err := h.svc.GetPayment(ctx, uuid)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, payment)
}
