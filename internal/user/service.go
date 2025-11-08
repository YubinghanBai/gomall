package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gomall/config"
	"gomall/utils/dbtypes"
	"gomall/utils/mail"
	"gomall/utils/password"
	"gomall/utils/random"
	"gomall/utils/token"
	"time"

	sqlc "gomall/db/sqlc"
)

// Service 用户业务逻辑接口
type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*UserResponse, error)
	Login(ctx context.Context, loginCtx LoginContext) (*LoginResponse, error)
	LoginWithUsername(ctx context.Context, loginCtx LoginContext) (*LoginResponse, error)
	GetProfile(ctx context.Context, userID int64) (*UserResponse, error)
	UpdateProfile(ctx context.Context, userID int64, req UpdateProfileRequest) (*UserResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error)
	Logout(ctx context.Context, sessionID string) error

	ChangePassword(ctx context.Context, userID int64, req ChangePasswordRequest) error
	SendEmailVerification(ctx context.Context, email string) error
	VerifyEmail(ctx context.Context, req VerifyEmailRequest) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, req ResetPasswordRequest) error
	GetUserSessions(ctx context.Context, userID int64, currentSessionID string) ([]SessionInfo, error)
	RevokeSession(ctx context.Context, userID int64, sessionID string) error
	RevokeAllOtherSessions(ctx context.Context, userID int64, currentSessionID string) error
}

type service struct {
	repo        Repository
	tokenMaker  token.Maker
	config      *config.Config
	emailSender mail.Sender
}

