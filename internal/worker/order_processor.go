package worker

import "context"

type OrderProcessor interface {
	Process(ctx context.Context, ev OrderEvent) error
}
