package tm

import (
	"context"
	"errors"

	"github.com/jongyunha/lunchbox/internal/am"
)

type OutBoxStore interface {
	Save(ctx context.Context, msg am.RawMessage) error
	FindUnpublished(ctx context.Context, limit int) ([]am.RawMessage, error)
	MarkPublished(ctx context.Context, ids ...string) error
}

type outbox struct {
	am.RawMessageStream
	store OutBoxStore
}

var _ am.RawMessageStream = (*outbox)(nil)

func NewOutboxStreamMiddleware(store OutBoxStore) am.RawMessageStreamMiddleware {
	o := outbox{store: store}

	return func(stream am.RawMessageStream) am.RawMessageStream {
		o.RawMessageStream = stream

		return o
	}
}

func (o outbox) Publish(ctx context.Context, topicName string, msg am.RawMessage) error {
	err := o.store.Save(ctx, msg)

	var errDupe ErrDuplicateMessage
	if errors.As(err, &errDupe) {
		return nil
	}

	return err
}
