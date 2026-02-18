package sflake

import (
	"sync"
	"testing"
)

func TestGenerateID_Basic(t *testing.T) {
	state := &State{}
	nodeID := int64(10)
	epoch := DefaultEpoch

	id, err := GenerateID(epoch, nodeID, state)
	if err != nil {
		t.Fatalf("Failed to generate ID: %v", err)
	}

	if id <= 0 {
		t.Errorf("Expected positive ID, got %d", id)
	}

	// Verify the components match using Describe
	info := Describe(id, epoch)
	if info.NodeID != nodeID {
		t.Errorf("Expected NodeID %d, got %d", nodeID, info.NodeID)
	}
}

func TestGenerateID_SequenceIncrement(t *testing.T) {
	state := &State{}
	nodeID := int64(1)
	epoch := DefaultEpoch

	// Generate two IDs in the same millisecond (likely)
	id1, _ := GenerateID(epoch, nodeID, state)
	id2, _ := GenerateID(epoch, nodeID, state)

	info1 := Describe(id1, epoch)
	info2 := Describe(id2, epoch)

	if info1.Timestamp == info2.Timestamp {
		if info2.Sequence != info1.Sequence+1 {
			t.Errorf("Sequence did not increment. Got %d and %d", info1.Sequence, info2.Sequence)
		}
	}
}

func TestGenerateID_NodeIDRange(t *testing.T) {
	state := &State{}

	// Test upper bound
	_, err := GenerateID(DefaultEpoch, NodeMax+1, state)
	if err == nil {
		t.Error("Expected error for NodeID exceeding NodeMax, got nil")
	}

	// Test lower bound
	_, err = GenerateID(DefaultEpoch, -1, state)
	if err == nil {
		t.Error("Expected error for negative NodeID, got nil")
	}
}

func TestGenerateID_Concurrency(t *testing.T) {
	state := &State{}
	nodeID := int64(42)
	numGoroutines := 10
	idsPerGoroutine := 1000

	var wg sync.WaitGroup
	idChan := make(chan int64, numGoroutines*idsPerGoroutine)

	// Launch multiple goroutines generating IDs simultaneously
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				id, err := GenerateID(DefaultEpoch, nodeID, state)
				if err == nil {
					idChan <- id
				}
			}
		}()
	}

	wg.Wait()
	close(idChan)

	// Check for duplicates
	uniqueIDs := make(map[int64]bool)
	for id := range idChan {
		if uniqueIDs[id] {
			t.Errorf("Duplicate ID detected: %d", id)
		}
		uniqueIDs[id] = true
	}

	expectedCount := numGoroutines * idsPerGoroutine
	if len(uniqueIDs) != expectedCount {
		t.Errorf("Expected %d unique IDs, but only got %d", expectedCount, len(uniqueIDs))
	}
}
