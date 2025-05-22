package fastime

import (
	"context"
	"sync"
	"time"
)

var (
	once    sync.Once
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

func IsDaemonRunning() (running bool) {
	return Default.IsDaemonRunning()
}

func GetLocation() (loc *time.Location) {
	return Default.GetLocation()
}

func GetFormat() (form string) {
	return Default.GetFormat()
}

// SetLocation replaces time location.
func SetLocation(location *time.Location) (ft *Time) {
	return Default.SetLocation(location)
}

// SetFormat replaces time format.
func SetFormat(format string) (ft *Time) {
	return Default.SetFormat(format)
}

// Now returns current time.
func Now() (now time.Time) {
	return Default.Now()
}

// Since returns the time elapsed since t.
// It is shorthand for fastime.Now().Sub(t).
func Since(t time.Time) (dur time.Duration) {
	return Default.Since(t)
}

// Stop stops stopping time refresh daemon.
func Stop() {
	Default.Stop()
}

// UnixNow returns current unix time.
func UnixNow() (now int64) {
	return Default.UnixNow()
}

// UnixUNow returns current unix time.
func UnixUNow() (now uint32) {
	return Default.UnixUNow()
}

// UnixNanoNow returns current unix nano time.
func UnixNanoNow() (now int64) {
	return Default.UnixNanoNow()
}

// UnixUNanoNow returns current unix nano time.
func UnixUNanoNow() (now uint32) {
	return Default.UnixUNanoNow()
}

// FormattedNow returns formatted byte time.
func FormattedNow() (now []byte) {
	return Default.FormattedNow()
}

// StartTimerD provides time refresh daemon.
func StartTimerD(ctx context.Context, dur time.Duration) (ft *Time) {
	return Default.StartTimerD(ctx, dur)
}
