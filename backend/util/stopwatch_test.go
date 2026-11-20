package util

import (
	"sync"
	"testing"
	"time"
)

func TestStopwatch_Basic(t *testing.T) {
	sw := &Stopwatch{}

	// Test initial state
	if elapsed := sw.Elapsed(); elapsed != 0 {
		t.Errorf("Expected initial elapsed time to be 0, got %v", elapsed)
	}

	// Test start and elapsed
	sw.Start()
	time.Sleep(10 * time.Millisecond)
	elapsed := sw.Elapsed()
	if elapsed < 10*time.Millisecond {
		t.Errorf("Expected at least 10ms elapsed, got %v", elapsed)
	}

	// Test stop
	sw.Stop()
	stoppedElapsed := sw.Elapsed()
	time.Sleep(10 * time.Millisecond)
	if sw.Elapsed() != stoppedElapsed {
		t.Error("Elapsed time should not increase after Stop()")
	}

	// Test reset
	sw.Reset()
	if elapsed := sw.Elapsed(); elapsed != 0 {
		t.Errorf("Expected elapsed time to be 0 after reset, got %v", elapsed)
	}
}

func TestStopwatch_StartStop(t *testing.T) {
	sw := &Stopwatch{}

	// Start, accumulate some time
	sw.Start()
	time.Sleep(10 * time.Millisecond)
	sw.Stop()
	firstElapsed := sw.Elapsed()

	// Start again, accumulate more time
	sw.Start()
	time.Sleep(10 * time.Millisecond)
	sw.Stop()
	secondElapsed := sw.Elapsed()

	if secondElapsed <= firstElapsed {
		t.Errorf("Expected elapsed time to accumulate, first=%v second=%v", firstElapsed, secondElapsed)
	}
}

func TestStopwatch_DoubleStart(t *testing.T) {
	sw := &Stopwatch{}

	sw.Start()
	time.Sleep(5 * time.Millisecond)
	firstStart := sw.Elapsed()

	// Second Start() should be no-op
	sw.Start()
	time.Sleep(5 * time.Millisecond)
	secondStart := sw.Elapsed()

	// Time should continue from first start
	if secondStart < firstStart {
		t.Error("Second Start() affected timing")
	}
}

func TestStopwatch_DoubleStop(t *testing.T) {
	sw := &Stopwatch{}

	sw.Start()
	time.Sleep(10 * time.Millisecond)
	sw.Stop()
	elapsed := sw.Elapsed()

	// Second Stop() should be no-op
	sw.Stop()
	if sw.Elapsed() != elapsed {
		t.Error("Second Stop() changed elapsed time")
	}
}

func TestStopwatch_ConcurrentAccess(t *testing.T) {
	sw := &Stopwatch{}
	var wg sync.WaitGroup

	// Test concurrent Start/Stop/Elapsed calls
	// This should not cause data races
	const goroutines = 10
	const iterations = 100

	wg.Add(goroutines * 3)

	// Concurrent starts
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				sw.Start()
				time.Sleep(time.Microsecond)
			}
		}()
	}

	// Concurrent stops
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				sw.Stop()
				time.Sleep(time.Microsecond)
			}
		}()
	}

	// Concurrent reads
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = sw.Elapsed()
				time.Sleep(time.Microsecond)
			}
		}()
	}

	wg.Wait()

	// If we got here without data races, the test passes
	// Run with: go test -race
}

func TestStopwatch_Reset(t *testing.T) {
	sw := &Stopwatch{}

	// Reset when stopped
	sw.Reset()
	if elapsed := sw.Elapsed(); elapsed != 0 {
		t.Errorf("Expected 0 after reset, got %v", elapsed)
	}

	// Reset when running
	sw.Start()
	time.Sleep(10 * time.Millisecond)
	sw.Reset()
	if elapsed := sw.Elapsed(); elapsed != 0 {
		t.Errorf("Expected 0 after reset while running, got %v", elapsed)
	}

	// After reset, should be able to start again
	sw.Start()
	time.Sleep(5 * time.Millisecond)
	elapsed := sw.Elapsed()
	if elapsed < 5*time.Millisecond {
		t.Errorf("Expected at least 5ms after reset and start, got %v", elapsed)
	}
}

func TestStopwatch_SetRate(t *testing.T) {
	sw := &Stopwatch{}

	sw.SetRate(2)
	sw.Start()
	time.Sleep(20 * time.Millisecond)
	if elapsed := sw.Elapsed(); elapsed < 40*time.Millisecond {
		t.Errorf("Expected at least 40ms at 2x rate, got %v", elapsed)
	}

	// changing rate while running keeps what was already accumulated
	sw.SetRate(0.5)
	wallStart := time.Now()
	before := sw.Elapsed()
	time.Sleep(20 * time.Millisecond)
	sw.Stop()
	wall := time.Since(wallStart)
	gained := sw.Elapsed() - before
	if gained < 10*time.Millisecond || gained > wall/2+time.Millisecond {
		t.Errorf("Expected between 10ms and %v gained at 0.5x rate, got %v", wall/2, gained)
	}
}
