package idgen

import (
	"errors"
	"sync"
	"time"

	"urlshortner/internal/domain/repository"
)

const (
	epoch int64 = 1704067200000

	timestampBits = 42
	machineIDBits = 10
	sequenceBits  = 12

	maxMachineID = (1 << machineIDBits) - 1
	maxSequence  = (1 << sequenceBits) - 1
)

type SnowflakeGenerator struct {
	machineID     uint16
	sequence      uint32
	lastTimestamp int64
	mutex         sync.Mutex
}

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

func (g *SnowflakeGenerator) Generate() (int64, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	timestamp := time.Now().UnixMilli() - epoch

	if timestamp < g.lastTimestamp {
		return 0, errors.New("clock moved backwards, refusing to generate ID")
	}

	if timestamp == g.lastTimestamp {

		g.sequence = (g.sequence + 1) & maxSequence

		if g.sequence == 0 {

			for timestamp <= g.lastTimestamp {
				timestamp = time.Now().UnixMilli() - epoch
			}
		}
	} else {

		g.sequence = 0
	}

	g.lastTimestamp = timestamp

	id := (timestamp << (machineIDBits + sequenceBits)) |
		(int64(g.machineID) << sequenceBits) |
		int64(g.sequence)

	return id, nil
}
