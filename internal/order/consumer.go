package order

import "context"

type Consumer interface {
	Subscribe(ctx context.Context) error
}
