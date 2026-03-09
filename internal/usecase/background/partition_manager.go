package background

import (
	"context"
	"database/sql"
	"log"
	"time"
)

type PartitionManager struct {
	db       *sql.DB
	interval time.Duration
}

func NewPartitionManager(db *sql.DB, interval time.Duration) *PartitionManager {
	return &PartitionManager{
		db:       db,
		interval: interval,
	}
}

func (m *PartitionManager) Start(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	log.Printf("[PartitionManager] Started with interval %v", m.interval)

	m.createNextPartition(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("[PartitionManager] Stopped")
			return

		case <-ticker.C:
			m.createNextPartition(ctx)
		}
	}
}

func (m *PartitionManager) createNextPartition(ctx context.Context) {
	_, err := m.db.ExecContext(ctx, "SELECT create_next_partition()")
	if err != nil {
		log.Printf("[PartitionManager] Error creating next partition: %v", err)
		return
	}

	log.Println("[PartitionManager] Successfully verified/created next month's partition")
}
