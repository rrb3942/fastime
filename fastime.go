package fastime

import (
	"context"
	"math"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

// *FastTime is FastTime's base struct, it's stores atomic time object.
type FastTime struct {
	location      atomic.Pointer[time.Location]
	format        atomic.Pointer[string]
	ft            atomic.Pointer[[]byte]
	t             atomic.Pointer[time.Time]
	wg            sync.WaitGroup
	ut            int64
	correctionDur time.Duration
	unt           int64
	dur           int64
	mu            sync.Mutex
	running       atomic.Bool
	uut           uint32
	formatValid   atomic.Bool
	uunt          uint32
}

const (
	bufSize   = 64
	bufMargin = 10
)

func New() (f *FastTime) {
	f = &FastTime{
		ut:            math.MaxInt64,
		unt:           math.MaxInt64,
		uut:           math.MaxUint32,
		uunt:          math.MaxUint32,
		correctionDur: time.Millisecond * 100,
	}

	form := time.RFC3339
	f.format.Store(&form)

	loc := func() (loc *time.Location) {
		tz, ok := syscall.Getenv("TZ")
		if ok && tz != "" {
			var err error

			loc, err = time.LoadLocation(tz)
			if err == nil {
				return loc
			}
		}

		return new(time.Location)
	}()

	f.location.Store(loc)

	buf := f.newBuffer(len(form) + bufMargin)
	f.ft.Store(&buf)

	return f.refresh()
}

func (f *FastTime) update() (ft *FastTime) {
	return f.store(f.Now().Add(time.Duration(atomic.LoadInt64(&f.dur))))
}

func (f *FastTime) refresh() (ft *FastTime) {
	return f.store(f.now())
}

func (f *FastTime) newBuffer(maxSize int) (b []byte) {
	if maxSize < bufSize {
		var buf [bufSize]byte
		b = buf[:0]
	} else {
		b = make([]byte, 0, maxSize)
	}

	return b
}

func (f *FastTime) store(t time.Time) (ft *FastTime) {
	f.t.Store(&t)
	f.formatValid.Store(false)

	ut := t.Unix()
	unt := t.UnixNano()

	atomic.StoreInt64(&f.ut, ut)
	atomic.StoreInt64(&f.unt, unt)
	atomic.StoreUint32(&f.uut, *(*uint32)(unsafe.Pointer(&ut)))
	atomic.StoreUint32(&f.uunt, *(*uint32)(unsafe.Pointer(&unt)))

	return f
}

func (f *FastTime) IsDaemonRunning() (running bool) {
	return f.running.Load()
}

func (f *FastTime) GetLocation() (loc *time.Location) {
	loc = f.location.Load()
	if loc == nil {
		return nil
	}

	return loc
}

func (f *FastTime) GetFormat() (form string) {
	return *f.format.Load()
}

// SetLocation replaces time location.
func (f *FastTime) SetLocation(loc *time.Location) (ft *FastTime) {
	if loc == nil {
		return f
	}

	f.location.Store(loc)
	f.refresh()

	return f
}

// SetFormat replaces time format.
func (f *FastTime) SetFormat(format string) (ft *FastTime) {
	f.format.Store(&format)
	f.formatValid.Store(false)
	f.refresh()

	return f
}

// Now returns current time.
func (f *FastTime) Now() (t time.Time) {
	return *f.t.Load()
}

// Stop stops stopping time refresh daemon.
func (f *FastTime) Stop() {
	f.mu.Lock()
	f.stop()
	f.mu.Unlock()
}

func (f *FastTime) stop() {
	if f.IsDaemonRunning() {
		atomic.StoreInt64(&f.dur, 0)
	}

	f.wg.Wait()
}

func (f *FastTime) Since(t time.Time) (dur time.Duration) {
	return f.Now().Sub(t)
}

// UnixNow returns current unix time.
func (f *FastTime) UnixNow() (now int64) {
	return atomic.LoadInt64(&f.ut)
}

// UnixNow returns current unix time.
func (f *FastTime) UnixUNow() (now uint32) {
	return atomic.LoadUint32(&f.uut)
}

// UnixNanoNow returns current unix nano time.
func (f *FastTime) UnixNanoNow() (now int64) {
	return atomic.LoadInt64(&f.unt)
}

// UnixNanoNow returns current unix nano time.
func (f *FastTime) UnixUNanoNow() (now uint32) {
	return atomic.LoadUint32(&f.uunt)
}

// FormattedNow returns formatted byte time.
func (f *FastTime) FormattedNow() (now []byte) {
	// only update formatted value on swap
	if f.formatValid.CompareAndSwap(false, true) {
		form := f.GetFormat()
		buf := f.Now().AppendFormat(f.newBuffer(len(form)+bufMargin), form)
		f.ft.Store(&buf)
	}

	return *f.ft.Load()
}

// StartTimerD provides time refresh daemon.
func (f *FastTime) StartTimerD(ctx context.Context, dur time.Duration) (ft *FastTime) {
	f.mu.Lock()
	defer f.mu.Unlock()
	// if the daemon was already running, restart
	if f.IsDaemonRunning() {
		f.stop()
	}

	f.running.Store(true)
	f.dur = math.MaxInt64
	atomic.StoreInt64(&f.dur, dur.Nanoseconds())
	ticker := time.NewTicker(time.Duration(atomic.LoadInt64(&f.dur)))
	lastCorrection := f.now()
	f.wg.Add(1)
	f.refresh()

	go func() {
		// daemon cleanup
		defer func() {
			f.running.Store(false)
			ticker.Stop()
			f.wg.Done()
		}()

		for atomic.LoadInt64(&f.dur) > 0 {
			tickTime := <-ticker.C
			// rely on ticker for approximation
			if tickTime.Sub(lastCorrection) < f.correctionDur {
				f.update()
			} else { // correct the system time at a fixed interval
				select {
				case <-ctx.Done():
					return
				default:
				}
				f.refresh()

				lastCorrection = tickTime
			}
		}
	}()

	return f
}
