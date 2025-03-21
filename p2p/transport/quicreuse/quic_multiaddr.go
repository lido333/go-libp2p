package quicreuse

import (
	"errors"
	"net"

	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/quic-go/quic-go"
)

var (
	quicV1MA = ma.StringCast("/quic-v1")
)

func ToQuicMultiaddr(na net.Addr, version quic.Version) (ma.Multiaddr, error) {
	udpMA, err := manet.FromNetAddr(na)
	if err != nil {
		return nil, err
	}
	switch version {
	case quic.Version1:
		return udpMA.Encapsulate(quicV1MA), nil
	default:
		return nil, errors.New("unknown QUIC version")
	}
}

func FromQuicMultiaddr(addr ma.Multiaddr) (*net.UDPAddr, quic.Version, error) {
	var version quic.Version
	partsBeforeQUIC := make([]ma.Component, 0, 2)
loop:
	for _, c := range addr {
		switch c.Protocol().Code {
		case ma.P_QUIC_V1:
			version = quic.Version1
			break loop
		default:
			partsBeforeQUIC = append(partsBeforeQUIC, c)
		}
	}
	if len(partsBeforeQUIC) == 0 {
		return nil, version, errors.New("no addr before QUIC component")
	}
	if version == 0 {
		// Not found
		return nil, version, errors.New("unknown QUIC version")
	}
	netAddr, err := manet.ToNetAddr(partsBeforeQUIC)
	if err != nil {
		return nil, version, err
	}
	udpAddr, ok := netAddr.(*net.UDPAddr)
	if !ok {
		return nil, 0, errors.New("not a *net.UDPAddr")
	}
	return udpAddr, version, nil
}
