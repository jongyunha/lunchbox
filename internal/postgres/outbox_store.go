package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jongyunha/lunchbox/internal/am"
	"github.com/jongyunha/lunchbox/internal/ddd"
	"github.com/jongyunha/lunchbox/internal/tm"
	"github.com/stackus/errors"
)

type OutboxStore struct {
	queries *Queries
}

type outboxMessage struct {
	id       string
	name     string
	subject  string
	data     []byte
	metadata ddd.Metadata
	sentAt   time.Time
}

var _ tm.OutboxStore = (*OutboxStore)(nil)
var _ am.Message = (*outboxMessage)(nil)

func NewOutboxStore(db DBTX) *OutboxStore {
	return &OutboxStore{
		queries: New(db),
	}
}

func (o *OutboxStore) Save(ctx context.Context, msg am.Message) error {
	metadata, err := json.Marshal(msg.Metadata())
	if err != nil {
		return err
	}
	param := SaveRestaurantOutboxMessageParams{
		ID:       msg.ID(),
		Name:     msg.MessageName(),
		Subject:  msg.Subject(),
		Data:     msg.Data(),
		Metadata: metadata,
		SentAt:   msg.SentAt(),
	}
	_, err = o.queries.SaveRestaurantOutboxMessage(ctx, param)
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

func (o *OutboxStore) FindUnpublished(ctx context.Context, limit int) ([]am.Message, error) {
	rows, err := o.queries.FindRestaurantUnpublishedOutboxMessages(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	messages := make([]am.Message, len(rows))
	for i, row := range rows {
		var metadata ddd.Metadata
		err = json.Unmarshal(row.Metadata, &metadata)
		if err != nil {
			return nil, err
		}
		outbox := outboxMessage{
			id:       row.ID,
			name:     row.Name,
			subject:  row.Subject,
			data:     row.Data,
			metadata: metadata,
			sentAt:   row.SentAt,
		}
		messages[i] = outbox
	}

	return messages, nil
}

func (o *OutboxStore) MarkPublished(ctx context.Context, ids ...string) error {
	return o.queries.MarkRestaurantOutboxMessageAsPublishedByIDs(ctx, ids)
}

func (r outboxMessage) MessageName() string {
	return r.name
}

func (r outboxMessage) Subject() string {
	return r.subject
}

func (r outboxMessage) Data() []byte {
	return r.data
}

func (r outboxMessage) ID() string {
	return r.id
}

func (r outboxMessage) Metadata() ddd.Metadata {
	return r.metadata
}

func (r outboxMessage) SentAt() time.Time {
	return r.sentAt
}
