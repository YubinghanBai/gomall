package category

import (
	"gomall/db/sqlc"
	"gomall/utils"
	"time"
)

// Request DTOs

type CreateCategoryRequest struct {
	Name     string `json:"name" binding:"required,min=1,max=100"`
	Slug     string `json:"slug" binding:"omitempty,max=100"`
	ParentID *int64 `json:"parent_id"`
	Icon     string `json:"icon" binding:"omitempty,max=500"`
	Sort     int32  `json:"sort"`
}

type UpdateCategoryRequest struct {
	Name     *string `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Slug     *string `json:"slug,omitempty" binding:"omitempty,max=100"`
	ParentID *int64  `json:"parent_id,omitempty"`
	Icon     *string `json:"icon,omitempty" binding:"omitempty,max=500"`
	Sort     *int32  `json:"sort,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// Response DTOs

type CategoryResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug,omitempty"`
	ParentID  *int64    `json:"parent_id,omitempty"`
	Icon      string    `json:"icon,omitempty"`
	Sort      int32     `json:"sort"`
	Level     int16     `json:"level"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CategoryTreeNode struct {
	ID       int64              `json:"id"`
	Name     string             `json:"name"`
	Slug     string             `json:"slug,omitempty"`
	Icon     string             `json:"icon,omitempty"`
	Sort     int32              `json:"sort"`
	Level    int16              `json:"level"`
	Children []CategoryTreeNode `json:"children,omitempty"`
}

// Conversion functions

func toCategoryResponse(category sqlc.Category) CategoryResponse {
	return CategoryResponse{
		ID:        category.ID,
		Name:      category.Name,
		Slug:      utils.PtrValue(category.Slug),
		ParentID:  category.ParentID,
		Icon:      utils.PtrValue(category.Icon),
		Sort:      category.Sort,
		Level:     category.Level,
		IsActive:  category.IsActive,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}
}

func toCategoryTreeNode(category sqlc.Category) CategoryTreeNode {
	return CategoryTreeNode{
		ID:       category.ID,
		Name:     category.Name,
		Slug:     utils.PtrValue(category.Slug),
		Icon:     utils.PtrValue(category.Icon),
		Sort:     category.Sort,
		Level:    category.Level,
		Children: []CategoryTreeNode{},
	}
}
