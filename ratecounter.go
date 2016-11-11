package ratecounter

import (
	"strconv"
	"sync"
	"time"
)

// A RateCounter is a thread-safe counter which returns the number of times
// 'Incr' has been called in the last interval
type RateCounter struct {
	periods        []Counter
	periodsTotal   int64
	periodsNow     int64
	periodsCurrent int64
	counter        Counter
	quantile       time.Duration
	tail           time.Duration
	ticker         *time.Ticker
	close          chan bool
	sync.RWMutex
}

// NewRateCounter constructs a new RateCounter, for the quantile provided and tail
func NewRateCounter(quantile time.Duration, tail time.Duration) *RateCounter {
	if tail == 0 || tail < quantile {
		tail = quantile
	}
	var periodsTotal int64
	if quantile == 0 {
		periodsTotal = 0
	} else {
		periodsTotal = int64(tail / quantile)
	}

	rc := &RateCounter{
		periods:        make([]Counter, periodsTotal),
		quantile:       quantile,
		tail:           tail,
		periodsTotal:   periodsTotal,
		periodsNow:     0,
		periodsCurrent: 0,
		close:          make(chan bool),
	}
	if quantile != 0 {
		rc.ticker = time.NewTicker(rc.quantile)
		go rc.mover()
	}
	return rc
}

func (r *RateCounter) mover() {
	for {
		select {
		case <-r.ticker.C:
			r.Lock()
			v := r.Rate()
			cnt := Counter(v)
			r.Incr(-1 * v)
			r.periodsNow = int64(len(r.periods))
			r.periodsCurrent = (r.periodsCurrent + 1) % r.periodsTotal
			if r.periodsNow < r.periodsTotal {
				r.periods = append(r.periods, cnt)
			} else {
				r.periods[r.periodsCurrent] = cnt
			}
			r.Unlock()

		case <-r.close:
			return
		}
	}
}

// Incr - incriments current period by given value
func (r *RateCounter) Incr(val int64) {
	r.counter.Incr(val)
}

// Hit - incriments current period by 1
func (r *RateCounter) Hit() {
	r.Incr(1)
}

// Rate returns the current number of events in the last interval
func (r *RateCounter) Rate() int64 {
	return r.counter.Value()
	// return r.periods[len(r.periods)-1].Value()
}

// TailRate returns the number of events across all tail
func (r *RateCounter) TailRate() int64 {
	r.RLock()
	defer r.RUnlock()
	var total int64
	for _, v := range r.periods {
		total += v.Value()
	}
	return total
}

// TotalRate returns the total number of events
func (r *RateCounter) TotalRate() int64 {
	return r.TailRate() + r.Rate()
}

func (r *RateCounter) String() string {
	return strconv.FormatInt(r.Rate(), 10)
}
