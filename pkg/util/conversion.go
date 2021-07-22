package util

func ConvertToFloat64(source interface{}) float64 {
	switch source.(type) {
	case float64:
		return source.(float64)
	case float32:
		return float64(source.(float32))
	case int:
		return float64(source.(int))
	case int64:
		return float64(source.(int64))
	case int32:
		return float64(source.(int32))
	default:
		return source.(float64)
	}
}

func ConvertToInt32(source interface{}) int32 {
	switch source.(type) {
	case int:
		return int32(source.(int))
	case int64:
		return int32(source.(int64))
	case float32:
		return int32(source.(float32))
	case float64:
		return int32(source.(float64))
	default:
		return source.(int32)
	}
}
