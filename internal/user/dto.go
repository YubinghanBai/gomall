package user

import "time"

type LoginContext struct {
	Username  string
	Email     string
	Password  string
	UserAgent string
	ClientIP  string
}

type SessionInfo struct {
	SessionID string    `json:"session_id"`
	UserAgent string    `json:"user_agent"`
	ClientIP  string    `json:"client_ip"`
	IsBlocked bool      `json:"is_blocked"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	IsCurrent bool      `json:"is_current"` // 是否是当前设备
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Phone    string `json:"phone" binding:"omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginWithUsernameRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type SendEmailVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}
type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type UserResponse struct {
	ID              int64     `json:"id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone,omitempty"`
	Nickname        string    `json:"nickname,omitempty"`
	Avatar          string    `json:"avatar,omitempty"`
	Gender          string    `json:"gender"`
	IsEmailVerified bool      `json:"is_email_verified"`
	CreatedAt       time.Time `json:"created_at"`
}

type LoginResponse struct {
	SessionID             string        `json:"session_id"`
	AccessToken           string        `json:"access_token"`
	AccessTokenExpiresAt  time.Time     `json:"access_token_expires_at"`
	RefreshToken          string        `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time     `json:"refresh_token_expires_at"`
	User                  *UserResponse `json:"user"`
}

type UpdateProfileRequest struct {
	Nickname *string    `json:"nickname" binding:"omitempty,min=1,max=50"`
	Avatar   *string    `json:"avatar" binding:"omitempty,url"`
	Gender   *string    `json:"gender" binding:"omitempty,oneof=male female unknown"`
	Birthday *time.Time `json:"birthday" binding:"omitempty"`
}
