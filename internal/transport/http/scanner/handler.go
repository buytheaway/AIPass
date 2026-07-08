package scanner

import (
	"net/http"

	"github.com/aipass/aipass/internal/config"
	"github.com/aipass/aipass/internal/domain"
	"github.com/aipass/aipass/internal/metrics"
	"github.com/aipass/aipass/internal/service"
	"github.com/aipass/aipass/internal/transport/http/dto"
	appmw "github.com/aipass/aipass/internal/transport/http/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	emw "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type handler struct {
	cfg      config.Config
	services *service.Container
	metrics  *metrics.Registry
	validate *validator.Validate
}

func Register(e *echo.Echo, cfg config.Config, services *service.Container, registry *metrics.Registry, log *zap.Logger) {
	h := &handler{cfg: cfg, services: services, metrics: registry, validate: validator.New()}

	e.HideBanner = true
	e.HTTPErrorHandler = appmw.HTTPErrorHandler(log)
	e.Use(emw.Recover())
	e.Use(emw.RequestID())
	e.Use(registry.Middleware())

	e.GET("/health", h.health)
	e.GET("/ready", h.health)
	e.GET("/metrics", registry.Handler())
	e.Static("/scanner/assets", "web/scanner")
	e.File("/scanner", "web/scanner/index.html")

	api := e.Group("/api/v1")
	api.POST("/scans/validate", h.validateScan)
}

func (h *handler) health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok", "service": h.cfg.App.Name})
}

func (h *handler) validateScan(c echo.Context) error {
	var req dto.ValidateScanRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid_json")
	}
	if err := h.validate.Struct(req); err != nil {
		return service.ErrInvalidInput
	}
	result, err := h.services.Access.ValidateScan(c.Request().Context(), req.QRToken, req.ScannerID)
	if err != nil {
		return err
	}
	h.metrics.Inc("qr_scans_total")
	if result.Decision == domain.DecisionAllowed {
		h.metrics.Inc("qr_scans_allowed_total")
		if result.EventType == domain.AccessCheckIn {
			h.metrics.Inc("access_checkins_total")
		}
		if result.EventType == domain.AccessCheckOut {
			h.metrics.Inc("access_checkouts_total")
		}
	} else {
		h.metrics.Inc("qr_scans_denied_total")
	}

	var user *dto.ScanUser
	if result.User != nil {
		user = &dto.ScanUser{ID: result.User.ID, FullName: result.User.FullName}
	}
	return c.JSON(http.StatusOK, dto.ValidateScanResponse{
		Decision: result.Decision, EventType: result.EventType, User: user, Reason: result.Reason,
	})
}
