package models

import "time"

// NowMillis returns the current time in Unix milliseconds.
func NowMillis() int64 {
	return time.Now().UnixMilli()
}
