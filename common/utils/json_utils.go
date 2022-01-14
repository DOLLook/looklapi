package utils

var json = ConfigWithCustomTimeFormat

// 转json
func StructToJson(t interface{}) string {
	if bytes, err := json.Marshal(t); err == nil {
		return string(bytes)
	}

	return ""
}

// json转对象
func JsonToStruct(jsonStr string, valPtr interface{}) error {
	return json.Unmarshal([]byte(jsonStr), valPtr)
}
