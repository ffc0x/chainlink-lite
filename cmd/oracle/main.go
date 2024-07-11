package main

import (
	"chainlink-lite/config"
	"context"
	"os/signal"
	"syscall"
	"time"

	"chainlink-lite/internal/app/domain"
	"chainlink-lite/internal/app/service"
	"chainlink-lite/internal/app/usecase"
	"chainlink-lite/internal/infra/db"
	"chainlink-lite/internal/infra/eth"

	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/rand"
)

func main() {
	// Sleep for a random number of seconds to avoid multiple nodes starting at the same time
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)

	// Create a context that is canceled when a signal is received
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Unable to load configuration: %v", err)
	}

	log.SetLevel(log.Level(cfg.LogLevel))

	// Create a the database repository
	repo, err := db.NewPriceMessageRepository(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatalf("Unable to create repository: %v", err)
	}
	defer repo.Close(ctx)

	signer, err := service.NewSignerService()
	if err != nil {
		log.Fatalf("Unable to create signer service: %v", err)
	}

	// Create a node and discovery service
	node, err := service.NewNode(ctx, cfg.PubSub.TopicName, signer.GetPrivateKey(), cfg.PubSub.Port)
	if err != nil {
		log.Fatalf("Unable to create node: %v", err)
	}
	defer node.Close()

	log.Info("Node created: ", node.Host.ID())

	discovery := service.NewDiscoveryService(ctx, node, cfg.PubSub.DiscoverPeersInterval)
	// Start the discovery service
	go discovery.FindPeers()
	// Advertise the node
	discovery.Advertise()

	// Create a pubsub service
	pubsub, err := service.NewPubSubService(ctx, cfg.PubSub.TopicName, node.Host)
	if err != nil {
		log.Fatalf("Unable to create pubsub service: %v", err)
	}
	defer pubsub.Close()

	var priceTicker domain.EthPriceTicker
	if cfg.PriceTicker.Mock {
		priceTicker = eth.NewMockTicker()
	} else {
		priceTicker = eth.NewEthPriceTicker(cfg.PriceTicker.URL)
	}

	// Create a publisher and subscriber
	publisher := usecase.NewPublisher(priceTicker, cfg.PubSub.FetchPriceInterval, pubsub, signer)
	subscriber := usecase.NewSubscriber(pubsub, repo, cfg.PubSub.MinSignaturesToWrite, cfg.PubSub.MinIntervalBetweenWrites, signer)

	// Start the publisher and subscriber
	go publisher.Start(ctx)
	go subscriber.Start(ctx)

	<-ctx.Done()
}
