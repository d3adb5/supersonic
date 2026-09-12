package dlna

import (
	"testing"
	"time"

	"github.com/supersonic-app/supersonic/backend/mediaprovider"
)

// The current and the next track each put a stream and a cover art URL
// into the proxy. All four have to stay reachable: if the current track's
// stream is evicted, a seek makes the renderer fetch a 404 and skip to
// whatever it has queued next.
func TestProxyKeepsCurrentAndNextTrackReachable(t *testing.T) {
	d := &DLNAPlayer{}

	added := []struct {
		name string
		url  string
		key  string
	}{
		{"current stream", "http://server/stream?id=cur", ""},
		{"current art", "/cache/art/cur.jpg", ""},
		{"next stream", "http://server/stream?id=next", ""},
		{"next art", "/cache/art/next.jpg", ""},
	}
	for i := range added {
		added[i].key = d.addURLToProxy(added[i].url)
	}

	for _, a := range added {
		got, ok := d.lookupProxyURL(a.key)
		if !ok {
			t.Errorf("%s was evicted from the proxy", a.name)
		} else if got != a.url {
			t.Errorf("%s resolved to %q, want %q", a.name, got, a.url)
		}
	}
}

func TestProxyEvictsLeastRecentlyUsed(t *testing.T) {
	d := &DLNAPlayer{}

	oldest := d.addURLToProxy("http://server/0")
	for i := 1; i < len(d.proxyURLs); i++ {
		d.addURLToProxy("http://server/" + string(rune('a'+i)))
	}
	// still the least recently used, so the next insert drops it
	if _, ok := d.lookupProxyURL(oldest); !ok {
		t.Fatal("proxy dropped an entry while it still had room")
	}

	// the lookup above promoted it, so refill past capacity to evict it
	for i := range d.proxyURLs {
		d.addURLToProxy("http://server/new" + string(rune('a'+i)))
	}
	if _, ok := d.lookupProxyURL(oldest); ok {
		t.Error("expected the least recently used entry to be evicted")
	}
}

// onTrackChange wires the player up so that a firing of the track change
// timer is observable without a renderer: with a next track queued, the
// change is reported through the OnTrackChange callback.
func onTrackChange(d *DLNAPlayer) <-chan struct{} {
	fired := make(chan struct{}, 1)
	d.nextTrackMeta = mediaprovider.MediaItemMetadata{ID: "next", Duration: time.Hour}
	d.OnTrackChange(func() { fired <- struct{}{} })
	return fired
}

func TestTrackChangeTimerFires(t *testing.T) {
	d := &DLNAPlayer{}
	fired := onTrackChange(d)
	d.setTrackChangeTimer(10 * time.Millisecond)
	select {
	case <-fired:
	case <-time.After(2 * time.Second):
		t.Fatal("track change timer never fired")
	}
}

func TestTrackChangeTimerCancelAndReschedule(t *testing.T) {
	d := &DLNAPlayer{}
	fired := onTrackChange(d)

	d.setTrackChangeTimer(10 * time.Millisecond)
	d.setTrackChangeTimer(0)
	select {
	case <-fired:
		t.Fatal("cancelled track change timer fired")
	case <-time.After(50 * time.Millisecond):
	}

	d.setTrackChangeTimer(10 * time.Millisecond)
	d.setTrackChangeTimer(time.Hour)
	select {
	case <-fired:
		t.Fatal("rescheduled track change timer fired on its old schedule")
	case <-time.After(50 * time.Millisecond):
	}
	d.setTrackChangeTimer(0)
}

// A pause that lands just as the timer fires must win. The timer can no
// longer be stopped at that point, so its firing arrives after the cancel
// and has to be recognized as stale.
func TestTrackChangeTimerFiringLosesToCancel(t *testing.T) {
	d := &DLNAPlayer{}
	fired := onTrackChange(d)

	d.setTrackChangeTimer(time.Hour)
	gen := d.timerGen
	d.setTrackChangeTimer(0)
	d.trackChangeTimerFired(gen)

	select {
	case <-fired:
		t.Fatal("a cancelled firing still changed the track")
	case <-time.After(50 * time.Millisecond):
	}
}
