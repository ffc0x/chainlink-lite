package usecase

import (
	"chainlink-lite/internal/app/domain"
	"chainlink-lite/internal/app/service"
	"chainlink-lite/internal/util"
	"context"
	"time"

	log "github.com/sirupsen/logrus"
)

type Publisher struct {
	ethClient domain.EthPriceTicker
	interval  time.Duration
	pubsub    *service.PubSubService
	signer    *service.SignerService
}

func NewPublisher(ethClient domain.EthPriceTicker, interval time.Duration, pubsub *service.PubSubService, signer *service.SignerService) *Publisher {
	return &Publisher{
		ethClient: ethClient,
		interval:  interval,
		pubsub:    pubsub,
		signer:    signer,
	}
}

func (p *Publisher) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Info("Fetching ETH price")
			price, err := p.ethClient.FetchPrice()
			if err != nil {
				log.Warnf("Failed to fetch ETH price: %v", err)
				continue
			}
			log.Info("ETH price fetched: ", price)

			signature, err := p.signer.SignMessage(price)
			if err != nil {
				log.Warnf("Failed to sign message: %v", err)
				continue
			}

			id, err := util.GenerateUUID()
			if err != nil {
				log.Warnf("Failed to generate UUID: %v", err)
				continue
			}
			priceMsg := domain.PriceMessage{
				MessageID:  id,
				Price:      price,
				Publisher:  p.pubsub.GetNodeID(),
				Signers:    []string{p.pubsub.GetNodeID()},
				Signatures: []string{signature},
				CreatedAt:  time.Now().Unix(),
			}

			if err := p.pubsub.Publish(&priceMsg); err != nil {
				log.Warnf("Failed to publish price message: %v", err)
				continue
			}
			log.Info("Price message published: ", priceMsg)
		}
	}
}
