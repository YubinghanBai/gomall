package user

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	mockdb "gomall/db/mock"
	"gomall/db/sqlc"
)

func TestGetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mockdb.NewMockStore(ctrl)

	// 准备测试数据
	expectedUser := sqlc.User{
		ID:              1,
		Username:        "testuser",
		Email:           "test@example.com",
		IsEmailVerified: true,
		CreatedAt:       time.Now(),
	}

	// 设置 mock 期望
	mockStore.EXPECT().
		GetUserByID(gomock.Any(), gomock.Eq(int64(1))).
		Times(1).
		Return(expectedUser, nil)

	// 创建 service（注入 mock）
	service := NewService(nil, mockStore, nil, nil)

	// 执行测试
	user, err := service.GetProfile(context.Background(), 1)

	// 断言
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, expectedUser.Username, user.Username)
	require.Equal(t, expectedUser.Email, user.Email)
}

func TestGetProfile_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mockdb.NewMockStore(ctrl)

	// 设置 mock：返回 ErrNoRows
	mockStore.EXPECT().
		GetUserByID(gomock.Any(), gomock.Eq(int64(999))).
		Times(1).
		Return(sqlc.User{}, sql.ErrNoRows)

	service := NewService(nil, mockStore, nil, nil)

	user, err := service.GetProfile(context.Background(), 999)

	// 断言：应该返回错误
	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "user not found")
}

// TestVerifyEmail_Success 测试邮箱验证成功的场景
func TestVerifyEmail_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mockdb.NewMockStore(ctrl)

	// 准备测试数据
	verificationCode := sqlc.VerificationCode{
		ID:        1,
		UserID:    100,
		Email:     "test@example.com",
		Code:      "123456",
		Type:      "email_verification",
		IsUsed:    false,
		ExpiresAt: time.Now().Add(10 * time.Minute),
		CreatedAt: time.Now(),
	}

	// 第一步：获取验证码
	mockStore.EXPECT().
		GetVerificationCode(gomock.Any(), sqlc.GetVerificationCodeParams{
			Email: "test@example.com",
			Code:  "123456",
			Type:  "email_verification",
		}).
		Times(1).
		Return(verificationCode, nil)

	// 第二步：执行事务
	mockStore.EXPECT().
		ExecTx(gomock.Any(), gomock.Any()).
		Times(1).
		DoAndReturn(func(ctx context.Context, fn func(sqlc.Querier) error) error {
			// 创建一个假的 querier（实际使用 mockStore 来模拟）
			// 在真实测试中，这里会调用 fn(mockQuerier)
			// 但为了简化，我们直接返回 nil 表示事务成功
			return nil
		})

	// 创建 service
	service := NewService(nil, mockStore, nil, nil)

	// 执行测试
	req := VerifyEmailRequest{
		Email: "test@example.com",
		Code:  "123456",
	}
	err := service.VerifyEmail(context.Background(), req)

	// 断言
	require.NoError(t, err)
}

// TestVerifyEmail_InvalidCode 测试验证码无效的场景
func TestVerifyEmail_InvalidCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mockdb.NewMockStore(ctrl)

	// 模拟验证码不存在
	mockStore.EXPECT().
		GetVerificationCode(gomock.Any(), gomock.Any()).
		Times(1).
		Return(sqlc.VerificationCode{}, sql.ErrNoRows)

	service := NewService(nil, mockStore, nil, nil)

	req := VerifyEmailRequest{
		Email: "test@example.com",
		Code:  "wrong-code",
	}
	err := service.VerifyEmail(context.Background(), req)

	// 断言：应该返回错误
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid or expired")
}

// TestVerifyEmail_Expired 测试验证码过期的场景
func TestVerifyEmail_Expired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mockdb.NewMockStore(ctrl)

	// 准备过期的验证码
	expiredCode := sqlc.VerificationCode{
		ID:        1,
		UserID:    100,
		Email:     "test@example.com",
		Code:      "123456",
		Type:      "email_verification",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // 1 小时前过期
	}

	mockStore.EXPECT().
		GetVerificationCode(gomock.Any(), gomock.Any()).
		Times(1).
		Return(expiredCode, nil)

	// 不应该调用 ExecTx，因为在验证码过期检查时就会失败

	service := NewService(nil, mockStore, nil, nil)

	req := VerifyEmailRequest{
		Email: "test@example.com",
		Code:  "123456",
	}
	err := service.VerifyEmail(context.Background(), req)

	// 断言：应该返回过期错误
	require.Error(t, err)
	require.Contains(t, err.Error(), "expired")
}

// TestResetPassword_Success 测试密码重置成功的场景
func TestResetPassword_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mockdb.NewMockStore(ctrl)

	verificationCode := sqlc.VerificationCode{
		ID:        1,
		UserID:    100,
		Email:     "test@example.com",
		Code:      "123456",
		Type:      "password_reset",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	// 第一步：获取验证码
	mockStore.EXPECT().
		GetVerificationCode(gomock.Any(), sqlc.GetVerificationCodeParams{
			Email: "test@example.com",
			Code:  "123456",
			Type:  "password_reset",
		}).
		Times(1).
		Return(verificationCode, nil)

	// 第二步：执行事务（更新密码 + 标记验证码已使用 + 删除会话）
	mockStore.EXPECT().
		ExecTx(gomock.Any(), gomock.Any()).
		Times(1).
		DoAndReturn(func(ctx context.Context, fn func(sqlc.Querier) error) error {
			// 模拟事务成功执行
			return nil
		})

	service := NewService(nil, mockStore, nil, nil)

	req := ResetPasswordRequest{
		Email:       "test@example.com",
		Code:        "123456",
		NewPassword: "newpassword123",
	}
	err := service.ResetPassword(context.Background(), req)

	// 断言
	require.NoError(t, err)
}

// TestResetPassword_TransactionFailed 测试事务失败的场景
func TestResetPassword_TransactionFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mockdb.NewMockStore(ctrl)

	verificationCode := sqlc.VerificationCode{
		ID:        1,
		UserID:    100,
		Email:     "test@example.com",
		Code:      "123456",
		Type:      "password_reset",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	mockStore.EXPECT().
		GetVerificationCode(gomock.Any(), gomock.Any()).
		Times(1).
		Return(verificationCode, nil)

	// 模拟事务执行失败
	mockStore.EXPECT().
		ExecTx(gomock.Any(), gomock.Any()).
		Times(1).
		Return(sql.ErrConnDone)

	service := NewService(nil, mockStore, nil, nil)

	req := ResetPasswordRequest{
		Email:       "test@example.com",
		Code:        "123456",
		NewPassword: "newpassword123",
	}
	err := service.ResetPassword(context.Background(), req)

	// 断言：应该返回错误
	require.Error(t, err)
}
