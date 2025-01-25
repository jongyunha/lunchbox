package handlers

import (
	"context"

	"github.com/jongyunha/lunchbox/category/internal/application"
	"github.com/jongyunha/lunchbox/internal/ddd"
)

type commandHandlers struct {
	app application.App
}

func NewCommandHandlers(app application.App) ddd.CommandHandler[ddd.Command] {
	return &commandHandlers{
		app: app,
	}
}

func (c commandHandlers) HandleCommand(ctx context.Context, cmd ddd.Command) (ddd.Reply, error) {
	return nil, nil
}
