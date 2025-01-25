package postgres

import (
	"context"

	"github.com/jongyunha/lunchbox/category/internal/domain"
	"github.com/jongyunha/lunchbox/internal/postgres"
)

type CategoryRepository struct {
	queries *postgres.Queries
}

var _ domain.CategoryRepository = (*CategoryRepository)(nil)

func NewCategoryRepository(queries *postgres.Queries) CategoryRepository {
	return CategoryRepository{
		queries: queries,
	}
}

func (c CategoryRepository) Save(ctx context.Context, category *domain.Category) error {
	return nil
}
