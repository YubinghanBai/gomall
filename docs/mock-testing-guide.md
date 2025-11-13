# Mock 测试指南

## 一、Mock 生成

### 生成 Store Mock

```bash
# 生成 Store 接口的 mock
make mock

# 生成的文件位置
db/mock/store.go
```

### 为其他接口生成 Mock

如果你想为其他接口生成 mock（如 Service 接口）：

```bash
# 在 internal/user/service.go 中添加
//go:generate mockgen -package mockuser -destination mock/service_mock.go . Service

# 然后运行
go generate ./...
```

---

## 二、Mock 使用示例

### 1. 基础查询测试

```go
// internal/user/service_test.go
package user

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	mockdb "gomall/db/mock"
	"gomall/db/sqlc"
)

func TestGetProfile(t *testing.T) {
	// 1. 创建 gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 2. 创建 mock store
	mockStore := mockdb.NewMockStore(ctrl)

	// 3. 设置期望调用
	expectedUser := sqlc.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	mockStore.EXPECT().
		GetUserByID(gomock.Any(), gomock.Eq(int64(1))).
		Times(1).
		Return(expectedUser, nil)

	// 4. 创建 service（注入 mock）
	service := NewService(nil, mockStore, nil, nil)

	// 5. 执行测试
	user, err := service.GetProfile(context.Background(), 1)

	// 6. 断言
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, expectedUser.Username, user.Username)
	require.Equal(t, expectedUser.Email, user.Email)
}
```

---

### 2. 事务测试（重点 🌟）

**测试 VerifyEmail 事务：**

```go
func TestVerifyEmail(t *testing.T) {
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
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	// 设置期望：获取验证码
	mockStore.EXPECT().
		GetVerificationCode(gomock.Any(), gomock.Any()).
		Times(1).
		Return(verificationCode, nil)

	// 设置期望：执行事务
	mockStore.EXPECT().
		ExecTx(gomock.Any(), gomock.Any()).
		Times(1).
		DoAndReturn(func(ctx context.Context, fn func(sqlc.Querier) error) error {
			// 模拟事务执行
			// 创建一个 mock Querier 来验证事务内部的调用
			mockQuerier := mockdb.NewMockQuerier(ctrl)

			// 期望调用 MarkCodeAsUsed
			mockQuerier.EXPECT().
				MarkCodeAsUsed(gomock.Any(), verificationCode.ID).
				Times(1).
				Return(nil)

			// 期望调用 VerifyUserEmail
			mockQuerier.EXPECT().
				VerifyUserEmail(gomock.Any(), verificationCode.UserID).
				Times(1).
				Return(nil)

			// 执行传入的函数
			return fn(mockQuerier)
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
```

---

### 3. 测试事务回滚

```go
func TestVerifyEmail_TransactionRollback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mockdb.NewMockStore(ctrl)

	verificationCode := sqlc.VerificationCode{
		ID:        1,
		UserID:    100,
		Email:     "test@example.com",
		Code:      "123456",
		Type:      "email_verification",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	mockStore.EXPECT().
		GetVerificationCode(gomock.Any(), gomock.Any()).
		Times(1).
		Return(verificationCode, nil)

	// 模拟事务中的错误
	mockStore.EXPECT().
		ExecTx(gomock.Any(), gomock.Any()).
		Times(1).
		DoAndReturn(func(ctx context.Context, fn func(sqlc.Querier) error) error {
			mockQuerier := mockdb.NewMockQuerier(ctrl)

			// 第一步成功
			mockQuerier.EXPECT().
				MarkCodeAsUsed(gomock.Any(), verificationCode.ID).
				Times(1).
				Return(nil)

			// 第二步失败（模拟数据库错误）
			mockQuerier.EXPECT().
				VerifyUserEmail(gomock.Any(), verificationCode.UserID).
				Times(1).
				Return(sql.ErrConnDone)

			// 执行函数（会返回错误）
			return fn(mockQuerier)
		})

	service := NewService(nil, mockStore, nil, nil)

	req := VerifyEmailRequest{
		Email: "test@example.com",
		Code:  "123456",
	}
	err := service.VerifyEmail(context.Background(), req)

	// 断言：应该返回错误
	require.Error(t, err)
}
```

---

### 4. 测试 ResetPassword 事务

```go
func TestResetPassword(t *testing.T) {
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

	// 期望：获取验证码
	mockStore.EXPECT().
		GetVerificationCode(gomock.Any(), gomock.Any()).
		Times(1).
		Return(verificationCode, nil)

	// 期望：执行事务（包含 3 个操作）
	mockStore.EXPECT().
		ExecTx(gomock.Any(), gomock.Any()).
		Times(1).
		DoAndReturn(func(ctx context.Context, fn func(sqlc.Querier) error) error {
			mockQuerier := mockdb.NewMockQuerier(ctrl)

			// 1. 更新密码
			mockQuerier.EXPECT().
				UpdateUserPassword(gomock.Any(), gomock.Any()).
				Times(1).
				Return(nil)

			// 2. 标记验证码已使用
			mockQuerier.EXPECT().
				MarkCodeAsUsed(gomock.Any(), verificationCode.ID).
				Times(1).
				Return(nil)

			// 3. 删除所有会话
			mockQuerier.EXPECT().
				DeleteUserSessions(gomock.Any(), verificationCode.UserID).
				Times(1).
				Return(nil)

			return fn(mockQuerier)
		})

	service := NewService(nil, mockStore, nil, nil)

	req := ResetPasswordRequest{
		Email:       "test@example.com",
		Code:        "123456",
		NewPassword: "newpassword123",
	}
	err := service.ResetPassword(context.Background(), req)

	require.NoError(t, err)
}
```

