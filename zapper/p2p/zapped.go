package p2p

import (
	"context"
	"fmt"
	"io"

	"github.com/hinshun/zap/zapper"
	"github.com/libp2p/go-libp2p"
	net "github.com/libp2p/go-libp2p-core/network"
	host "github.com/libp2p/go-libp2p-host"
	tcp "github.com/libp2p/go-tcp-transport"
	"github.com/rs/zerolog"
)

type p2pZapped struct {
	host host.Host
}

func NewZapped() (zapper.Zapped, error) {
	return &p2pZapped{}, nil
}

func (z *p2pZapped) Listen(ctx context.Context, port int) (io.ReadCloser, error) {
	var err error
	z.host, err = libp2p.New(ctx,
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)),
		libp2p.Transport(tcp.NewTCPTransport),
	)
	if err != nil {
		return nil, err
	}

	firstByte := make(chan struct{})
	r, w := io.Pipe()
	z.host.SetStreamHandler("/zap", func(s net.Stream) {
		close(firstByte)
		_, err := io.Copy(w, s)
		if err != nil {
			w.CloseWithError(err)
			return
		}
	})

	var addrs []string
	for _, addr := range z.host.Addrs() {
		addrs = append(addrs, fmt.Sprintf("%s/p2p/%s", addr.String(), z.host.ID()))
	}

	zerolog.Ctx(ctx).Info().Strs("addrs", addrs).Msg("listening for zaps")
	<-firstByte
	return r, nil
}

func (z *p2pZapped) Close() error {
	if z.host != nil {
		return z.host.Close()
	}
	return nil
}
