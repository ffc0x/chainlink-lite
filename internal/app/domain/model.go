package domain

import (
	"fmt"
)

// PriceMessage represents the message that will be published to the pubsub topic
// MessageID is the unique identifier of the message
// Price is the price of the cryptocurrency
// Publisher is the node ID of the original publisher
// Writer is the node ID of node that persisted the message
// Signers are the node IDs of the nodes that signed the message
// Signatures are the signatures of the message
// CreatedAt is the timestamp when the message was originaly created
// Timestamp is the timestamp when the message was persisted
type PriceMessage struct {
	MessageID  string   `json:"message_id" validate:"required"`
	Price      string   `json:"price" validate:"required,numeric"`
	Publisher  string   `json:"publisher" validate:"required"`
	Writer     string   `json:"-"`
	Signers    []string `json:"signers" validate:"required,min=1"`
	Signatures []string `json:"signatures" validate:"required,min=1"`
	CreatedAt  int64    `json:"timestamp" validate:"required"`
	Timestamp  int64    `json:"-"`
}

func (p PriceMessage) String() string {
	return fmt.Sprintf("{MessageID: %s, Price: %s, Publisher: %s, Writer: %s, Signers: %v, Signatures: %v, CreatedAt: %d, Timestamp: %d}",
		p.MessageID, p.Price, p.Publisher, p.Writer, p.Signers, p.Signatures, p.CreatedAt, p.Timestamp)
}
