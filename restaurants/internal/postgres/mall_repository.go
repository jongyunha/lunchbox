package postgres

import (
	"context"

	"github.com/jongyunha/lunchbox/internal/postgres"
	"github.com/jongyunha/lunchbox/restaurants/internal/domain"
)

type MallRepository struct {
	queries *postgres.Queries
}

var _ domain.MallRepository = (*MallRepository)(nil)

func NewMallRepository(db postgres.DBTX) MallRepository {
	return MallRepository{
		queries: postgres.New(db),
	}
}

func (m MallRepository) RegisterRestaurant(ctx context.Context, restaurantID, name string) error {
	return m.queries.SaveRestaurant(ctx, postgres.SaveRestaurantParams{
		ID:   restaurantID,
		Name: name,
	})
}

func (m MallRepository) FindByID(ctx context.Context, restaurantID string) (*domain.MallRestaurant, error) {
	panic("implement me")
}
