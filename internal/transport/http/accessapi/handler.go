package accessapi

import (
	"context"
	"net/http"
	"time"

	_ "github.com/aipass/aipass/docs/swagger"
	"github.com/aipass/aipass/internal/auth"
	"github.com/aipass/aipass/internal/config"
	"github.com/aipass/aipass/internal/domain"
	"github.com/aipass/aipass/internal/metrics"
	"github.com/aipass/aipass/internal/service"
	"github.com/aipass/aipass/internal/transport/http/dto"
	appmw "github.com/aipass/aipass/internal/transport/http/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	emw "github.com/labstack/echo/v4/middleware"
	"github.com/shopspring/decimal"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
)

type handler struct {
	cfg      config.Config
	services *service.Container
	metrics  *metrics.Registry
	validate *validator.Validate
	db       *sqlx.DB
}

func Register(e *echo.Echo, cfg config.Config, services *service.Container, registry *metrics.Registry, log *zap.Logger, database *sqlx.DB) {
	h := &handler{cfg: cfg, services: services, metrics: registry, validate: validator.New(), db: database}

	e.HideBanner = true
	e.HTTPErrorHandler = appmw.HTTPErrorHandler(log)
	e.Use(emw.Recover())
	e.Use(emw.RequestID())
	e.Use(registry.Middleware())

	e.GET("/health", h.health)
	e.GET("/ready", h.ready)
	e.GET("/metrics", registry.Handler())
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	api := e.Group("/api/v1")
	api.POST("/auth/login", h.login)

	protected := api.Group("", auth.Middleware(services.Auth.TokenManager()))
	protected.GET("/auth/me", h.me)

	admin := protected.Group("", auth.RequireRole(domain.RoleAdmin))
	admin.POST("/users", h.createUser)
	admin.GET("/users", h.listUsers)
	admin.GET("/users/:id", h.getUser)
	admin.PATCH("/users/:id", h.updateUser)
	admin.POST("/users/:id/photo", h.notImplemented)

	admin.POST("/plans", h.createPlan)
	admin.GET("/plans", h.listPlans)
	admin.GET("/plans/:id", h.getPlan)
	admin.PATCH("/plans/:id", h.updatePlan)
	admin.DELETE("/plans/:id", h.deletePlan)

	admin.POST("/users/:id/subscriptions", h.assignSubscription)
	admin.GET("/users/:id/subscriptions", h.listUserSubscriptions)
	admin.GET("/subscriptions/:id", h.getSubscription)
	admin.PATCH("/subscriptions/:id/status", h.updateSubscriptionStatus)

	admin.POST("/subscriptions/:id/qr-pass", h.generateQRPass)
	admin.GET("/users/:id/qr-pass", h.getUserQRPass)
	admin.POST("/qr-passes/:id/revoke", h.revokeQRPass)

	admin.POST("/subscriptions/:id/payments/receipt", h.createManualPayment)
	admin.GET("/payments", h.listPayments)
	admin.GET("/payments/:id", h.getPayment)
	admin.POST("/payments/:id/approve", h.approvePayment)
	admin.POST("/payments/:id/reject", h.rejectPayment)

	admin.GET("/reports/access-events.xlsx", h.accessEventsReport)
	admin.GET("/reports/payments.xlsx", h.paymentsReport)
}

func (h *handler) health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok", "service": h.cfg.App.Name})
}

func (h *handler) ready(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()
	if err := h.db.PingContext(ctx); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status":   "not_ready",
			"service":  h.cfg.App.Name,
			"postgres": "down",
		})
	}
	return c.JSON(http.StatusOK, map[string]string{
		"status":   "ready",
		"service":  h.cfg.App.Name,
		"postgres": "up",
	})
}

// login godoc
// @Summary Login admin or member
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *handler) login(c echo.Context) error {
	var req dto.LoginRequest
	if err := bindValidate(c, h.validate, &req); err != nil {
		return err
	}
	token, user, err := h.services.Auth.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.LoginResponse{AccessToken: token, User: user})
}

func (h *handler) me(c echo.Context) error {
	claims, _ := auth.ClaimsFromContext(c)
	user, err := h.services.Users.Get(c.Request().Context(), claims.UserID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, user)
}

func (h *handler) createUser(c echo.Context) error {
	var req dto.CreateUserRequest
	if err := bindValidate(c, h.validate, &req); err != nil {
		return err
	}
	user, err := h.services.Users.Create(c.Request().Context(), service.CreateUserInput{
		Email: req.Email, Phone: req.Phone, FullName: req.FullName, Role: req.Role, Password: req.Password,
	})
	if err != nil {
		return err
	}
	h.metrics.Inc("users_created_total")
	return c.JSON(http.StatusCreated, user)
}

func (h *handler) listUsers(c echo.Context) error {
	users, err := h.services.Users.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, users)
}

func (h *handler) getUser(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	user, err := h.services.Users.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, user)
}

