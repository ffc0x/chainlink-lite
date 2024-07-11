package service

import (
	"chainlink-lite/internal/app/domain"
	"context"
	"encoding/json"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
)

type PubSubService struct {
	gossip    *pubsub.PubSub
	topic     *pubsub.Topic
	sub       *pubsub.Subscription
	ctx       context.Context
	host      host.Host
	topicName string
}

func NewPubSubService(ctx context.Context, topicName string, host host.Host) (*PubSubService, error) {
	gossip, err := pubsub.NewGossipSub(ctx, host,
		pubsub.WithMessageSignaturePolicy(pubsub.StrictSign),
		pubsub.WithStrictSignatureVerification(true),
		pubsub.WithPeerExchange(true),
		pubsub.WithMessageSigning(true))
	if err != nil {
		return nil, err
	}

	topic, err := gossip.Join(topicName)
	if err != nil {
		return nil, err
	}

	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	return &PubSubService{
		ctx:       ctx,
		gossip:    gossip,
		topic:     topic,
		sub:       sub,
		host:      host,
		topicName: topicName,
	}, nil
}

// Publish publishes a message to the pubsub topic
func (p *PubSubService) Publish(msg *domain.PriceMessage) error {
	// Publish message to topic
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return p.topic.Publish(p.ctx, data)
}

// Receive receives a message from the pubsub topic
// If skipFromSelf is true, messages from the current node are skipped
func (p *PubSubService) Receive(skipFromSelf bool) (*domain.PriceMessage, error) {
	// Receive message from topic
	msg, err := p.sub.Next(p.ctx)
	if err != nil {
		return nil, err
	}

	// Skip messages from the current node
	if (msg.ReceivedFrom == p.host.ID() || msg.GetFrom() == p.host.ID()) && skipFromSelf {
		log.Debug("Skipping message from self")
		return nil, nil
	}

	var priceMsg domain.PriceMessage
	err = json.Unmarshal(msg.Data, &priceMsg)
	if err != nil {
		return nil, err
	}

	// Validate message
	err = ValidateMessage(priceMsg)
	if err != nil {
		log.Info("Invalid message received: ", priceMsg, err)
		return nil, err
	}

	return &priceMsg, nil
}

func (p *PubSubService) Close() {
	p.sub.Cancel()
}

func (p *PubSubService) GetTopicName() string {
	return p.topicName
}

func (p *PubSubService) GetNodeID() string {
	return p.host.ID().String()
}

func ValidateMessage(priceMsg domain.PriceMessage) error {
	validate := validator.New()
	return validate.Struct(priceMsg)
}
