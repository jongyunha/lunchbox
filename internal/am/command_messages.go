package am

import (
	"time"

	"github.com/jongyunha/lunchbox/internal/ddd"
	"github.com/jongyunha/lunchbox/internal/registry"
)

type (
	CommandMessage interface {
		Message
		ddd.Command
	}

	IncomingCommandMessage interface {
		IncomingMessage
		ddd.Command
	}

	CommandPublisher  = MessagePublisher[ddd.Command]
	CommandSubscriber interface {
		Subscribe(topicName string, handler CommandMessageHandler, options ...SubscriberOption) error
	}
	CommandStream interface {
		MessagePublisher[ddd.Command]
		CommandSubscriber
	}

	commandStream struct {
		reg    registry.Registry
		stream RawMessageStream
	}

	commandMessage struct {
		id         string
		name       string
		payload    ddd.CommandPayload
		metadata   ddd.Metadata
		occurredAt time.Time
		msg        IncomingMessage
	}
)

var _ CommandMessage = (*commandMessage)(nil)

var _ CommandStream = (*commandStream)(nil)

func NewCommandStream(reg registry.Registry, stream RawMessageStream) CommandStream {
	return &commandStream{
		reg:    reg,
		stream: stream,
	}
}
