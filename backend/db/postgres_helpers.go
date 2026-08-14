package db

func toInt16(v interface{}) int16 {
	if b, ok := v.(bool); ok {
		if b {
			return 1
		}
		return 0
	}
	return int16(toInt64(v))
}
