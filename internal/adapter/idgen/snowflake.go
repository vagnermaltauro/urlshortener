package idgen

import (
	"errors"
	"sync"
	"time"

	"urlshortner/internal/domain/repository"
)

const (
	// Epoch is the custom epoch (2024-01-01 00:00:00 UTC) in milliseconds
	// This gives us 139 years of IDs from 2024 to 2163
	epoch int64 = 1704067200000

	// Bit allocation for Snowflake ID (64 bits total):
	// [timestamp:42][machineID:10][sequence:12]
	timestampBits  = 42 // ~139 years of timestamps
	machineIDBits  = 10 // 1024 unique machines
	sequenceBits   = 12 // 4096 IDs per millisecond per machine

	// Maximum values
	maxMachineID = (1 << machineIDBits) - 1 // 1023
	maxSequence  = (1 << sequenceBits) - 1  // 4095
)

// SnowflakeGenerator generates unique 64-bit IDs using the Snowflake algorithm
// It's thread-safe and can generate up to 4,096,000 IDs per second per machine
type SnowflakeGenerator struct {
	machineID     uint16
	sequence      uint32
	lastTimestamp int64
	mutex         sync.Mutex
}

// NewSnowflakeGenerator creates a new Snowflake ID generator
// machineID must be unique across all machines in the cluster (0-1023)
func NewSnowflakeGenerator(machineID uint16) (repository.IDGenerator, error) {
	if machineID > maxMachineID {
		return nil, errors.New("machine ID exceeds maximum value of 1023")
	}

	return &SnowflakeGenerator{
		machineID:     machineID,
		sequence:      0,
		lastTimestamp: 0,
	}, nil
}

// Generate creates a new unique 64-bit ID
// Format: [timestamp:42 bits][machineID:10 bits][sequence:12 bits]
// This provides:
// - Uniqueness across machines (via machineID)
// - Uniqueness in time (via timestamp)
// - Uniqueness within same millisecond (via sequence)
func (g *SnowflakeGenerator) Generate() (int64, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Get current timestamp in milliseconds since epoch
	timestamp := time.Now().UnixMilli() - epoch

	// Clock moved backwards - this should never happen in production
	if timestamp < g.lastTimestamp {
		return 0, errors.New("clock moved backwards, refusing to generate ID")
	}

	if timestamp == g.lastTimestamp {
		// Same millisecond - increment sequence
		g.sequence = (g.sequence + 1) & maxSequence

		if g.sequence == 0 {
			// Sequence exhausted for this millisecond - wait for next millisecond
			for timestamp <= g.lastTimestamp {
				timestamp = time.Now().UnixMilli() - epoch
			}
		}
	} else {
		// New millisecond - reset sequence to 0
		g.sequence = 0
	}

	g.lastTimestamp = timestamp

	// Compose the 64-bit ID:
	// - Left shift timestamp by (machineIDBits + sequenceBits) = 22 bits
	// - Left shift machineID by sequenceBits = 12 bits
	// - OR with sequence (no shift needed)
	id := (timestamp << (machineIDBits + sequenceBits)) |
		(int64(g.machineID) << sequenceBits) |
		int64(g.sequence)

	return id, nil
}
