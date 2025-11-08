package user

import (
	"github.com/gin-gonic/gin"
	"gomall/internal/common/middleware"
	"gomall/utils/response"
	"gomall/utils/token"
	"net/http"
)

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	users := r.Group("/users")
	{
		// 公开路由
		users.POST("/register", h.Register)
		users.POST("/login", h.Login)
		users.POST("/login/username", h.LoginWithUsername)
		users.POST("/refresh", h.RefreshToken)

		// 邮箱验证
		users.POST("/email/send-verification", h.SendEmailVerification)
		users.POST("/email/verify", h.VerifyEmail)

		// 忘记密码
		users.POST("/password/forgot", h.ForgotPassword)
		users.POST("/password/reset", h.ResetPassword)
	}

	// 需要认证的路由
	auth := r.Group("/users")
	auth.Use(middleware.AuthMiddleware(h.tokenMaker))
	{
		auth.GET("/profile", h.GetProfile)
		auth.PUT("/profile", h.UpdateProfile)
		auth.POST("/password/change", h.ChangePassword)
		auth.POST("/logout", h.Logout)

		// Session 管理
		auth.GET("/sessions", h.GetSessions)
		auth.DELETE("/sessions/:session_id", h.RevokeSession)
		auth.DELETE("/sessions/others", h.RevokeAllOtherSessions)
	}
}

type Handler struct {
	service    Service
	tokenMaker token.Maker
}

// NewHandler 创建 Handler 实例
func NewHandler(service Service, tokenMaker token.Maker) *Handler {
	return &Handler{
		service:    service,
		tokenMaker: tokenMaker,
	}
}

// Register 用户注册
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, user)
}

// Login 用户登录
func (h *Handler) Login(c *gin.Context) {
	var req LoginContext
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	loginCtx := LoginContext{
		Email:     req.Email,
		Password:  req.Password,
		UserAgent: c.Request.UserAgent(),
		ClientIP:  c.ClientIP(),
	}

	result, err := h.service.Login(c.Request.Context(), loginCtx)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	response.Success(c, result)
}

// GetProfile 获取用户信息
func (h *Handler) GetProfile(c *gin.Context) {
	payload := middleware.GetPayload(c)
	if payload == nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.service.GetProfile(c.Request.Context(), payload.UserID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, user)
}

// UpdateProfile 更新用户信息
func (h *Handler) UpdateProfile(c *gin.Context) {
	payload := middleware.GetPayload(c)
	if payload == nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.service.UpdateProfile(c.Request.Context(), payload.UserID, req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, user)
}

func (h *Handler) Logout(c *gin.Context) {
	//get payload from auth middleware
	payload := middleware.GetPayload(c)
	if payload == nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	sessionID := payload.ID.String()
	if err := h.service.Logout(c.Request.Context(), sessionID); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "logged out successfully"})

}

func (h *Handler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	response.Success(c, result)
}

// ChangePassword 修改密码
func (h *Handler) ChangePassword(c *gin.Context) {
	payload := middleware.GetPayload(c)
	if payload == nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err := h.service.ChangePassword(c.Request.Context(), payload.UserID, req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "password changed successfully"})
}

// SendEmailVerification 发送邮箱验证码
func (h *Handler) SendEmailVerification(c *gin.Context) {
	var req SendEmailVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err := h.service.SendEmailVerification(c.Request.Context(), req.Email)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "verification code sent to email"})
}

// VerifyEmail 验证邮箱
func (h *Handler) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err := h.service.VerifyEmail(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "email verified successfully"})
}

// ForgotPassword 忘记密码 - 发送重置验证码
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err := h.service.ForgotPassword(c.Request.Context(), req.Email)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "if the email exists, a reset code has been sent"})
}

// ResetPassword 重置密码
func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err := h.service.ResetPassword(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "password reset successfully"})
}

// LoginWithUsername 用户名登录
func (h *Handler) LoginWithUsername(c *gin.Context) {
	var req LoginWithUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	loginCtx := LoginContext{
		Email:     req.Username, // 复用 Email 字段传递 username
		Password:  req.Password,
		UserAgent: c.Request.UserAgent(),
		ClientIP:  c.ClientIP(),
	}

	result, err := h.service.LoginWithUsername(c.Request.Context(), loginCtx)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	response.Success(c, result)
}

// GetSessions 获取用户所有 session
func (h *Handler) GetSessions(c *gin.Context) {
	payload := middleware.GetPayload(c)
	if payload == nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	sessions, err := h.service.GetUserSessions(c.Request.Context(), payload.UserID, payload.ID.String())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, sessions)
}

// RevokeSession 撤销指定 session
func (h *Handler) RevokeSession(c *gin.Context) {
	payload := middleware.GetPayload(c)
	if payload == nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	sessionID := c.Param("session_id")
	err := h.service.RevokeSession(c.Request.Context(), payload.UserID, sessionID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "session revoked successfully"})
}

// RevokeAllOtherSessions 撤销所有其他设备
func (h *Handler) RevokeAllOtherSessions(c *gin.Context) {
	payload := middleware.GetPayload(c)
	if payload == nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	err := h.service.RevokeAllOtherSessions(c.Request.Context(), payload.UserID, payload.ID.String())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "all other sessions revoked successfully"})
}
