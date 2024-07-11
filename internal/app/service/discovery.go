package service

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/util"

	log "github.com/sirupsen/logrus"
)

type DiscoveryService struct {
	node             *Node
	context          context.Context
	refreshRate      time.Duration
	routingDiscovery *routing.RoutingDiscovery
}

func NewDiscoveryService(ctx context.Context, node *Node, refreshRate time.Duration) *DiscoveryService {
	routingDiscovery := routing.NewRoutingDiscovery(node.DHT)

	return &DiscoveryService{
		node:             node,
		context:          ctx,
		refreshRate:      refreshRate,
		routingDiscovery: routingDiscovery,
	}
}

// DiscoverPeers is a utility function that persistently discovers new peers
func (d *DiscoveryService) FindPeers() {
	ticker := time.NewTicker(d.refreshRate)
	defer ticker.Stop()

	for {
		select {
		case <-d.context.Done():
			return
		case <-ticker.C:
			peers, err := d.routingDiscovery.FindPeers(d.context, d.node.Topic)
			if err != nil {
				log.Warnf("Failed to find peers: %s", err)
				continue
			}

			for p := range peers {
				if p.ID == d.node.Host.ID() {
					log.Debug("Found self in discovery")
					continue
				}

				if d.node.Host.Network().Connectedness(p.ID) == network.Connected {
					log.Debugf("Already connected to peer %s", p.ID)
					continue
				}

				err := d.node.Host.Connect(d.context, p)
				if err != nil {
					log.Debugf("Failed to connect to peer %s: %v", p.ID, err)
				} else {
					log.Infof("Connected to new peer %s", p.ID)
				}
			}
		}
	}
}

// Advertise is a utility function that persistently advertises the service
func (d *DiscoveryService) Advertise() {
	util.Advertise(d.context, d.routingDiscovery, d.node.Topic)
}
