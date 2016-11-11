package ratecounter

// From github.com/paulbellamy/ratecounter

import "sync/atomic"

// A Counter is a thread-safe counter implementation
type Counter int64

// Incr - increment the counter by some value
func (c *Counter) Incr(val int64) {
	atomic.AddInt64((*int64)(c), val)
}

// Value - return the counter's current value
func (c *Counter) Value() int64 {
	return atomic.LoadInt64((*int64)(c))
}
