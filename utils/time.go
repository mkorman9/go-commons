package utils

import "time"

func TimePtr(t time.Time) *time.Time {
	return &t
}

func TimePtrToUnix(t *time.Time) *int64 {
	if t == nil {
		return nil
	}

	value := (*t).Unix()
	return &value
}

func UnixPtrToTime(u *int64) *time.Time {
	if u == nil {
		return nil
	}

	value := time.Unix(*u, 0)
	return &value
}
