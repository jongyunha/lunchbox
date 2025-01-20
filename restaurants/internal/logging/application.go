package logging

import (
	"context"

	"github.com/jongyunha/lunchbox/restaurants/internal/application"
	"github.com/jongyunha/lunchbox/restaurants/internal/application/commands"
	"github.com/rs/zerolog"
)

type Application struct {
	application.App
	logger zerolog.Logger
}

var _ application.App = (*Application)(nil)

func LogApplicationAccess(application application.App, logger zerolog.Logger) Application {
	return Application{
		App:    application,
		logger: logger,
	}
}

func (a Application) RegisterRestaurant(ctx context.Context, cmd commands.RegisterRestaurant) (err error) {
	a.logger.Info().Msg("--> Restaurant.RegisterRestaurant")
	defer func() { a.logger.Info().Err(err).Msg("<-- Restaurant.RegisterRestaurant") }()
	return a.App.RegisterRestaurant(ctx, cmd)
}
