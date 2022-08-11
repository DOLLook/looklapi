package utils

const (
	MinUint = 0

	MaxUint8 = ^uint8(0)
	MaxInt8  = int8(MaxUint8 >> 1)
	MinInt8  = -MaxInt8 - 1

	MaxUint16 = ^uint16(0)
	MaxInt16  = int16(MaxUint16 >> 1)
	MinInt16  = -MaxInt16 - 1

	MaxUint32 = ^uint32(0)
	MaxInt32  = int32(MaxUint32 >> 1)
	MinInt32  = -MaxInt32 - 1

	MaxUint64 = ^uint64(0)
	MaxInt64  = int64(MaxUint64 >> 1)
	MinInt64  = -MaxInt64 - 1
)

const HttpRequestHeader = "Http-Request-Header"
const HttpContextStore = "Http-Context-Store"
const ControllerRespContent = "controller-resp-content"

type EnumServiceName string

const (
	TEST_SERVICE EnumServiceName = "testService"
)
