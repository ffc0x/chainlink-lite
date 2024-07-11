package usecase

import (
	"chainlink-lite/internal/app/domain"
	"chainlink-lite/internal/app/service"
	"context"
	"time"

	log "github.com/sirupsen/logrus"
)

type Subscriber struct {
	pubsub        *service.PubSubService
	repo          domain.PriceMessageRepository
	minSignatures int
	minInterval   time.Duration
	signer        *service.SignerService
}

func NewSubscriber(pubsub *service.PubSubService, repo domain.PriceMessageRepository, minSignatures int,
	minInterval time.Duration, signer *service.SignerService) *Subscriber {
	return &Subscriber{
		pubsub:        pubsub,
		repo:          repo,
		minSignatures: minSignatures,
		minInterval:   minInterval,
		signer:        signer,
	}
}

func (s *Subscriber) Start(ctx context.Context) {
	// Start a ticker to check for new messages every second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			msg, err := s.pubsub.Receive(true)
			if err != nil {
				log.Warnf("Failed to receive message: %v", err)
				continue
			}
			if msg == nil {
				// Message from self, skip.
				continue
			}

			log.Info("Received message: ", msg)

			if len(msg.Signatures) >= s.minSignatures {
				msg.Writer = s.pubsub.GetNodeID()
				// Store the message if at least s.minInterval has passed since the last write
				_, err := s.repo.StorePriceIfAllowed(ctx, msg, s.minInterval)
				if err != nil {
					log.Warnf("Failed to store message: %v", err)
					continue
				}
			} else {
				// Check if the message has already been signed by the current node
				// If not, sign the message and republish it
				if !s.AlreadySigned(*msg) {
					signedMsg, err := s.signer.SignMessage(msg.Price)
					if err != nil {
						log.Warnf("Failed to sign message: %v", err)
						continue
					}
					msg.Signatures = append(msg.Signatures, signedMsg)
					msg.Signers = append(msg.Signers, s.pubsub.GetNodeID())

					// Republish the message
					if err := s.pubsub.Publish(msg); err != nil {
						log.Printf("Failed to re-publish message: %v", err)
					}
				}
			}
		}
	}
}

// Check if the message has already been signed by the current node
func (s *Subscriber) AlreadySigned(priceMsg domain.PriceMessage) bool {
	for _, sig := range priceMsg.Signatures {
		signedMsg, err := s.signer.SignMessage(priceMsg.Price)
		if err != nil {
			log.Warnf("Failed to sign message: %v", err)
			continue
		}

		if sig == signedMsg {
			log.Debug("Skipping message already signed by self")
			return true
		}
	}
	return false
}
