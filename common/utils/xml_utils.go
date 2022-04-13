package utils

import (
	"encoding/xml"
)

/**
结构体转xml
*/
func ObjectToXml(obj interface{}) (string, error) {
	var xmlStr []byte
	if data, err := xml.MarshalIndent(&obj, "", "  "); err != nil {
		return "", err
	} else {
		xmlStr = data
	}
	return string(xmlStr), nil
}

//xml转结构体
func XmlToObject(xmlStr []byte, v interface{}) error {
	return xml.Unmarshal(xmlStr, v)
}
