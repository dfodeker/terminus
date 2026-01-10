package gid

import (
	"errors"
	"sync"
	"time"
)

const (
	// Epoch is January 1, 2024 00:00:00 UTC in milliseconds
	epoch int64 = 1704067200000

	// Bit allocations
	timestampBits = 41
	machineIDBits = 10
	sequenceBits  = 12

	// Max values
	maxMachineID = (1 << machineIDBits) - 1 // 1023
	maxSequence  = (1 << sequenceBits) - 1  // 4095

	// Shifts
	timestampShift = machineIDBits + sequenceBits
	machineIDShift = sequenceBits
)

// Generator generates Snowflake IDs
type Generator struct {
	mu        sync.Mutex
	machineID uint16
	sequence  uint16
	lastTime  int64
}

// NewGenerator creates a new Snowflake ID generator.
// machineID must be between 0 and 1023 (inclusive).
func NewGenerator(machineID uint16) (*Generator, error) {
	if machineID > maxMachineID {
		return nil, errors.New("machine ID must be between 0 and 1023")
	}
	return &Generator{
		machineID: machineID,
		sequence:  0,
		lastTime:  0,
	}, nil
}

// Generate creates a new unique 64-bit ID.
// Thread-safe.
func (g *Generator) Generate() uint64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UnixMilli() - epoch

	if now == g.lastTime {
		g.sequence = (g.sequence + 1) & maxSequence
		if g.sequence == 0 {
			// Sequence overflow, wait for next millisecond
			for now <= g.lastTime {
				now = time.Now().UnixMilli() - epoch
			}
		}
	} else {
		g.sequence = 0
	}

	g.lastTime = now

	return uint64(now)<<timestampShift |
		uint64(g.machineID)<<machineIDShift |
		uint64(g.sequence)
}

// ExtractTime extracts the timestamp from a Snowflake ID
func ExtractTime(id uint64) time.Time {
	ms := int64(id>>timestampShift) + epoch
	return time.UnixMilli(ms)
}

// ExtractMachineID extracts the machine ID from a Snowflake ID
func ExtractMachineID(id uint64) uint16 {
	return uint16((id >> machineIDShift) & maxMachineID)
}

// ExtractSequence extracts the sequence number from a Snowflake ID
func ExtractSequence(id uint64) uint16 {
	return uint16(id & maxSequence)
}
