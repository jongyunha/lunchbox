package application

import (
	"context"

	"github.com/jongyunha/lunchbox/restaurants/internal/application/commands"
	"github.com/jongyunha/lunchbox/restaurants/internal/domain"
)

type (
	App interface {
		Commands
		Queries
	}

	Commands interface {
		RegisterRestaurant(ctx context.Context, cmd commands.RegisterRestaurant) error
	}

	Queries interface {
	}

	Application struct {
		appCommands
		appQueries
	}

	appCommands struct {
		commands.RegisterRestaurantHandler
	}

	appQueries struct {
	}
)

var _ App = (*Application)(nil)

func New(
	restaurants domain.RestaurantRepository,
) *Application {
	return &Application{
		appCommands: appCommands{
			RegisterRestaurantHandler: commands.NewRegisterRestaurantHandler(restaurants),
		},
	}
}
