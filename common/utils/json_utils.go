package utils

import "bytes"

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

// 转json 不转义HTML特殊标签 < > &
func StructToJsonWithoutEscapeHTML(v any) string {
	if b, err := StructToJsonBytesWithoutEscapeHTML(v); err != nil {
		return ""
	} else {
		return string(b)
	}
}

// 转json 不转义HTML特殊标签 < > &
func StructToJsonBytesWithoutEscapeHTML(v any) ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(v); err != nil {
		return nil, err
	} else {
		return buffer.Bytes(), nil
	}
}
