package util

func GetValue[V comparable, T any](key V, val map[V]interface{}, defVal T) T {
	if val == nil {
		return defVal
	}
	aInterface, ok := val[key]
	if !ok {
		return defVal
	}

	return ToType(aInterface, defVal)
}

func GetValueArray[T any](key string, val map[string]interface{}, defVal []T) []T {
	if val == nil {
		return nil
	}

	aInterface, ok := val[key]
	if !ok {
		return defVal
	}

	var iArr []interface{}
	switch aInterface.(type) {
	case []interface{}:
		iArr = aInterface.([]interface{})
	default:
		return defVal
	}

	arr := make([]T, len(iArr))
	for i, v := range iArr {
		switch v.(type) {
		case T:
			arr[i] = v.(T)
		}
	}
	return arr
}

func GetTypedMap[V comparable, T any](val map[V]interface{}, def T) map[V]T {
	mp := map[V]T{}
	for k, v := range val {
		switch v.(type) {
		case T:
			mp[k] = v.(T)
		}
	}
	return mp
}

func GetValueMap[V comparable, T any](key string, val map[string]interface{}, def map[V]T) map[V]T {
	if val == nil {
		return def
	}
	aInterface, ok := val[key]
	if !ok {
		return nil
	}

	switch aInterface.(type) {
	case map[V]T:
		return aInterface.(map[V]T)
	case map[V]interface{}:
		var dd T
		return GetTypedMap(aInterface.(map[V]interface{}), dd)
	}

	return def
}

func ToType[T any](val interface{}, defVal T) T {
	if val == nil {
		return defVal
	}
	switch val.(type) {
	case T:
		return val.(T)
	}
	return defVal
}
