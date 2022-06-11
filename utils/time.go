package utils

import "time"

func ToPtr[V any](v V) *V {
	return &v
}

func TimePtrToUnix(t *time.Time) *int64 {
	if t == nil {
		return nil
	}

	value := (*t).UTC().Unix()
	return &value
}

func UnixPtrToTime(u *int64) *time.Time {
	if u == nil {
		return nil
	}

	value := time.Unix(*u, 0).UTC()
	return &value
}
