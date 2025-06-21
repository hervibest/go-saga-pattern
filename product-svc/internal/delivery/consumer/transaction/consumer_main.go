package consumer

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func (s *TransactionConsumer) ConsumeAllEvents(ctx context.Context) error {
	for _, subject := range s.subjects {
		if err := s.setupConsumer(subject); err != nil {
			return fmt.Errorf("failed to setup consumer for %s: %w", subject, err)
		}

		sub, err := s.js.PullSubscribe(
			subject,
			s.durableNames[subject],
			nats.BindStream("TRANSACTION_STREAM"),
		)
		if err != nil {
			return fmt.Errorf("failed to subscribe to %s: %w", subject, err)
		}

		go s.startConsumer(ctx, sub, subject)
	}

	return nil
}

func (s *TransactionConsumer) startConsumer(ctx context.Context, sub *nats.Subscription, subject string) {
	s.logs.Info("Started consumer for", zap.String("subject", subject))

	for {
		select {
		case <-ctx.Done():
			s.logs.Info("Stopping consumer", zap.String("subject", subject))
			return
		default:
			msgs, err := sub.Fetch(10, nats.MaxWait(2*time.Second))
			if err != nil && err != nats.ErrTimeout {
				s.logs.Error("fetch error", zap.String("subject", subject), zap.Error(err))
				continue
			}

			for _, msg := range msgs {
				s.handleMessage(ctx, msg)
			}
		}
	}
}
