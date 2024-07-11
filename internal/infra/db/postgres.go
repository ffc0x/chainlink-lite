package db

// PG implementation of PriceMessageRepository interface

import (
	"context"
	"fmt"

	"time"

	"github.com/jackc/pgx/v4"

	"chainlink-lite/internal/app/domain"

	log "github.com/sirupsen/logrus"
)

type PgPriceMessageRepository struct {
	db *pgx.Conn
}

var _ domain.PriceMessageRepository = (*PgPriceMessageRepository)(nil)

func NewPriceMessageRepository(ctx context.Context, connString string) (*PgPriceMessageRepository, error) {
	config, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %v", err)
	}

	db, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	return &PgPriceMessageRepository{db: db}, nil
}

// Store the priceMsg if at least minInterval has passed since the last write
// Timestamp is used to check the last write time
// Returns true if the message was stored, false if it was skipped
func (conn *PgPriceMessageRepository) StorePriceIfAllowed(ctx context.Context, priceMsg *domain.PriceMessage, minInterval time.Duration) (bool, error) {
	tx, err := conn.db.Begin(ctx)
	if err != nil {
		log.Debugf("Failed to start transaction: %v", err)
		return false, err
	}
	defer tx.Rollback(ctx) //nolint:all

	lockID := int64(1)
	_, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", lockID)
	if err != nil {
		log.Warnf("Failed to acquire advisory lock: %v", err)
		return false, err
	}

	var lastTimestamp time.Time
	err = tx.QueryRow(ctx, "SELECT timestamp FROM eth_price_messages ORDER BY timestamp DESC LIMIT 1").Scan(&lastTimestamp)
	if err != nil {
		if err == pgx.ErrNoRows {
			// No rows found, use Unix epoch as the default value
			lastTimestamp = time.Unix(0, 0)
		} else {
			log.Debugf("Failed to get latest timestamp: %v", err)
			return false, err
		}
	}

	if time.Since(lastTimestamp) >= minInterval {
		query := "INSERT INTO eth_price_messages (message_id, price, publisher, writer, signers, signatures, created_at, timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT DO NOTHING"
		_, err = tx.Exec(ctx, query, priceMsg.MessageID, priceMsg.Price, priceMsg.Publisher, priceMsg.Writer, priceMsg.Signers, priceMsg.Signatures, time.Unix(priceMsg.CreatedAt, 0), time.Now())
		if err != nil {
			log.Debugf("Failed to store ETH price in the database: %v", err)
			return false, err
		}
	} else {
		log.Debug("Not enough time has passed since the last message")
		return false, nil
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Debugf("Failed to commit transaction: %v", err)
		return false, err
	}

	return true, nil
}

func (r *PgPriceMessageRepository) Close(ctx context.Context) error {
	r.db.Close(ctx)
	return nil
}
