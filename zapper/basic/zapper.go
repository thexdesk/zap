package basic

import (
	"context"
	"io"
	"net"
	"os"

	"github.com/hinshun/zap/zapper"
	"github.com/rs/zerolog"
)

type basicZapper struct {
}

func NewZapper() (zapper.Zapper, error) {
	return &basicZapper{}, nil
}

func (z *basicZapper) Zap(ctx context.Context, src, dst string) error {
	s, err := net.Dial("tcp", dst)
	if err != nil {
		return err
	}
	defer s.Close()

	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	n, err := io.Copy(s, f)
	if err != nil {
		return err
	}

	zerolog.Ctx(ctx).Debug().Int64("bytes", n).Msg("zapped file")
	return nil
}