---

## 三、常用 Mock 模式

### 1. 匹配任意参数

```go
mockStore.EXPECT().
	GetUserByID(gomock.Any(), gomock.Any()).
	Return(user, nil)
```

### 2. 匹配特定值

```go
mockStore.EXPECT().
	GetUserByID(gomock.Any(), gomock.Eq(int64(1))).
	Return(user, nil)
```

### 3. 自定义匹配器

```go
mockStore.EXPECT().
	CreateUser(gomock.Any(), gomock.AssignableToTypeOf(sqlc.CreateUserParams{})).
	Return(user, nil)
```

### 4. 验证调用次数

```go
// 必须调用 1 次
mockStore.EXPECT().
	GetUserByID(gomock.Any(), gomock.Any()).
	Times(1)

// 可以调用 0 次或多次
mockStore.EXPECT().
	GetUserByID(gomock.Any(), gomock.Any()).
	AnyTimes()

// 至少调用 1 次
mockStore.EXPECT().
	GetUserByID(gomock.Any(), gomock.Any()).
	MinTimes(1)
```

### 5. 验证调用顺序

```go
gomock.InOrder(
	mockStore.EXPECT().GetUserByID(gomock.Any(), gomock.Any()),
	mockStore.EXPECT().UpdateUser(gomock.Any(), gomock.Any()),
)
```

### 6. 返回不同结果

```go
mockStore.EXPECT().
	GetUserByID(gomock.Any(), gomock.Eq(int64(1))).
	Return(user1, nil)

mockStore.EXPECT().
	GetUserByID(gomock.Any(), gomock.Eq(int64(2))).
	Return(user2, nil)
```

---

## 四、表驱动测试

```go
func TestGetProfile_TableDriven(t *testing.T) {
	testCases := []struct {
		name          string
		userID        int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, user *UserResponse, err error)
	}{
		{
			name:   "Success",
			userID: 1,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByID(gomock.Any(), gomock.Eq(int64(1))).
					Times(1).
					Return(sqlc.User{
						ID:       1,
						Username: "testuser",
						Email:    "test@example.com",
					}, nil)
			},
			checkResponse: func(t *testing.T, user *UserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, user)
				require.Equal(t, "testuser", user.Username)
			},
		},
		{
			name:   "UserNotFound",
			userID: 999,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByID(gomock.Any(), gomock.Eq(int64(999))).
					Times(1).
					Return(sqlc.User{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, user *UserResponse, err error) {
				require.Error(t, err)
				require.Nil(t, user)
				require.Equal(t, "user not found", err.Error())
			},
		},
		{
			name:   "DatabaseError",
			userID: 1,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByID(gomock.Any(), gomock.Eq(int64(1))).
					Times(1).
					Return(sqlc.User{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, user *UserResponse, err error) {
				require.Error(t, err)
				require.Nil(t, user)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mockdb.NewMockStore(ctrl)
			tc.buildStubs(mockStore)

			service := NewService(nil, mockStore, nil, nil)
			user, err := service.GetProfile(context.Background(), tc.userID)

			tc.checkResponse(t, user, err)
		})
	}
}
```

---

## 五、运行测试

```bash
# 运行所有测试
make test

# 运行特定包的测试
go test -v ./internal/user/...

# 运行特定测试
go test -v -run TestVerifyEmail ./internal/user/...

# 生成覆盖率报告
make test-coverage
```

---

## 六、注意事项

### 1. Mock Querier 问题

如果你需要 mock Querier（在测试事务时），需要单独生成：

```bash
# 添加到 Makefile
mock-querier:
	mockgen -package mockdb -destination db/mock/querier.go gomall/db/sqlc Querier
```

或者使用 `mockStore.EXPECT().ExecTx().DoAndReturn()` 来模拟。

### 2. 不要过度 Mock

```go
// ❌ 不要：Mock 太细
mockStore.EXPECT().GetUserByID(...).Times(1)
mockStore.EXPECT().GetUserByEmail(...).Times(1)
mockStore.EXPECT().GetUserByUsername(...).Times(1)
// ... 100 行 mock 设置

// ✅ 正确：只 mock 关键路径
mockStore.EXPECT().CreateUser(...).Return(user, nil)
```

### 3. 使用 testify 简化断言

```bash
# 安装 testify
go get github.com/stretchr/testify
```

```go
import "github.com/stretchr/testify/require"

// 使用 require 替代 if err != nil
require.NoError(t, err)
require.Equal(t, expected, actual)
require.NotNil(t, user)
```

---

## 七、完整测试文件示例

见 `internal/user/service_test.go`（下一个文件）