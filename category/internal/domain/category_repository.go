package domain

import "context"

type CategoryRepository interface {
	Save(ctx context.Context, category *Category) error
}
