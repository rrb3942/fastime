//go:build aix || darwin || dragonfly || freebsd || (js && wasm) || linux || nacl || netbsd || openbsd || solaris
// +build aix darwin dragonfly freebsd js,wasm linux nacl netbsd openbsd solaris

package fastime

import (
	"syscall"
	"time"
)

func (f *FastTime) now() (now time.Time) {
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
