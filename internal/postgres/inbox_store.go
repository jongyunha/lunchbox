package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jongyunha/lunchbox/internal/am"
	"github.com/jongyunha/lunchbox/internal/tm"
)

type InboxStore struct {
	db      *pgxpool.Pool
	queries *Queries
}

var _ tm.InboxStore = (*InboxStore)(nil)

func NewInboxStore(db *pgxpool.Pool) InboxStore {
	return InboxStore{
		db:      db,
		queries: New(db),
	}
}

func (i InboxStore) Save(ctx context.Context, msg am.RawMessage) error {
	_, err := i.queries.SaveRestaurantInboxMessage(ctx, SaveRestaurantInboxMessageParams{
		ID:      msg.ID(),
		Name:    msg.MessageName(),
		Subject: msg.Subject(),
		Data:    msg.Data(),
	})
	return err
}
