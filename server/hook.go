package server

import "context"

type Hook func(ctx context.Context) error