func (h *handler) updateUser(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var req dto.UpdateUserRequest
	if err := bindValidate(c, h.validate, &req); err != nil {
		return err
	}
	user, err := h.services.Users.Update(c.Request().Context(), id, req.Phone, req.FullName, req.IsActive)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, user)
}

func (h *handler) createPlan(c echo.Context) error {
	var req dto.CreatePlanRequest
	if err := bindValidate(c, h.validate, &req); err != nil {
		return err
	}
	price, err := decimal.NewFromString(req.Price)
	if err != nil {
		return service.ErrInvalidInput
	}
	plan, err := h.services.Plans.Create(c.Request().Context(), req.Name, req.Description, req.DurationDays, price, req.Currency)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, plan)
}

func (h *handler) listPlans(c echo.Context) error {
	plans, err := h.services.Plans.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, plans)
}

func (h *handler) getPlan(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	plan, err := h.services.Plans.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, plan)
}

func (h *handler) updatePlan(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var req dto.UpdatePlanRequest
	if err := bindValidate(c, h.validate, &req); err != nil {
		return err
	}
	var price *decimal.Decimal
	if req.Price != nil {
		parsed, err := decimal.NewFromString(*req.Price)
		if err != nil {
			return service.ErrInvalidInput
		}
		price = &parsed
	}
	plan, err := h.services.Plans.Update(c.Request().Context(), id, req.Name, req.Description, req.DurationDays, price, req.Currency, req.IsActive)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, plan)
}

func (h *handler) deletePlan(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	plan, err := h.services.Plans.Deactivate(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, plan)
}

func (h *handler) assignSubscription(c echo.Context) error {
	userID, err := parseID(c)
	if err != nil {
		return err
	}
	var req dto.AssignSubscriptionRequest
	if err := bindValidate(c, h.validate, &req); err != nil {
		return err
	}
	sub, err := h.services.Subscriptions.Assign(c.Request().Context(), userID, req.PlanID, req.StartsAt, req.Status)
	if err != nil {
		return err
	}
	h.metrics.Inc("subscriptions_assigned_total")
	return c.JSON(http.StatusCreated, sub)
}

func (h *handler) listUserSubscriptions(c echo.Context) error {
	userID, err := parseID(c)
	if err != nil {
		return err
	}
	subs, err := h.services.Subscriptions.ListByUser(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, subs)
}

func (h *handler) getSubscription(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	sub, err := h.services.Subscriptions.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, sub)
}

func (h *handler) updateSubscriptionStatus(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var req dto.UpdateSubscriptionStatusRequest
	if err := bindValidate(c, h.validate, &req); err != nil {
		return err
	}
	sub, err := h.services.Subscriptions.UpdateStatus(c.Request().Context(), id, req.Status)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, sub)
}

func (h *handler) generateQRPass(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	result, err := h.services.QR.Generate(c.Request().Context(), id)
	if err != nil {
		return err
	}
	h.metrics.Inc("qr_passes_generated_total")
	return c.JSON(http.StatusCreated, dto.QRPassResponse{Pass: result.Pass, Token: result.Token})
}

func (h *handler) getUserQRPass(c echo.Context) error {
	userID, err := parseID(c)
	if err != nil {
		return err
	}
	pass, err := h.services.QR.GetLatestByUser(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.QRPassResponse{Pass: pass})
}

func (h *handler) revokeQRPass(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	if err := h.services.QR.Revoke(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *handler) createManualPayment(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	amount, err := decimal.NewFromString(c.FormValue("amount"))
	if err != nil {
		return service.ErrInvalidInput
	}
	payment, err := h.services.Payments.CreateManual(c.Request().Context(), id, amount, domain.PaymentKaspiManual, nil)
	if err != nil {
		return err
	}
	h.metrics.Inc("payments_uploaded_total")
	return c.JSON(http.StatusCreated, payment)
}

func (h *handler) listPayments(c echo.Context) error {
	payments, err := h.services.Payments.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, payments)
}

func (h *handler) getPayment(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	payment, err := h.services.Payments.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, payment)
}

func (h *handler) approvePayment(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	claims, _ := auth.ClaimsFromContext(c)
	payment, err := h.services.Payments.Approve(c.Request().Context(), id, claims.UserID)
	if err != nil {
		return err
	}
	h.metrics.Inc("payments_approved_total")
	return c.JSON(http.StatusOK, payment)
}

func (h *handler) rejectPayment(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	payment, err := h.services.Payments.Reject(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, payment)
}

func (h *handler) accessEventsReport(c echo.Context) error {
	data, err := h.services.Reports.AccessEventsXLSX(c.Request().Context())
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

func (h *handler) paymentsReport(c echo.Context) error {
	data, err := h.services.Reports.PaymentsXLSX(c.Request().Context())
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

func (h *handler) notImplemented(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "will_be_implemented_in_minio_version"})
}

func bindValidate(c echo.Context, v *validator.Validate, dest any) error {
	if err := c.Bind(dest); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid_json")
	}
	if err := v.Struct(dest); err != nil {
		return service.ErrInvalidInput
	}
	return nil
}

func parseID(c echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, service.ErrInvalidInput
	}
	return id, nil
}
