package util

import "strconv"

func GetMapStringKey(mp map[string]string, key string, def string) string {
	if mp == nil {
		return def
	}

	value, ok := mp[key]
	if !ok {
		return def
	}
	return value
}

func GetMapIntKey(mp map[string]string, key string, def int) int {
	if mp == nil {
		return def
	}

	value, ok := mp[key]
	if !ok {
		return def
	}
	intVal, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return def
	}
	return int(intVal)
}

func GetMapInt64Key(mp map[string]string, key string, def int64) int64 {
	if mp == nil {
		return def
	}
	value, ok := mp[key]
	if !ok {
		return def
	}
	intVal, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return def
	}
	return intVal
}

func GeMapBoolKey(mp map[string]string, key string, def bool) bool {
	if mp == nil {
		return def
	}
	value, ok := mp[key]
	if !ok {
		return def
	}
	boolVal, err := strconv.ParseBool(value)
	if err != nil {
		return def
	}
	return boolVal
}
