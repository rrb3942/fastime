//go:build windows
// +build windows

package fastime

import "time"

func (f *FastTime) now() time.Time {
	return time.Now().In(f.GetLocation())
}
