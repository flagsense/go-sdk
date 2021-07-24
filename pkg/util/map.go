package util

func ContainsKey(mapObject map[string]interface{}, key string) bool {
	if mapObject == nil {
		return false
	}
	_, ok := mapObject[key]
	return ok
}

func SafeGetValue(mapObject map[string]interface{}, key string) (interface{}, bool) {
	if mapObject == nil {
		return nil, false
	}
	val, ok := mapObject[key]
	if ok {
		return val, true
	}
	return nil, false
}
