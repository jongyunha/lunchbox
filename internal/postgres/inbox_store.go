package postgres

import (
	"context"

	"github.com/jongyunha/lunchbox/internal/am"
	"github.com/jongyunha/lunchbox/internal/tm"
)

type InboxStore struct {
	db      DBTX
	queries *Queries
}

var _ tm.InboxStore = (*InboxStore)(nil)

func NewInboxStore(db DBTX) InboxStore {
	return InboxStore{
		db:      db,
		queries: New(db),
	}
}

func (i InboxStore) Save(ctx context.Context, msg am.IncomingMessage) error {
	_, err := i.queries.SaveRestaurantInboxMessage(ctx, SaveRestaurantInboxMessageParams{
		ID:      msg.ID(),
		Name:    msg.MessageName(),
		Subject: msg.Subject(),
		Data:    msg.Data(),
	})
	return err
}
