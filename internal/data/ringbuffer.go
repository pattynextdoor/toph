package data

// RingBuffer is a fixed-capacity circular buffer of Events.
// When full, new pushes overwrite the oldest entry.
// Used for the 1,000-event activity feed.
type RingBuffer struct {
	buf   []Event
	cap   int
	head  int
	count int
}

// NewRingBuffer creates a ring buffer with the given capacity.
func NewRingBuffer(capacity int) *RingBuffer {
	return &RingBuffer{
		buf: make([]Event, capacity),
		cap: capacity,
	}
}

// Push appends an event, overwriting the oldest if at capacity.
func (rb *RingBuffer) Push(e Event) {
	rb.buf[rb.head] = e
	rb.head = (rb.head + 1) % rb.cap
	if rb.count < rb.cap {
		rb.count++
	}
}

// Len returns the number of events currently stored.
func (rb *RingBuffer) Len() int { return rb.count }

// All returns all stored events in chronological order (oldest first).
func (rb *RingBuffer) All() []Event { return rb.Slice(0, rb.count) }

// Slice returns events from index start (inclusive) to end (exclusive),
// in chronological order. Indices are relative to the oldest event (0 = oldest).
func (rb *RingBuffer) Slice(start, end int) []Event {
	if start < 0 {
		start = 0
	}
	if end > rb.count {
		end = rb.count
	}
	if start >= end {
		return nil
	}

	result := make([]Event, 0, end-start)
	oldest := (rb.head - rb.count + rb.cap) % rb.cap
	for i := start; i < end; i++ {
		idx := (oldest + i) % rb.cap
		result = append(result, rb.buf[idx])
	}
	return result
}

// Last returns the most recent n events in chronological order.
func (rb *RingBuffer) Last(n int) []Event {
	if n > rb.count {
		n = rb.count
	}
	return rb.Slice(rb.count-n, rb.count)
}
