//go:build aix || darwin || dragonfly || freebsd || (js && wasm) || linux || nacl || netbsd || openbsd || solaris
// +build aix darwin dragonfly freebsd js,wasm linux nacl netbsd openbsd solaris

package fastime

import (
	"syscall"
	"time"
)

// now returns the current time. This is the POSIX-specific implementation
// using syscall.Gettimeofday for potentially higher precision.
// It falls back to time.Now() if syscall.Gettimeofday fails.
// The returned time is in the location set for the Time instance.
func (f *Time) now() (now time.Time) {
	var timeValue syscall.Timeval
	err := syscall.Gettimeofday(&timeValue)
	loc := f.GetLocation()

	if err != nil {
		now = time.Now()
		if loc != nil {
			return now.In(loc)
		}

		return now
	}

	now = time.Unix(0, syscall.TimevalToNsec(timeValue))
	if loc != nil {
		return now.In(loc)
	}

	return now
}
