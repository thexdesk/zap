package p2p

import (
	"context"
	"io"
	"os"

	"github.com/hinshun/zap/zapper"
	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	host "github.com/libp2p/go-libp2p-host"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	tcp "github.com/libp2p/go-tcp-transport"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/rs/zerolog"
)

type p2pZapper struct {
	host host.Host
}

func NewZapper() (zapper.Zapper, error) {
	return &p2pZapper{}, nil
}

func (z *p2pZapper) Zap(ctx context.Context, src, dst string) error {
	var err error
	z.host, err = libp2p.New(ctx,
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.Transport(tcp.NewTCPTransport),
	)
	if err != nil {
		return err
	}

	p2pAddr, err := multiaddr.NewMultiaddr(dst)
	if err != nil {
		return err
	}

	addrInfo, err := peer.AddrInfoFromP2pAddr(p2pAddr)
	if err != nil {
		return err
	}

	z.host.Peerstore().AddAddrs(addrInfo.ID, addrInfo.Addrs, peerstore.PermanentAddrTTL)
	s, err := z.host.NewStream(ctx, addrInfo.ID, "/zap")
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
