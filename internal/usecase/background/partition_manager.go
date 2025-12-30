package background

import (
	"context"
	"database/sql"
	"log"
	"time"
)

// PartitionManager is a background job that auto-creates next month's partition
// This ensures the partitioned table is always ready for new data
type PartitionManager struct {
	db       *sql.DB
	interval time.Duration
}

// NewPartitionManager creates a new partition manager background job
func NewPartitionManager(db *sql.DB, interval time.Duration) *PartitionManager {
	return &PartitionManager{
		db:       db,
		interval: interval,
	}
}

// Start begins the background partition creation loop
// This should be called as a goroutine: go partitionMgr.Start(ctx)
func (m *PartitionManager) Start(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	log.Printf("[PartitionManager] Started with interval %v", m.interval)

	// Create next partition immediately on startup
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

// createNextPartition calls the PostgreSQL function to create next month's partition
func (m *PartitionManager) createNextPartition(ctx context.Context) {
	_, err := m.db.ExecContext(ctx, "SELECT create_next_partition()")
	if err != nil {
		log.Printf("[PartitionManager] Error creating next partition: %v", err)
		return
	}

	log.Println("[PartitionManager] Successfully verified/created next month's partition")
}
