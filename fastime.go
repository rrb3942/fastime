// Package fastime is a low-overhead, high-performance time package that caches time.
// It is designed to be a faster alternative to time.Now() for applications
// that require frequent time lookups.
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

// Time is the core struct of the fastime package. It stores and manages
// the cached time, location, and format string. It uses atomic operations
// for thread-safe access to its fields.
type Time struct {
	location      atomic.Pointer[time.Location] // Current time zone location.
	format        atomic.Pointer[string]        // Current time format string.
	ft            atomic.Pointer[string]        // Cached formatted time as a byte slice.
	t             atomic.Pointer[time.Time]     // Cached time.Time object.
	wg            sync.WaitGroup                // WaitGroup for managing the timer goroutine.
	ut            atomic.Int64                  // Cached Unix time in seconds.
	correctionDur time.Duration                 // Interval for system time correction (in nanoseconds).
	unt           atomic.Int64                  // Cached Unix time in nanoseconds.
	dur           atomic.Int64                  // Refresh duration in nanoseconds for the timer goroutine.
	mu            sync.Mutex                    // Mutex for controlling access to StartTimerD and Stop.
	running       atomic.Bool                   // Flag indicating if the timer goroutine is running.
	uut           atomic.Uint32                 // Cached Unix time in seconds (uint32).
	formatValid   atomic.Bool                   // Flag indicating if the formatted time (ft) is valid.
	uunt          atomic.Uint32                 // Cached Unix time in nanoseconds (uint32).
}

// New creates and initializes a new Time instance.
// It sets the initial time, default format (time.RFC3339), and location.
// The location is determined by the TZ environment variable if set, otherwise it defaults to UTC.
func New() (f *Time) {
	f = &Time{}
	f.ut.Store(math.MaxInt64)
	f.unt.Store(math.MaxInt64)
	f.uut.Store(math.MaxUint32)
	f.uunt.Store(math.MaxUint32)
	f.correctionDur = time.Millisecond * 100

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

	fmt := time.Now().Format(*f.format.Load())
	f.ft.Store(&fmt)

	return f.refresh()
}

func (f *Time) update() (ft *Time) {
	return f.store(f.Now().Add(time.Duration(f.dur.Load())))
}

func (f *Time) refresh() (ft *Time) {
	return f.store(f.now())
}

func (f *Time) store(t time.Time) (ft *Time) {
	f.t.Store(&t)
	f.formatValid.Store(false)

	ut := t.Unix()
	unt := t.UnixNano()

	f.ut.Store(ut)
	f.unt.Store(unt)
	f.uut.Store(*(*uint32)(unsafe.Pointer(&ut)))
	f.uunt.Store(*(*uint32)(unsafe.Pointer(&unt)))

	return f
}

// IsDaemonRunning reports whether the background refresh daemon is running.
func (f *Time) IsDaemonRunning() (running bool) {
	return f.running.Load()
}

// GetLocation returns the current time.Location used by the Time instance.
// It returns nil if the location is not properly initialized (e.g. TZ env var points to a non-existent location).
func (f *Time) GetLocation() (loc *time.Location) {
	loc = f.location.Load()
	if loc == nil {
		return nil
	}

	return loc
}

// GetFormat returns the current format string used for FormattedNow.
func (f *Time) GetFormat() (form string) {
	return *f.format.Load()
}

// SetLocation updates the time location for the Time instance.
// If loc is nil, the current location remains unchanged.
// After setting the location, the cached time is refreshed.
func (f *Time) SetLocation(loc *time.Location) (ft *Time) {
	if loc == nil {
		return f
	}

	f.location.Store(loc)
	f.refresh()

	return f
}

// SetFormat updates the time format string for the Time instance.
// After setting the format, the cached time is refreshed and the
// formatted time cache is invalidated.
func (f *Time) SetFormat(format string) (ft *Time) {
	f.format.Store(&format)
	f.formatValid.Store(false)
	f.refresh()

	return f
}

// Now returns the cached current time.
func (f *Time) Now() (t time.Time) {
	return *f.t.Load()
}

// Stop halts the background time refresh daemon if it is running.
// It waits for the daemon goroutine to exit before returning.
func (f *Time) Stop() {
	f.mu.Lock()
	f.stop()
	f.mu.Unlock()
}

func (f *Time) stop() {
	if f.IsDaemonRunning() {
		f.dur.Store(0)
	}

	f.wg.Wait()
}

// Since returns the time elapsed since t.
// It is shorthand for f.Now().Sub(t).
func (f *Time) Since(t time.Time) (dur time.Duration) {
	return f.Now().Sub(t)
}

// UnixNow returns the cached current Unix time (seconds since January 1, 1970 UTC).
func (f *Time) UnixNow() (now int64) {
	return f.ut.Load()
}

// UnixUNow returns the cached current Unix time as a uint32 (seconds since January 1, 1970 UTC).
// Note: This will overflow in the year 2106.
func (f *Time) UnixUNow() (now uint32) {
	return f.uut.Load()
}

// UnixNanoNow returns the cached current Unix time in nanoseconds.
func (f *Time) UnixNanoNow() (now int64) {
	return f.unt.Load()
}

// UnixUNanoNow returns the cached current Unix time in nanoseconds as a uint32.
// Note: This will overflow frequently (approximately every 4.29 seconds).
// It is generally recommended to use UnixNanoNow for nanosecond precision.
func (f *Time) UnixUNanoNow() (now uint32) {
	return f.uunt.Load()
}

// FormattedNow returns the cached current time formatted as a byte slice
// according to the format string set by SetFormat or the default (time.RFC3339).
// The formatted time is cached and only recomputed if the format or time changes.
func (f *Time) FormattedNow() string {
	// only update formatted value on swap
	if f.formatValid.CompareAndSwap(false, true) {
		fmt := time.Now().Format(*f.format.Load())
		f.ft.Store(&fmt)

		return fmt
	}

	return *f.ft.Load()
}

// StartTimerD starts a background goroutine (daemon) that periodically refreshes the cached time.
// The refresh interval is specified by the 'dur' parameter.
// If a daemon is already running, it is stopped and a new one is started.
// The daemon will also stop if the provided context is cancelled.
// It uses a ticker for regular updates and periodically corrects against the system clock
// to mitigate drift. The correction interval is defined by `f.correctionDur`.
func (f *Time) StartTimerD(ctx context.Context, dur time.Duration) (ft *Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	// if the daemon was already running, restart
	if f.IsDaemonRunning() {
		f.stop()
	}

	f.running.Store(true)
	f.dur.Store(math.MaxInt64)
	f.dur.Store(dur.Nanoseconds())
	ticker := time.NewTicker(time.Duration(f.dur.Load()))
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

		for f.dur.Load() > 0 {
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
