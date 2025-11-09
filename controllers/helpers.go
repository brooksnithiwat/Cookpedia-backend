package controllers

import (
	"strconv"
)

// ToInt64 normalizes different numeric or string representations to int64.
// Returns 0 on parse/convert failure.
func ToInt64(v interface{}) int64 {
	switch t := v.(type) {
	case int64:
		return t
	case int:
		return int64(t)
	case float64:
		return int64(t)
	case string:
		if tmp, err := strconv.ParseInt(t, 10, 64); err == nil {
			return tmp
		}
		return 0
	default:
		return 0
	}
}
