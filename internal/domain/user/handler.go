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

// Register godoc
// @Summary      User Register
// @Description  Create new user account
// @Tags         User Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      RegisterRequest  true  "Registration Infomation"
// @Success      200      {object}  response.Response{data=UserResponse}
// @Failure      400      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /users/register [post]
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

// Login godoc
// @Summary      User Login
// @Description  Login with email and password
// @Tags         User Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      LoginRequest  true  "Login Info"
// @Success      200      {object}  response.Response{data=LoginResponse}
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Router       /users/login [post]
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

// GetProfile godoc
// @Summary      Get User Profile
// @Description  Get current logged-in user profile information
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  response.Response{data=UserResponse}
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /users/profile [get]
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

// UpdateProfile godoc
// @Summary      Update User Profile
// @Description  Update current logged-in user profile information
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request  body      UpdateProfileRequest  true  "Profile information"
// @Success      200      {object}  response.Response{data=UserResponse}
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Router       /users/profile [put]
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

// Logout godoc
// @Summary      User Logout
// @Description  Logout current user session
// @Tags         User Authentication
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /users/logout [post]
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

// RefreshToken godoc
// @Summary      Refresh Access Token
// @Description  Refresh access token using refresh token
// @Tags         User Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      object{refresh_token=string}  true  "Refresh token"
// @Success      200      {object}  response.Response{data=LoginResponse}
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Router       /users/refresh [post]
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

// ChangePassword godoc
// @Summary      Change Password
// @Description  Change current user password
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request  body      ChangePasswordRequest  true  "Password information"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Router       /users/password/change [post]
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

// SendEmailVerification godoc
// @Summary      Send Email Verification Code
// @Description  Send verification code to user email for email verification
// @Tags         User Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      SendEmailVerificationRequest  true  "Email information"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      404      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /users/email/send-verification [post]
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

// VerifyEmail godoc
// @Summary      Verify Email
// @Description  Verify user email using verification code
// @Tags         User Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      VerifyEmailRequest  true  "Verification information"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /users/email/verify [post]
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

// ForgotPassword godoc
// @Summary      Forgot Password
// @Description  Send password reset code to user email
// @Tags         User Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      ForgotPasswordRequest  true  "Email information"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /users/password/forgot [post]
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

// ResetPassword godoc
// @Summary      Reset Password
// @Description  Reset user password using verification code
// @Tags         User Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      ResetPasswordRequest  true  "Reset information"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /users/password/reset [post]
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

// LoginWithUsername godoc
// @Summary      Login with Username
// @Description  Login with username and password
// @Tags         User Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      LoginWithUsernameRequest  true  "Login information"
// @Success      200      {object}  response.Response{data=LoginResponse}
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Router       /users/login/username [post]
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

// GetSessions godoc
// @Summary      Get User Sessions
// @Description  Get all active sessions for current user
// @Tags         Session Management
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  response.Response{data=[]SessionInfo}
// @Failure      401  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /users/sessions [get]
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

// RevokeSession godoc
// @Summary      Revoke Session
// @Description  Revoke a specific user session
// @Tags         Session Management
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        session_id  path      string  true  "Session ID"
// @Success      200         {object}  response.Response
// @Failure      400         {object}  response.Response
// @Failure      401         {object}  response.Response
// @Failure      404         {object}  response.Response
// @Router       /users/sessions/{session_id} [delete]
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

// RevokeAllOtherSessions godoc
// @Summary      Revoke All Other Sessions
// @Description  Revoke all user sessions except current one
// @Tags         Session Management
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /users/sessions/others [delete]
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
