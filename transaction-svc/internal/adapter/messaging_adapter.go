package adapter

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/nats-io/nats.go"
)

type MessagingAdapter interface {
	Publish(ctx context.Context, subject string, data any) error
}

type messagingAdapter struct {
	js nats.JetStreamContext
}

func NewMessagingAdapter(js nats.JetStreamContext) MessagingAdapter {
	return &messagingAdapter{js: js}
}

func (n *messagingAdapter) Publish(ctx context.Context, subject string, data any) error {
	payload, err := sonic.ConfigFastest.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	_, err = n.js.Publish(subject, payload)
	if err != nil {
		return fmt.Errorf("failed to publish to subject %q: %w", subject, err)
	}

	return nil
}