// NewService 创建 Service 实例
func NewService(config *config.Config, repo Repository, maker token.Maker, emailSender mail.Sender) Service {
	return &service{
		config:      config,
		repo:        repo,
		tokenMaker:  maker,
		emailSender: emailSender,
	}
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*UserResponse, error) {
	// 1. 检查邮箱是否已存在
	_, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, errors.New("email already exists")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}

	// 2. 检查用户名是否已存在
	_, err = s.repo.GetUserByUsername(ctx, req.Username)
	if err == nil {
		return nil, errors.New("username already exists")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}

	// 3. 加密密码
	hashedPassword, err := password.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 4. 创建用户
	var phone *string
	if req.Phone != "" {
		phone = &req.Phone
	}

	user, err := s.repo.CreateUser(ctx, sqlc.CreateUserParams{
		Username: req.Username,
		Email:    req.Email,
		Phone:    phone,
		Password: hashedPassword,
		Nickname: nil,
		Avatar:   nil,
		Gender:   "unknown",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return toUserResponse(user), nil
}

func (s *service) Login(ctx context.Context, loginCtx LoginContext) (*LoginResponse, error) {

	// 1. Search User
	user, err := s.repo.GetUserByEmail(ctx, loginCtx.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invalid email or password")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 2. Verify password
	if err := password.VerifyPassword(user.Password, loginCtx.Password); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// 3. Generate Access Token
	accessToken, accessPayload, err := s.tokenMaker.CreateToken(user.ID, user.Username, "user", s.config.AccessTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(user.ID, user.Username, "user", s.config.RefreshTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}
	//create Session
	session, err := s.repo.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:           refreshPayload.ID,
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    loginCtx.UserAgent,
		ClientIp:     loginCtx.ClientIP,
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	//update last login time
	_ = s.repo.UpdateUserLastLogin(ctx, sqlc.UpdateUserLastLoginParams{
		LastLoginAt: dbtypes.NullTime{Time: time.Now(), Valid: true},
		LastLoginIp: &loginCtx.ClientIP,
		ID:          user.ID,
	})

	return &LoginResponse{
		SessionID:             session.ID.String(),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  toUserResponse(user),
	}, nil
}

func (s *service) LoginWithUsername(ctx context.Context, loginCtx LoginContext) (*LoginResponse, error) {
	user, err := s.repo.GetUserByUsername(ctx, loginCtx.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invalid username or password")
		}
	}
	// 2. Verify password
	if err := password.VerifyPassword(user.Password, loginCtx.Password); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// 3. Generate Access Token
	accessToken, accessPayload, err := s.tokenMaker.CreateToken(user.ID, user.Username, "user", s.config.AccessTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(user.ID, user.Username, "user", s.config.RefreshTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}
	//create Session
	session, err := s.repo.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:           refreshPayload.ID,
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    loginCtx.UserAgent,
		ClientIp:     loginCtx.ClientIP,
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	//update last login time
	_ = s.repo.UpdateUserLastLogin(ctx, sqlc.UpdateUserLastLoginParams{
		LastLoginAt: dbtypes.NullTime{Time: time.Now(), Valid: true},
		LastLoginIp: &loginCtx.ClientIP,
		ID:          user.ID,
	})

	return &LoginResponse{
		SessionID:             session.ID.String(),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  toUserResponse(user),
	}, nil
}

func (s *service) GetProfile(ctx context.Context, userID int64) (*UserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return toUserResponse(user), nil
}

func (s *service) UpdateProfile(ctx context.Context, userID int64, req UpdateProfileRequest) (*UserResponse, error) {
	// 检查用户是否存在
	_, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 更新用户信息
	err = s.repo.UpdateUser(ctx, sqlc.UpdateUserParams{
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		Gender:   req.Gender,
		Birthday: dbtypes.NewNullTime(req.Birthday),
		ID:       userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// 返回更新后的用户信息
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated user: %w", err)
	}

	return toUserResponse(user), nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	//Verify refreshToken
	refreshPayload, err := s.tokenMaker.VerifyToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	//Check Session
	session, err := s.repo.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	if session.IsBlocked {
		return nil, errors.New("session blocked")
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, errors.New("session has expired")
	}
	if session.RefreshToken != refreshToken {
		return nil, errors.New("invalid refresh token")
	}

	user, err := s.repo.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	accessToken, accessPayload, err := s.tokenMaker.CreateToken(user.ID, user.Username, "user", s.config.AccessTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}
	return &LoginResponse{
		SessionID:             session.ID.String(),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  toUserResponse(user),
	}, nil
}

func (s *service) Logout(ctx context.Context, sessionID string) error {
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return errors.New("invalid session id")
	}

	return s.repo.DeleteSession(ctx, id)
}

func (s *service) ChangePassword(ctx context.Context, id int64, req ChangePasswordRequest) error {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := password.VerifyPassword(req.OldPassword, user.Password); err != nil {
		return errors.New("old password is incorrect")
	}

	//check old password == new password
	if req.OldPassword == req.NewPassword {
		return errors.New("new password must be different from old password")
	}

	hashedPassword, err := password.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	err = s.repo.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		Password: hashedPassword,
		ID:       id,
	})

	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

func (s *service) SendEmailVerification(ctx context.Context, email string) error {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.IsEmailVerified {
		return errors.New("email is already verified")
	}

	code, err := random.GenerateCode(6)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	//save code (15 min)
	_, err = s.repo.CreateVerificationCode(ctx, sqlc.CreateVerificationCodeParams{
		UserID:    user.ID,
		Email:     email,
		Code:      code,
		Type:      "email_verification",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		return fmt.Errorf("failed to create verification code: %w", err)
	}

	//send email
	subject := "GoMall - Verification Code"
	body := mail.EmailVerificationTemplate(user.Username, code)
	err = s.emailSender.SendEmail(subject, body, []string{email}, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}
	return nil
}

func (s *service) VerifyEmail(ctx context.Context, req VerifyEmailRequest) error {
	// 1. 查找验证码
	verificationCode, err := s.repo.GetVerificationCode(ctx, sqlc.GetVerificationCodeParams{
		Email: req.Email,
		Code:  req.Code,
		Type:  "email_verification",
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("invalid or expired verification code")
		}
		return fmt.Errorf("failed to get verification code: %w", err)
	}

	// 2. 检查是否过期
	if time.Now().After(verificationCode.ExpiresAt) {
		return errors.New("verification code has expired")
	}

	return s.repo.ExecTx(ctx, func(q sqlc.Querier) error {
		// 3.1 标记验证码为已使用
		if err := q.MarkCodeAsUsed(ctx, verificationCode.ID); err != nil {
			return fmt.Errorf("failed to mark code as used: %w", err)
		}

		// 3.2 验证用户邮箱
		if err := q.VerifyUserEmail(ctx, verificationCode.UserID); err != nil {
			return fmt.Errorf("failed to verify user email: %w", err)
		}

		return nil
	})
}

// ForgotPassword 忘记密码 - 发送重置验证码
func (s *service) ForgotPassword(ctx context.Context, email string) error {
	// 1. 查找用户
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// 为了安全，即使邮箱不存在也返回成功
			return nil
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 2. 生成6位验证码
	code, err := random.GenerateCode(6)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// 3. 保存验证码（15分钟有效期）
	_, err = s.repo.CreateVerificationCode(ctx, sqlc.CreateVerificationCodeParams{
		UserID:    user.ID,
		Email:     email,
		Code:      code,
		Type:      "password_reset",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		return fmt.Errorf("failed to create verification code: %w", err)
	}

	// 4. 发送邮件
	subject := "GoMall - 密码重置"
	body := mail.PasswordResetTemplate(user.Username, code)
	err = s.emailSender.SendEmail(subject, body, []string{email}, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// ResetPassword 重置密码
func (s *service) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	// 1. 查找并验证验证码
	verificationCode, err := s.repo.GetVerificationCode(ctx, sqlc.GetVerificationCodeParams{
		Email: req.Email,
		Code:  req.Code,
		Type:  "password_reset",
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("invalid or expired verification code")
		}
		return fmt.Errorf("failed to get verification code: %w", err)
	}

	// 2. 检查是否过期
	if time.Now().After(verificationCode.ExpiresAt) {
		return errors.New("verification code has expired")
	}

	// 3. 加密新密码
	hashedPassword, err := password.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return s.repo.ExecTx(ctx, func(q sqlc.Querier) error {
		// 4.1 更新密码
		if err := q.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
			Password: hashedPassword,
			ID:       verificationCode.UserID,
		}); err != nil {
			return fmt.Errorf("failed to update password: %w", err)
		}

		// 4.2 标记验证码为已使用
		if err := q.MarkCodeAsUsed(ctx, verificationCode.ID); err != nil {
			return fmt.Errorf("failed to mark code as used: %w", err)
		}

		// 4.3 删除该用户的所有会话（强制重新登录）
		if err := q.DeleteUserSessions(ctx, verificationCode.UserID); err != nil {
			return fmt.Errorf("failed to delete user sessions: %w", err)
		}

		return nil
	})
}

// GetUserSessions 获取用户的所有 session
func (s *service) GetUserSessions(ctx context.Context, userID int64, currentSessionID string) ([]SessionInfo, error) {
	sessions, err := s.repo.GetUserSessions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	sessionInfos := make([]SessionInfo, 0, len(sessions))
	for _, session := range sessions {
		sessionInfos = append(sessionInfos, SessionInfo{
			SessionID: session.ID.String(),
			UserAgent: session.UserAgent,
			ClientIP:  session.ClientIp,
			IsBlocked: session.IsBlocked,
			ExpiresAt: session.ExpiresAt,
			CreatedAt: session.CreatedAt,
			IsCurrent: session.ID.String() == currentSessionID,
		})
	}

	return sessionInfos, nil
}

// RevokeSession 撤销指定 session
func (s *service) RevokeSession(ctx context.Context, userID int64, sessionID string) error {
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return errors.New("invalid session id")
	}

	// 验证 session 属于该用户
	session, err := s.repo.GetSession(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("session not found")
		}
		return fmt.Errorf("failed to get session: %w", err)
	}

	if session.UserID != userID {
		return errors.New("unauthorized to revoke this session")
	}

	return s.repo.DeleteSession(ctx, id)
}

// RevokeAllOtherSessions 撤销除当前 session 外的所有 session
func (s *service) RevokeAllOtherSessions(ctx context.Context, userID int64, currentSessionID string) error {
	currentID, err := uuid.Parse(currentSessionID)
	if err != nil {
		return errors.New("invalid current session id")
	}

	// 获取所有 session
	sessions, err := s.repo.GetUserSessions(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	// 删除除当前 session 外的所有 session
	for _, session := range sessions {
		if session.ID != currentID {
			_ = s.repo.DeleteSession(ctx, session.ID)
		}
	}

	return nil
}

// toUserResponse 转换为响应对象
func toUserResponse(user sqlc.User) *UserResponse {
	return &UserResponse{
		ID:              user.ID,
		Username:        user.Username,
		Email:           user.Email,
		Phone:           stringPtrToString(user.Phone),
		Nickname:        stringPtrToString(user.Nickname),
		Avatar:          stringPtrToString(user.Avatar),
		Gender:          user.Gender,
		IsEmailVerified: user.IsEmailVerified,
		CreatedAt:       user.CreatedAt,
	}
}

// stringPtrToString 将 *string 转为 string（nil 转为空字符串）
func stringPtrToString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}
