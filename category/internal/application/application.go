package application

import (
	"context"

	"github.com/jongyunha/lunchbox/category/internal/domain"
	"github.com/jongyunha/lunchbox/internal/ddd"
)

type (
	RegisterCategory struct {
		ID   string
		Name string
	}

	App interface {
		RegisterCategory(ctx context.Context, register RegisterCategory) error
	}

	Application struct {
		categories      domain.CategoryRepository
		domainPublisher ddd.EventPublisher[ddd.AggregateEvent]
	}
)

var _ App = (*Application)(nil)

func New(categories domain.CategoryRepository, domainPublisher ddd.EventPublisher[ddd.AggregateEvent]) *Application {
	return &Application{
		categories:      categories,
		domainPublisher: domainPublisher,
	}
}

func (a Application) RegisterCategory(ctx context.Context, register RegisterCategory) error {
	category, err := domain.RegisterCategory(register.ID, register.Name)
	if err != nil {
		return err
	}

	if err = a.categories.Save(ctx, category); err != nil {
		return err
	}

	// publish domain event
	if err = a.domainPublisher.Publish(ctx, category.Events()...); err != nil {
		return err
	}

	return nil
}
