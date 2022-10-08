package util

func GetStringArrayKey(key string, val map[string]interface{}) []string {
	aInterface := GetInterfaceArray(key, val)
	if aInterface == nil {
		return []string{}
	}

	aString := make([]string, len(aInterface))
	for i, v := range aInterface {
		aString[i] = v.(string)
	}
	return aString
}

func GetStringKey(key string, val map[string]interface{}) string {
	aInterface, ok := val[key].(interface{})
	if !ok {
		return ""
	}
	return aInterface.(string)
}

func GetIntKey(key string, val map[string]interface{}) int {
	aInterface, ok := val[key].(interface{})
	if !ok {
		return 0
	}
	return aInterface.(int)
}

func GetBoolKey(key string, val map[string]interface{}) bool {
	aInterface, ok := val[key].(interface{})
	if !ok {
		return false
	}
	return aInterface.(bool)
}

func GetInterfaceKey(key string, val map[string]interface{}) interface{} {
	aInterface, ok := val[key].(interface{})
	if !ok {
		return nil
	}
	return aInterface
}

func GetInterfaceArray(key string, val map[string]interface{}) []interface{} {
	aInterface, ok := val[key].([]interface{})
	if !ok {
		return nil
	}
	return aInterface
}

func GetInterfaceMapKey(key string, val map[string]interface{}) map[string]interface{} {
	aInterface, ok := val[key].(map[string]interface{})
	if !ok {
		return nil
	}
	return aInterface
}
