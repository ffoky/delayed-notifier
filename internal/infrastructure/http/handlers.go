package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"DelayedNotifier/internal/domain"
	"DelayedNotifier/internal/infrastructure/http/generated"
)

type Handler struct {
	svc      domain.NotificationService
	usersSvc domain.UserService
}

func NewHandler(svc domain.NotificationService, usersSvc domain.UserService) *Handler {
	return &Handler{svc: svc, usersSvc: usersSvc}
}

func (h *Handler) ListNotifications(c *gin.Context) {
	ns, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, generated.ErrorResponse{Error: "internal server error"})
		return
	}

	resp := make([]generated.NotificationResponse, len(ns))
	for i, n := range ns {
		resp[i] = toNotificationResponse(n)
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateNotification(c *gin.Context) {
	var req generated.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, generated.ErrorResponse{Error: err.Error()})
		return
	}

	n, err := h.svc.Create(c.Request.Context(), req.UserId, domain.Channel(req.Channel), req.Text, req.PlannedAt)
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toNotificationResponse(n))
}

func (h *Handler) GetNotification(c *gin.Context, id generated.NotificationID) {
	n, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, toNotificationResponse(n))
}

func (h *Handler) CancelNotification(c *gin.Context, id generated.NotificationID) {
	n, err := h.svc.Cancel(c.Request.Context(), id)
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, toNotificationResponse(n))
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req generated.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, generated.ErrorResponse{Error: err.Error()})
		return
	}

	var email *string
	if req.Email != nil {
		s := string(*req.Email)
		email = &s
	}

	u, err := h.usersSvc.Create(c.Request.Context(), req.TelegramId, email)
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toUserResponse(u))
}

func (h *Handler) AuthUser(c *gin.Context) {
	var req generated.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, generated.ErrorResponse{Error: err.Error()})
		return
	}

	u, err := h.usersSvc.GetOrCreate(c.Request.Context(), req.TelegramId, req.Email)
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, toUserResponse(u))
}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound), errors.Is(err, domain.ErrUserNotFound):
		c.JSON(http.StatusNotFound, generated.ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrAlreadySent), errors.Is(err, domain.ErrAlreadyCancelled):
		c.JSON(http.StatusConflict, generated.ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrInvalidChannel),
		errors.Is(err, domain.ErrEmptyText),
		errors.Is(err, domain.ErrPastPlannedAt),
		errors.Is(err, domain.ErrNoTelegramID),
		errors.Is(err, domain.ErrNoEmail),
		errors.Is(err, domain.ErrNoContact):
		c.JSON(http.StatusBadRequest, generated.ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, generated.ErrorResponse{Error: "internal server error"})
	}
}

func toNotificationResponse(n *domain.Notification) generated.NotificationResponse {
	retries := n.Retries
	return generated.NotificationResponse{
		Id:        n.ID,
		UserId:    n.UserID,
		Channel:   generated.Channel(n.Channel),
		Text:      n.Text,
		Status:    generated.Status(n.Status),
		PlannedAt: n.PlannedAt,
		CreatedAt: n.CreatedAt,
		SentAt:    n.SentAt,
		Retries:   &retries,
	}
}

func toUserResponse(u *domain.User) generated.UserResponse {
	return generated.UserResponse{
		Id:         u.ID,
		TelegramId: u.TelegramID,
		Email:      u.Email,
	}
}
