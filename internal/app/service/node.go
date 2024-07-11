package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Node struct {
	Host  host.Host
	DHT   *dht.IpfsDHT
	Topic string
}

func NewNode(ctx context.Context, topic string, priv crypto.PrivKey, port int) (*Node, error) {
	host, err := newHost(priv, port)
	if err != nil {
		return nil, err
	}

	dht, err := newDHT(ctx, host)
	if err != nil {
		return nil, err
	}

	err = bootstrapPeers(ctx, host)
	if err != nil {
		return nil, err
	}

	return &Node{
		Host:  host,
		DHT:   dht,
		Topic: topic,
	}, nil
}

// newHost creates a new libp2p host
func newHost(priv crypto.PrivKey, port int) (host.Host, error) {
	return libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port),
			fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic", port),
		),
		libp2p.DefaultTransports,
		libp2p.NATPortMap(),
		libp2p.DefaultSecurity,
		libp2p.DefaultMuxers,
		libp2p.DefaultPeerstore,
	)
}

// bootstrapPeers connects to the default bootstrap peers
func bootstrapPeers(ctx context.Context, host host.Host) error {
	var wg sync.WaitGroup
	var connected bool
	for _, peerAddr := range dht.DefaultBootstrapPeers {
		peerinfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
		if err != nil {
			log.Warnf("Invalid peer info: %s", err)
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := host.Connect(ctx, *peerinfo); err != nil {
				log.Debug("Bootstrap warning: ", err)
			} else {
				log.Infof("Connected to bootstrap peer: %s", peerinfo.ID)
				connected = true
			}
		}()
		// Rate limit bootstrapping to avoid overwhelming the network
		time.Sleep(1 * time.Second)
	}
	wg.Wait()
	if !connected {
		return fmt.Errorf("failed to connect to any bootstrap peers")
	}
	return nil
}

// Start a DHT, for use in peer discovery. We can't just make a new DHT
// client because we want each peer to maintain its own local copy of the
// DHT, so that the bootstrapping node of the DHT can go down without
// inhibiting future peer discovery.
func newDHT(ctx context.Context, host host.Host) (*dht.IpfsDHT, error) {
	kademliaDHT, err := dht.New(ctx, host)
	if err != nil {
		return nil, err
	}

	if err := kademliaDHT.Bootstrap(ctx); err != nil {
		return nil, err
	}

	return kademliaDHT, nil
}

func (n *Node) Close() {
	if err := n.DHT.Close(); err != nil {
		log.Warnf("Failed to close DHT: %v", err)
	}

	if err := n.Host.Close(); err != nil {
		log.Warnf("Failed to close host: %v", err)
	}
}
