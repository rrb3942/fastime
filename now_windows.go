//go:build windows
// +build windows

package fastime

import "time"

// now returns the current time. This is the Windows-specific implementation
// using time.Now(). The returned time is in the location set for the Time instance.
func (f *Time) now() time.Time {
	return time.Now().In(f.GetLocation())
}
