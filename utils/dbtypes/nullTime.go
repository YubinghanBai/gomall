package dbtypes

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// NullTime 可空时间类型，兼容 sql.Scanner 和 driver.Valuer
type NullTime struct {
	Time  time.Time
	Valid bool
}

// Scan 实现 sql.Scanner 接口（pgx 会调用）
func (nt *NullTime) Scan(value interface{}) error {
	if value == nil {
		nt.Time, nt.Valid = time.Time{}, false
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		nt.Time, nt.Valid = v, true
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into NullTime", value)
	}
}

// Value 实现 driver.Valuer 接口
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

// Ptr 返回 *time.Time，如果无效返回 nil
func (nt NullTime) Ptr() *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}

// ValueOr 返回时间值，如果无效返回默认值
func (nt NullTime) ValueOr(defaultValue time.Time) time.Time {
	if !nt.Valid {
		return defaultValue
	}
	return nt.Time
}

// NewNullTime 从 *time.Time 创建 NullTime
func NewNullTime(t *time.Time) NullTime {
	if t == nil {
		return NullTime{Valid: false}
	}
	return NullTime{Time: *t, Valid: true}
}

// NewNullTimeValue 从 time.Time 创建有效的 NullTime
func NewNullTimeValue(t time.Time) NullTime {
	return NullTime{Time: t, Valid: true}
}
