package ports

import "context"

type EventConsumer interface {
	Listen(ctx context.Context) error
}
