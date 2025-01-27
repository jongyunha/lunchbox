package postgres

import (
	"context"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jongyunha/lunchbox/internal/am"
	"github.com/jongyunha/lunchbox/internal/tm"
	"github.com/stackus/errors"
)

type OutboxStore struct {
	queries *Queries
}

type outboxMessage struct {
	id      string
	name    string
	subject string
	data    []byte
}

var _ tm.OutBoxStore = (*OutboxStore)(nil)
var _ am.RawMessage = (*outboxMessage)(nil)

func NewOutboxStore(db DBTX) OutboxStore {
	return OutboxStore{
		queries: New(db),
	}
}

func (o OutboxStore) Save(ctx context.Context, msg am.RawMessage) error {
	param := SaveRestaurantOutboxMessageParams{
		ID:      msg.ID(),
		Name:    msg.MessageName(),
		Subject: msg.Subject(),
		Data:    msg.Data(),
	}
	_, err := o.queries.SaveRestaurantOutboxMessage(ctx, param)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return tm.ErrDuplicateMessage(msg.ID())
			}
		}
	}

	return err
}

func (o OutboxStore) FindUnpublished(ctx context.Context, limit int) ([]am.RawMessage, error) {
	rows, err := o.queries.FindRestaurantUnpublishedOutboxMessages(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	messages := make([]am.RawMessage, len(rows))
	for i, row := range rows {
		outbox := outboxMessage{
			id:      row.ID,
			name:    row.Name,
			subject: row.Subject,
			data:    row.Data,
		}
		messages[i] = outbox
	}

	return messages, nil
}

func (o OutboxStore) MarkPublished(ctx context.Context, ids ...string) error {
	return o.queries.MarkRestaurantOutboxMessageAsPublishedByIDs(ctx, ids)
}

func (r outboxMessage) MessageName() string {
	return r.MessageName()
}

func (r outboxMessage) Subject() string {
	return r.Subject()
}

func (r outboxMessage) Data() []byte {
	return r.Data()
}

func (r outboxMessage) ID() string {
	return r.ID()
}
