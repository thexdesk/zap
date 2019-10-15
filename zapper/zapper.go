package zapper

import (
	"context"
	"io"
)

type Zapper interface {
	Zap(ctx context.Context, src, dst string) error
}

type Zapped interface {
	io.Closer

	Listen(ctx context.Context, port int) (io.ReadCloser, error)
}
