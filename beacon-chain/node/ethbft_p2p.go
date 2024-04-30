package node

import (
	"context"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/urfave/cli/v2"
)

const DiscoveryServiceTag = "im-ethbft-pubsub"

type P2PNode struct {
	Host host.Host
	//Pubsub *pubsub.PubSub
	// Used to track new peer conns
	//PeerChan chan peer.AddrInfo
	// Used to receive events
	Subscription *pubsub.Subscription
	// Used to publish
	Topic *pubsub.Topic
}

func initDHT(ctx context.Context, h host.Host) *dht.IpfsDHT {
	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	kademliaDHT, err := dht.New(ctx, h)
	if err != nil {
		panic(err)
	}
	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	for _, peerAddr := range dht.DefaultBootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := h.Connect(ctx, *peerinfo); err != nil {
				fmt.Println("Bootstrap warning:", err)
			}
		}()
	}
	wg.Wait()

	return kademliaDHT
}

func discoverPeers(ctx context.Context, h host.Host) {
	kademliaDHT := initDHT(ctx, h)
	routingDiscovery := drouting.NewRoutingDiscovery(kademliaDHT)
	dutil.Advertise(ctx, routingDiscovery, DiscoveryServiceTag)

	// Look for others who have announced and attempt to connect to them
	anyConnected := false
	for !anyConnected {
		fmt.Println("Searching for peers...")
		peerChan, err := routingDiscovery.FindPeers(ctx, DiscoveryServiceTag)
		if err != nil {
			panic(err)
		}
		for peer := range peerChan {
			if peer.ID == h.ID() {
				continue // No self connection
			}
			err := h.Connect(ctx, peer)
			if err != nil {
				fmt.Printf("Failed connecting to %s, error: %s\n", peer.ID, err)
			} else {
				fmt.Println("Connected to:", peer.ID)
				anyConnected = true
			}
		}
	}
	fmt.Println("Peer discovery complete")
}

func startP2P(ctx context.Context, cliCtx *cli.Context, isSequencer bool) (*P2PNode, error) {
	// host and non-host listen to different ports
	var h host.Host
	var err error

	h, err = libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
	)

	go discoverPeers(ctx, h)

	if err != nil {
		return nil, err
	}
	fmt.Println("Node is running with ID:", h.ID().String())
	fmt.Println("Listen addresses:", h.Addrs())

	// Create a new GossipSub pubsub service
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		fmt.Println("Failed to create pubsub: ", err)
		return nil, err
	}

	//fmt.Println("Setting up Peer discovery using MDNS (local only)")
	n := &P2PNode{}
	n.Host = h

	topic, err := ps.Join(DiscoveryServiceTag)
	if err != nil {
		return nil, err
	}
	n.Topic = topic
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}
	n.Subscription = sub

	return n, nil
}
