package category

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"gomall/db/sqlc"
)

// Service defines the business logic interface for category domain
type Service interface {
	CreateCategory(ctx context.Context, req CreateCategoryRequest) (*CategoryResponse, error)
	GetCategory(ctx context.Context, id int64) (*CategoryResponse, error)
	UpdateCategory(ctx context.Context, id int64, req UpdateCategoryRequest) error
	DeleteCategory(ctx context.Context, id int64) error

	ListCategories(ctx context.Context, isActive *bool) ([]CategoryResponse, error)
	GetCategoryTree(ctx context.Context) ([]CategoryTreeNode, error)
	GetChildren(ctx context.Context, parentID int64) ([]CategoryResponse, error)
	GetRoots(ctx context.Context) ([]CategoryResponse, error)
}

type service struct {
	repo Repository
}

// NewService creates a new Service instance
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// CreateCategory creates a new category
func (s *service) CreateCategory(ctx context.Context, req CreateCategoryRequest) (*CategoryResponse, error) {
	// Calculate level based on parent
	level := int16(1)
	if req.ParentID != nil {
		parent, err := s.repo.GetCategoryByID(ctx, *req.ParentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, errors.New("parent category not found")
			}
			return nil, fmt.Errorf("failed to get parent category: %w", err)
		}
		level = parent.Level + 1
	}

	// Generate slug if not provided
	slug := req.Slug
	if slug == "" {
		slug = generateSlug(req.Name)
	}

	// Create category
	category, err := s.repo.CreateCategory(ctx, sqlc.CreateCategoryParams{
		Name:     req.Name,
		Slug:     &slug,
		ParentID: req.ParentID,
		Icon:     &req.Icon,
		Sort:     req.Sort,
		Level:    level,
		IsActive: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	response := toCategoryResponse(category)
	return &response, nil
}

// GetCategory retrieves a category by ID
func (s *service) GetCategory(ctx context.Context, id int64) (*CategoryResponse, error) {
	category, err := s.repo.GetCategoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("category not found")
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	response := toCategoryResponse(category)
	return &response, nil
}

// UpdateCategory updates a category
func (s *service) UpdateCategory(ctx context.Context, id int64, req UpdateCategoryRequest) error {
	// Check if category exists
	_, err := s.repo.GetCategoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("category not found")
		}
		return fmt.Errorf("failed to get category: %w", err)
	}

	// If parent_id is being updated, verify new parent exists
	if req.ParentID != nil {
		if *req.ParentID == id {
			return errors.New("category cannot be its own parent")
		}
		_, err := s.repo.GetCategoryByID(ctx, *req.ParentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("parent category not found")
			}
			return fmt.Errorf("failed to get parent category: %w", err)
		}
	}

	// Update category
	err = s.repo.UpdateCategory(ctx, sqlc.UpdateCategoryParams{
		Name:     req.Name,
		Slug:     req.Slug,
		ParentID: req.ParentID,
		Icon:     req.Icon,
		Sort:     req.Sort,
		IsActive: req.IsActive,
		ID:       id,
	})
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	return nil
}

// DeleteCategory soft deletes a category
func (s *service) DeleteCategory(ctx context.Context, id int64) error {
	// Check if category has children
	childCount, err := s.repo.CountCategoryChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to count children: %w", err)
	}
	if childCount > 0 {
		return errors.New("cannot delete category with children")
	}

	// Check if category has products
	productCount, err := s.repo.CountProductsByCategory(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to count products: %w", err)
	}
	if productCount > 0 {
		return errors.New("cannot delete category with products")
	}

	// Delete category
	err = s.repo.DeleteCategory(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}

// ListCategories lists all categories with optional filter
func (s *service) ListCategories(ctx context.Context, isActive *bool) ([]CategoryResponse, error) {
	categories, err := s.repo.ListCategories(ctx, isActive)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	responses := make([]CategoryResponse, len(categories))
	for i, c := range categories {
		responses[i] = toCategoryResponse(c)
	}

	return responses, nil
}

// GetCategoryTree builds a hierarchical tree of all active categories
func (s *service) GetCategoryTree(ctx context.Context) ([]CategoryTreeNode, error) {
	// Get all active categories
	isActive := true
	categories, err := s.repo.ListCategories(ctx, &isActive)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	// Build category map
	categoryMap := make(map[int64]CategoryTreeNode)
	for _, c := range categories {
		categoryMap[c.ID] = toCategoryTreeNode(c)
	}

	// Build tree structure
	var roots []CategoryTreeNode
	for _, c := range categories {
		node := categoryMap[c.ID]

		if c.ParentID == nil {
			// Root category
			roots = append(roots, node)
		} else {
			// Child category - attach to parent
			if parent, exists := categoryMap[*c.ParentID]; exists {
				parent.Children = append(parent.Children, node)
				categoryMap[*c.ParentID] = parent
			}
		}
	}

	return roots, nil
}

// GetChildren gets all children of a category
func (s *service) GetChildren(ctx context.Context, parentID int64) ([]CategoryResponse, error) {
	categories, err := s.repo.GetCategoryChildren(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get children: %w", err)
	}

	responses := make([]CategoryResponse, len(categories))
	for i, c := range categories {
		responses[i] = toCategoryResponse(c)
	}

	return responses, nil
}

// GetRoots gets all root categories
func (s *service) GetRoots(ctx context.Context) ([]CategoryResponse, error) {
	categories, err := s.repo.GetRootCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get root categories: %w", err)
	}

	responses := make([]CategoryResponse, len(categories))
	for i, c := range categories {
		responses[i] = toCategoryResponse(c)
	}

	return responses, nil
}

// Helper functions

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters (basic implementation)
	return slug
}
