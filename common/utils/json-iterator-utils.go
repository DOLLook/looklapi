package utils

import (
	"github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
	"strings"
	"time"
	"unsafe"
)

// time format alias
const (
	ANSIC          = "ANSIC"
	UnixDate       = "UnixDate"
	RubyDate       = "RubyDate"
	RFC822         = "RFC822"
	RFC822Z        = "RFC822Z"
	RFC850         = "RFC850"
	RFC1123        = "RFC1123"
	RFC1123Z       = "RFC1123Z"
	RFC3339        = "RFC3339"
	RFC3339Nano    = "RFC3339Nano"
	Kitchen        = "Kitchen"
	Stamp          = "Stamp"
	StampMilli     = "StampMilli"
	StampMicro     = "StampMicro"
	StampNano      = "StampNano"
	SimpleDatetime = "SimpleDatetime" // 自定义基本格式
)

// time zone alias
const (
	Local = "Local"
	UTC   = "UTC"
)

const (
	tagNameTimeFormat   = "time_format"
	tagNameTimeLocation = "time_location"
)

var ConfigWithCustomTimeFormat = jsoniter.ConfigCompatibleWithStandardLibrary

var formatAlias = map[string]string{
	ANSIC:          time.ANSIC,
	UnixDate:       time.UnixDate,
	RubyDate:       time.RubyDate,
	RFC822:         time.RFC822,
	RFC822Z:        time.RFC822Z,
	RFC850:         time.RFC850,
	RFC1123:        time.RFC1123,
	RFC1123Z:       time.RFC1123Z,
	RFC3339:        time.RFC3339,
	RFC3339Nano:    time.RFC3339Nano,
	Kitchen:        time.Kitchen,
	Stamp:          time.Stamp,
	StampMilli:     time.StampMilli,
	StampMicro:     time.StampMicro,
	StampNano:      time.StampNano,
	SimpleDatetime: "2006-01-02 15:04:05",
}

var localeAlias = map[string]*time.Location{
	Local: time.Local,
	UTC:   time.UTC,
}

var (
	defaultFormat = time.RFC3339
	defaultLocale = time.Local
)

func init() {
	ConfigWithCustomTimeFormat.RegisterExtension(&CustomTimeExtension{})
	extra.RegisterFuzzyDecoders()
}

func AddTimeFormatAlias(alias, format string) {
	formatAlias[alias] = format
}

func AddLocaleAlias(alias string, locale *time.Location) {
	localeAlias[alias] = locale
}

func SetDefaultTimeFormat(timeFormat string, timeLocation *time.Location) {
	defaultFormat = timeFormat
	defaultLocale = timeLocation
}

type CustomTimeExtension struct {
	jsoniter.DummyExtension
}

func (extension *CustomTimeExtension) UpdateStructDescriptor(structDescriptor *jsoniter.StructDescriptor) {
	for _, binding := range structDescriptor.Fields {
		var typeErr error
		var isPtr bool
		typeName := binding.Field.Type().String()
		if typeName == "time.Time" {
			isPtr = false
		} else if typeName == "*time.Time" {
			isPtr = true
		} else {
			continue
		}

		timeFormat := defaultFormat
		formatTag := binding.Field.Tag().Get(tagNameTimeFormat)
		if format, ok := formatAlias[formatTag]; ok {
			timeFormat = format
		} else if formatTag != "" {
			timeFormat = formatTag
		}
		locale := defaultLocale
		if localeTag := binding.Field.Tag().Get(tagNameTimeLocation); localeTag != "" {
			if loc, ok := localeAlias[localeTag]; ok {
				locale = loc
			} else {
				loc, err := time.LoadLocation(localeTag)
				if err != nil {
					typeErr = err
				} else {
					locale = loc
				}
			}
		}

		binding.Encoder = &funcEncoder{fun: func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			if typeErr != nil {
				stream.Error = typeErr
				return
			}

			var tp *time.Time
			if isPtr {
				tpp := (**time.Time)(ptr)
				tp = *(tpp)
			} else {
				tp = (*time.Time)(ptr)
			}

			if tp != nil {
				lt := tp.In(locale)
				str := lt.Format(timeFormat)
				stream.WriteString(str)
			} else {
				stream.Write([]byte("null"))
			}
		}, isEmptyFunc: func(ptr unsafe.Pointer) bool {
			if typeErr != nil {
				return false
			}

			var tp *time.Time
			if isPtr {
				tpp := (**time.Time)(ptr)
				tp = *(tpp)
			} else {
				tp = (*time.Time)(ptr)
			}

			if tp != nil && !tp.IsZero() {
				return false
			} else {
				return true
			}
		}}
		binding.Decoder = &funcDecoder{fun: func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			if typeErr != nil {
				iter.Error = typeErr
				return
			}

			str := iter.ReadString()
			var t *time.Time
			if str != "" {
				str = dateTimeSimplify(str, timeFormat)
				var err error
				// 始终用2006-01-02 15:04:05 decode 以保持时间精度。timeFormat仅作为序列化使用
				//tmp, err := time.ParseInLocation(timeFormat, str, locale)
				tmp, err := time.ParseInLocation("2006-01-02 15:04:05", str, locale)
				if err != nil {
					iter.Error = err
					return
				}
				t = &tmp
			} else {
				t = nil
			}

			if isPtr {
				tpp := (**time.Time)(ptr)
				*tpp = t
			} else {
				tp := (*time.Time)(ptr)
				if tp != nil && t != nil {
					*tp = *t
				}
			}
		}}
	}
}

type funcDecoder struct {
	fun jsoniter.DecoderFunc
}

func (decoder *funcDecoder) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	decoder.fun(ptr, iter)
}

type funcEncoder struct {
	fun         jsoniter.EncoderFunc
	isEmptyFunc func(ptr unsafe.Pointer) bool
}

func (encoder *funcEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	encoder.fun(ptr, stream)
}

func (encoder *funcEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	if encoder.isEmptyFunc == nil {
		return false
	}
	return encoder.isEmptyFunc(ptr)
}

// 时间简化格式 simpleFormat(简化的时间格式如2006-01-02 15:04:05   2006-01-02 15:04:05.000不带任何符号)
func dateTimeSimplify(datetime string, simpleFormat string) string {
	datetime = strings.TrimSuffix(datetime, "Z")
	datetime = strings.ReplaceAll(datetime, "T", " ")
	datetime = strings.Split(datetime, "+")[0]
	datetime = strings.Split(datetime, "Z")[0]
	datetime = strings.TrimSpace(datetime)

	// 始终用2006-01-02 15:04:05 decode 以保持时间精度
	//formatSplit := strings.Split(simpleFormat, ".")
	//if len(formatSplit) == 2 {
	//	formatZeroLen := len(formatSplit[1])
	//	datetimeSplit := strings.Split(datetime, ".")
	//	if len(datetimeSplit) == 2 {
	//		datetimeZeroLen := len(datetimeSplit[1])
	//		subLen := formatZeroLen - datetimeZeroLen
	//		if subLen > 0 {
	//			datetime = datetime + strings.Repeat("0", subLen)
	//		} else {
	//			datetime = datetime[:len(datetime)+subLen]
	//		}
	//	} else {
	//		datetime = datetime + "." + strings.Repeat("0", formatZeroLen)
	//	}
	//}

	return datetime
}
