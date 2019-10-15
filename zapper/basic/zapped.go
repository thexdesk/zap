package basic

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/hinshun/zap/zapper"
	"github.com/rs/zerolog"
)

type basicZapped struct {
	listener net.Listener
}

func NewZapped() (zapper.Zapped, error) {
	return &basicZapped{}, nil
}

func (z *basicZapped) Listen(ctx context.Context, port int) (io.ReadCloser, error) {
	var (
		lc  net.ListenConfig
		err error
	)
	z.listener, err = lc.Listen(ctx, "tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		err := z.listener.Close()
		if err != nil {
			zerolog.Ctx(ctx).Debug().Err(err).Msg("failed to close listener")
		}
	}()
	zerolog.Ctx(ctx).Info().Str("addr", z.listener.Addr().String()).Msg("listening for zaps")

	return z.listener.Accept()
}

func (z *basicZapped) Close() error {
	if z.listener != nil {
		return z.listener.Close()
	}
	return nil
}
