package utils

var json = ConfigWithCustomTimeFormat

// 转json
func StructToJson(t interface{}) string {
	if bytes, err := json.Marshal(t); err == nil {
		return string(bytes)
	}

	return ""
}

// 转json bytes
func StructToJsonBytes(t interface{}) ([]byte, error) {
	return json.Marshal(t)
}

// json转对象
func JsonToStruct(jsonStr string, valPtr interface{}) error {
	return json.Unmarshal([]byte(jsonStr), valPtr)
}

// json转对象
func JsonBytesToStruct(data []byte, valPtr interface{}) error {
	return json.Unmarshal(data, valPtr)
}
