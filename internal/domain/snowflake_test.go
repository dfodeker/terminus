package domain

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewGenerator(t *testing.T) {
	tests := []struct {
		name      string
		machineID uint16
		wantErr   bool
	}{
		{
			name:      "valid machine ID 0",
			machineID: 0,
			wantErr:   false,
		},
		{
			name:      "valid machine ID 1",
			machineID: 1,
			wantErr:   false,
		},
		{
			name:      "valid machine ID max (1023)",
			machineID: 1023,
			wantErr:   false,
		},
		{
			name:      "invalid machine ID 1024",
			machineID: 1024,
			wantErr:   true,
		},
		{
			name:      "invalid machine ID very large",
			machineID: 65535,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen, err := NewGenerator(tt.machineID)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGenerator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && gen == nil {
				t.Error("NewGenerator() returned nil generator without error")
			}
			if !tt.wantErr && gen.machineID != tt.machineID {
				t.Errorf("NewGenerator() machineID = %v, want %v", gen.machineID, tt.machineID)
			}

		})
	}
}

func TestGenerator_Generate(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	id := gen.Generate()
	if id == 0 {
		t.Error("Generate() returned 0")
	}
	fmt.Printf("value: %d \n", id)
}

func TestGenerator_Generate_Uniqueness(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	ids := make(map[uint64]bool)
	count := 10000

	for i := 0; i < count; i++ {

		id := gen.Generate()
		if ids[id] {
			t.Errorf("Duplicate ID generated: %d", id)
		}
		ids[id] = true
	}

	if len(ids) != count {
		t.Errorf("Expected %d unique IDs, got %d", count, len(ids))
	}
	fmt.Printf("Default format (%%v): %v\n", ids)
}

func TestGenerator_Generate_Monotonic(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	var lastID uint64 = 0
	for i := 0; i < 1000; i++ {
		id := gen.Generate()
		if id <= lastID {
			t.Errorf("ID not monotonically increasing: previous=%d, current=%d", lastID, id)
		}
		lastID = id
	}
}

func TestGenerator_Generate_Concurrent(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	var wg sync.WaitGroup
	ids := make(chan uint64, 10000)
	goroutines := 10
	idsPerGoroutine := 1000

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				ids <- gen.Generate()
			}
		}()
	}

	wg.Wait()
	close(ids)

	seen := make(map[uint64]bool)
	for id := range ids {
		if seen[id] {
			t.Errorf("Duplicate ID in concurrent test: %d", id)
		}
		seen[id] = true
	}

	expected := goroutines * idsPerGoroutine
	if len(seen) != expected {
		t.Errorf("Expected %d unique IDs, got %d", expected, len(seen))
	}
}

func TestExtractTime(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	before := time.Now()
	id := gen.Generate()
	after := time.Now()

	extracted := ExtractTime(id)

	// Allow 1ms tolerance on either side
	if extracted.Before(before.Add(-time.Millisecond)) || extracted.After(after.Add(time.Millisecond)) {
		t.Errorf("ExtractTime() = %v, want between %v and %v", extracted, before, after)
	}
}

func TestExtractMachineID(t *testing.T) {
	tests := []struct {
		name      string
		machineID uint16
	}{
		{"machine ID 0", 0},
		{"machine ID 1", 1},
		{"machine ID 512", 512},
		{"machine ID 1023", 1023},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen, err := NewGenerator(tt.machineID)
			if err != nil {
				t.Fatalf("Failed to create generator: %v", err)
			}

			id := gen.Generate()
			extracted := ExtractMachineID(id)

			if extracted != tt.machineID {
				t.Errorf("ExtractMachineID() = %v, want %v", extracted, tt.machineID)
			}
		})
	}
}

func TestExtractSequence(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Generate multiple IDs quickly to get sequence numbers
	ids := make([]uint64, 10)
	for i := 0; i < 10; i++ {
		ids[i] = gen.Generate()
	}

	// The first ID in a millisecond should have sequence 0
	// Subsequent IDs in the same millisecond should have incrementing sequences
	for i, id := range ids {
		seq := ExtractSequence(id)
		// Sequence should be less than max (4095)
		if seq > 4095 {
			t.Errorf("ExtractSequence() = %v for ID %d, exceeds max 4095", seq, i)
		}
	}
}

func TestExtractSequence_Increments(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Generate IDs as fast as possible to ensure they're in the same millisecond
	var idsInSameMs []uint64
	firstID := gen.Generate()
	firstTime := ExtractTime(firstID)
	idsInSameMs = append(idsInSameMs, firstID)

	for i := 0; i < 100; i++ {
		id := gen.Generate()
		if ExtractTime(id).Equal(firstTime) {
			idsInSameMs = append(idsInSameMs, id)
		} else {
			break
		}
	}

	if len(idsInSameMs) > 1 {
		for i := 1; i < len(idsInSameMs); i++ {
			prevSeq := ExtractSequence(idsInSameMs[i-1])
			currSeq := ExtractSequence(idsInSameMs[i])
			if currSeq != prevSeq+1 {
				t.Errorf("Sequence not incrementing: prev=%d, curr=%d", prevSeq, currSeq)
			}
		}
	}
}

func TestDifferentMachines_DifferentIDs(t *testing.T) {
	gen1, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator 1: %v", err)
	}

	gen2, err := NewGenerator(2)
	if err != nil {
		t.Fatalf("Failed to create generator 2: %v", err)
	}

	id1 := gen1.Generate()
	id2 := gen2.Generate()

	if id1 == id2 {
		t.Error("Different machines generated same ID")
	}

	if ExtractMachineID(id1) != 1 {
		t.Errorf("Expected machine ID 1, got %d", ExtractMachineID(id1))
	}
	if ExtractMachineID(id2) != 2 {
		t.Errorf("Expected machine ID 2, got %d", ExtractMachineID(id2))
	}
}

func TestRoundTrip(t *testing.T) {
	machineID := uint16(42)
	gen, err := NewGenerator(machineID)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	before := time.Now()
	id := gen.Generate()
	after := time.Now()

	extractedTime := ExtractTime(id)
	extractedMachine := ExtractMachineID(id)
	extractedSeq := ExtractSequence(id)

	// Verify machine ID
	if extractedMachine != machineID {
		t.Errorf("Machine ID mismatch: got %d, want %d", extractedMachine, machineID)
	}

	// Verify time is in expected range
	if extractedTime.Before(before.Add(-time.Millisecond)) || extractedTime.After(after.Add(time.Millisecond)) {
		t.Errorf("Extracted time %v not in range [%v, %v]", extractedTime, before, after)
	}

	// First ID should have sequence 0
	if extractedSeq != 0 {
		t.Errorf("First ID sequence should be 0, got %d", extractedSeq)
	}
}

func BenchmarkGenerate(b *testing.B) {
	gen, err := NewGenerator(1)
	if err != nil {
		b.Fatalf("Failed to create generator: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Generate()
	}
}

func BenchmarkGenerate_Parallel(b *testing.B) {
	gen, err := NewGenerator(1)
	if err != nil {
		b.Fatalf("Failed to create generator: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			gen.Generate()
		}
	})
}
