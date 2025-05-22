package fastime

import (
	"context"
	"sync"
	"time"
)

var (
	once sync.Once
	// Default is the default global Time instance used by top-level functions.
	// It is initialized in an init() function and starts a background timer
	// with a default refresh duration.
	Default *Time
)

const (
	defaultRefreshDuration = time.Millisecond * 5
)

func init() {
	once.Do(func() {
		Default = New().StartTimerD(context.Background(), defaultRefreshDuration)
	})
}

// IsDaemonRunning reports whether the background refresh daemon for the Default instance is running.
func IsDaemonRunning() (running bool) {
	return Default.IsDaemonRunning()
}

// GetLocation returns the current time.Location used by the Default instance.
func GetLocation() (loc *time.Location) {
	return Default.GetLocation()
}

// GetFormat returns the current format string used by the Default instance.
func GetFormat() (form string) {
	return Default.GetFormat()
}

// SetLocation updates the time location for the Default instance.
// See (*Time).SetLocation for more details.
func SetLocation(location *time.Location) (ft *Time) {
	return Default.SetLocation(location)
}

// SetFormat updates the time format string for the Default instance.
// See (*Time).SetFormat for more details.
func SetFormat(format string) (ft *Time) {
	return Default.SetFormat(format)
}

// Now returns the cached current time from the Default instance.
// See (*Time).Now for more details.
func Now() (now time.Time) {
	return Default.Now()
}

// Since returns the time elapsed since t, using the Default instance's current time.
// It is shorthand for fastime.Now().Sub(t).
// See (*Time).Since for more details.
func Since(t time.Time) (dur time.Duration) {
	return Default.Since(t)
}

// Stop halts the background time refresh daemon for the Default instance.
// See (*Time).Stop for more details.
func Stop() {
	Default.Stop()
}

// UnixNow returns the cached current Unix time from the Default instance.
// See (*Time).UnixNow for more details.
func UnixNow() (now int64) {
	return Default.UnixNow()
}

// UnixUNow returns the cached current Unix time as a uint32 from the Default instance.
// See (*Time).UnixUNow for more details.
func UnixUNow() (now uint32) {
	return Default.UnixUNow()
}

// UnixNanoNow returns the cached current Unix time in nanoseconds from the Default instance.
// See (*Time).UnixNanoNow for more details.
func UnixNanoNow() (now int64) {
	return Default.UnixNanoNow()
}

// UnixUNanoNow returns the cached current Unix time in nanoseconds as a uint32 from the Default instance.
// See (*Time).UnixUNanoNow for more details.
func UnixUNanoNow() (now uint32) {
	return Default.UnixUNanoNow()
}

// FormattedNow returns the cached current time formatted as a byte slice from the Default instance.
// See (*Time).FormattedNow for more details.
func FormattedNow() (now []byte) {
	return Default.FormattedNow()
}

// StartTimerD starts or restarts the background timer daemon for the Default instance.
// See (*Time).StartTimerD for more details.
func StartTimerD(ctx context.Context, dur time.Duration) (ft *Time) {
	return Default.StartTimerD(ctx, dur)
}
