package peer

import (
	"fmt"

	ma "github.com/multiformats/go-multiaddr"
)

// AddrInfo is a small struct used to pass around a peer with
// a set of addresses (and later, keys?).
type AddrInfo struct {
	ID    ID
	Addrs []ma.Multiaddr
}

var _ fmt.Stringer = AddrInfo{}

func (pi AddrInfo) String() string {
	return fmt.Sprintf("{%v: %v}", pi.ID, pi.Addrs)
}

var ErrInvalidAddr = fmt.Errorf("invalid p2p multiaddr")

// AddrInfosFromP2pAddrs converts a set of Multiaddrs to a set of AddrInfos.
func AddrInfosFromP2pAddrs(maddrs ...ma.Multiaddr) ([]AddrInfo, error) {
	m := make(map[ID][]ma.Multiaddr)
	for _, maddr := range maddrs {
		transport, id := SplitAddr(maddr)
		if id == "" {
			return nil, ErrInvalidAddr
		}
		if transport == nil {
			if _, ok := m[id]; !ok {
				m[id] = nil
			}
		} else {
			m[id] = append(m[id], transport)
		}
	}
	ais := make([]AddrInfo, 0, len(m))
	for id, maddrs := range m {
		ais = append(ais, AddrInfo{ID: id, Addrs: maddrs})
	}
	return ais, nil
}

// SplitAddr splits a p2p Multiaddr into a transport multiaddr and a peer ID.
//
// * Returns a nil transport if the address only contains a /p2p part.
// * Returns an empty peer ID if the address doesn't contain a /p2p part.
func SplitAddr(m ma.Multiaddr) (transport ma.Multiaddr, id ID) {
	if m == nil {
		return nil, ""
	}

	transport, p2ppart := ma.SplitLast(m)
	if p2ppart == nil || p2ppart.Protocol().Code != ma.P_P2P {
		return m, ""
	}
	id = ID(p2ppart.RawValue()) // already validated by the multiaddr library.
	return transport, id
}

// IDFromP2PAddr extracts the peer ID from a p2p Multiaddr
func IDFromP2PAddr(m ma.Multiaddr) (ID, error) {
	if m == nil {
		return "", ErrInvalidAddr
	}
	var lastComponent ma.Component
	ma.ForEach(m, func(c ma.Component) bool {
		lastComponent = c
		return true
	})
	if lastComponent.Protocol().Code != ma.P_P2P {
		return "", ErrInvalidAddr
	}

	id := ID(lastComponent.RawValue()) // already validated by the multiaddr library.
	return id, nil
}

// AddrInfoFromString builds an AddrInfo from the string representation of a Multiaddr
func AddrInfoFromString(s string) (*AddrInfo, error) {
	a, err := ma.NewMultiaddr(s)
	if err != nil {
		return nil, err
	}

	return AddrInfoFromP2pAddr(a)
}

// AddrInfoFromP2pAddr converts a Multiaddr to an AddrInfo.
func AddrInfoFromP2pAddr(m ma.Multiaddr) (*AddrInfo, error) {
	transport, id := SplitAddr(m)
	if id == "" {
		return nil, ErrInvalidAddr
	}
	info := &AddrInfo{ID: id}
	if transport != nil {
		info.Addrs = []ma.Multiaddr{transport}
	}
	return info, nil
}

// AddrInfoToP2pAddrs converts an AddrInfo to a list of Multiaddrs.
func AddrInfoToP2pAddrs(pi *AddrInfo) ([]ma.Multiaddr, error) {
	p2ppart, err := ma.NewComponent("p2p", pi.ID.String())
	if err != nil {
		return nil, err
	}
	if len(pi.Addrs) == 0 {
		return []ma.Multiaddr{p2ppart.Multiaddr()}, nil
	}
	addrs := make([]ma.Multiaddr, 0, len(pi.Addrs))
	for _, addr := range pi.Addrs {
		addrs = append(addrs, addr.Encapsulate(p2ppart))
	}
	return addrs, nil
}

func (pi *AddrInfo) Loggable() map[string]interface{} {
	return map[string]interface{}{
		"peerID": pi.ID.String(),
		"addrs":  pi.Addrs,
	}
}

// AddrInfosToIDs extracts the peer IDs from the passed AddrInfos and returns them in-order.
func AddrInfosToIDs(pis []AddrInfo) []ID {
	ps := make([]ID, len(pis))
	for i, pi := range pis {
		ps[i] = pi.ID
	}
	return ps
}
