package util

import (
	"sync"
	"time"
)

// Stopwatch is a thread-safe timer for measuring elapsed time.
// It can be started, stopped, and reset, and supports reading
// the elapsed time while running or stopped.
//
// The rate at which time accumulates can be changed with SetRate,
// e.g. to measure media time while playing at a speed other than 1x.
type Stopwatch struct {
	mu      sync.Mutex
	running bool
	started time.Time
	elapsed time.Duration
	rate    float64 // zero value means 1
}

func (s *Stopwatch) sinceStarted() time.Duration {
	d := time.Since(s.started)
	if s.rate > 0 {
		d = time.Duration(float64(d) * s.rate)
	}
	return d
}

// SetRate changes the rate at which elapsed time accumulates from now on.
// Time already accumulated is unaffected. A rate <= 0 resets it to 1.
func (s *Stopwatch) SetRate(rate float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		s.elapsed += s.sinceStarted()
		s.started = time.Now()
	}
	s.rate = rate
}

// Start begins or resumes the stopwatch.
// If already running, this is a no-op.
func (s *Stopwatch) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return
	}
	s.started = time.Now()
	s.running = true
}

// Stop pauses the stopwatch and accumulates the elapsed time.
// If already stopped, this is a no-op.
func (s *Stopwatch) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}
	s.elapsed += s.sinceStarted()
	s.running = false
}

// Elapsed returns the total elapsed time.
// If the stopwatch is running, includes time since last Start().
// Safe to call concurrently with other methods.
func (s *Stopwatch) Elapsed() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	e := s.elapsed
	if s.running {
		e += s.sinceStarted()
	}
	return e
}

// Reset stops the stopwatch and clears the elapsed time.
func (s *Stopwatch) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.running = false
	s.elapsed = time.Duration(0)
}
