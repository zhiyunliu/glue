package engine

import "context"

type Hook func(ctx context.Context) error
