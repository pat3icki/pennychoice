package sflake

import (
	"errors"
	"sync"
	"time"
)

// State tracks the sequence and time to prevent collisions
type State struct {
	mu            sync.Mutex
	LastTimestamp int64
	Sequence      int64
}

// IDInfo holds the decrypted parts of a Snowflake ID
type IDInfo struct {
	Timestamp int64
	NodeID    int64
	Sequence  int64
}

const (
	NodeBits  = 10
	StepBits  = 12
	NodeMax   = -1 ^ (-1 << NodeBits)
	StepMax   = -1 ^ (-1 << StepBits)
	TimeShift = NodeBits + StepBits
	NodeShift = StepBits
)

var (
	DefaultEpoch = time.UnixMilli(1288834974657)
)

// GenerateID creates a 64-bit Snowflake ID
func GenerateID(epoch time.Time, nodeID int64, state *State) (int64, error) {
	if state == nil {
		state = &State{}
	}
	state.mu.Lock()
	defer state.mu.Unlock()

	if nodeID > NodeMax || nodeID < 0 {
		return 0, errors.New("node ID out of range")
	}

	now := time.Now().UnixMilli()

	if now < state.LastTimestamp {
		return 0, errors.New("clock moved backwards")
	}

	if now == state.LastTimestamp {
		state.Sequence = (state.Sequence + 1) & StepMax
		if state.Sequence == 0 {
			// Sequence exhausted, wait for next millisecond
			for now <= state.LastTimestamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		state.Sequence = 0
	}

	state.LastTimestamp = now

	// Bitwise assembly:
	// (Time since epoch) << 22 | (Node ID) << 12 | (Sequence)
	id := (now-epoch.UnixMilli())<<TimeShift | (nodeID << NodeShift) | (state.Sequence)

	return id, nil
}

// Describe breaks down an existing Snowflake ID into its components
func Describe(id int64, epoch time.Time) IDInfo {
	return IDInfo{
		Timestamp: (id >> TimeShift) + epoch.UnixMilli(),
		NodeID:    (id >> NodeShift) & NodeMax,
		Sequence:  id & StepMax,
	}
}
