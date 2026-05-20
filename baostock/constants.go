// Package baostock 是 baostock.com 的 Go 语言客户端实现。
// 协议基于 TCP 长连接，消息头 21 字节，消息体 zlib 压缩（仅 K 线响应）。
package baostock

// 服务端连接信息
const (
	ServerHost    = "public-api.baostock.com"
	ServerPort    = 10030
	ClientVersion = "00.9.10"
)

// 协议分隔符
const (
	MessageSplit     = "\x01"           // 消息内部字段分隔符
	AttributeSplit   = ","              // 字段名列表分隔符
	MessageDelim     = "\n"             // 消息间分隔符（发送时追加）
	EndMarker        = "<![CDATA[]]>\n" // 响应结束标志（压缩时）
	HeaderLength     = 21               // 消息头固定长度（字节）
	BodyLenWidth     = 10               // 消息头中 body 长度字段宽度
	PerPageCount     = 10000            // 每页最多条数
	DefaultStartDate = "2015-01-01"
)

// 消息类型
const (
	MsgTypeLoginReq   = "00"
	MsgTypeLoginResp  = "01"
	MsgTypeLogoutReq  = "02"
	MsgTypeLogoutResp = "03"

	MsgTypeKDataPlusReq  = "95"
	MsgTypeKDataPlusResp = "96"

	MsgTypeDividendReq  = "13"
	MsgTypeDividendResp = "14"

	MsgTypeAdjustFactorReq  = "15"
	MsgTypeAdjustFactorResp = "16"

	MsgTypeProfitReq  = "17"
	MsgTypeProfitResp = "18"

	MsgTypeOperationReq  = "19"
	MsgTypeOperationResp = "20"

	MsgTypeGrowthReq  = "21"
	MsgTypeGrowthResp = "22"

	MsgTypeDupontReq  = "23"
	MsgTypeDupontResp = "24"

	MsgTypeBalanceReq  = "25"
	MsgTypeBalanceResp = "26"

	MsgTypeCashFlowReq  = "27"
	MsgTypeCashFlowResp = "28"

	MsgTypePerfExpressReq  = "29"
	MsgTypePerfExpressResp = "30"

	MsgTypeForecastReq  = "31"
	MsgTypeForecastResp = "32"

	MsgTypeTradeDatesReq  = "33"
	MsgTypeTradeDatesResp = "34"

	MsgTypeAllStockReq  = "35"
	MsgTypeAllStockResp = "36"

	MsgTypeStockBasicReq  = "45"
	MsgTypeStockBasicResp = "46"

	MsgTypeDepositRateReq  = "47"
	MsgTypeDepositRateResp = "48"

	MsgTypeLoanRateReq  = "49"
	MsgTypeLoanRateResp = "50"

	MsgTypeRequiredReserveReq  = "51"
	MsgTypeRequiredReserveResp = "52"

	MsgTypeMoneySupplyMonthReq  = "53"
	MsgTypeMoneySupplyMonthResp = "54"

	MsgTypeMoneySupplyYearReq  = "55"
	MsgTypeMoneySupplyYearResp = "56"

	MsgTypeShiborReq  = "57"
	MsgTypeShiborResp = "58"

	MsgTypeStockIndustryReq  = "59"
	MsgTypeStockIndustryResp = "60"

	MsgTypeHS300Req  = "61"
	MsgTypeHS300Resp = "62"

	MsgTypeSZ50Req  = "63"
	MsgTypeSZ50Resp = "64"

	MsgTypeZZ500Req  = "65"
	MsgTypeZZ500Resp = "66"

	MsgTypeTerminatedReq  = "67"
	MsgTypeTerminatedResp = "68"

	MsgTypeSuspendedReq  = "69"
	MsgTypeSuspendedResp = "70"

	MsgTypeSTReq  = "71"
	MsgTypeSTResp = "72"

	MsgTypeStarSTReq  = "73"
	MsgTypeStarSTResp = "74"

	MsgTypeCPIReq  = "75"
	MsgTypeCPIResp = "76"

	MsgTypePPIReq  = "77"
	MsgTypePPIResp = "78"

	MsgTypePMIReq  = "79"
	MsgTypePMIResp = "80"

	MsgTypeConceptReq  = "81"
	MsgTypeConceptResp = "82"

	MsgTypeAreaReq  = "83"
	MsgTypeAreaResp = "84"

	MsgTypeSMEReq  = "85"
	MsgTypeSMEResp = "86"

	MsgTypeGEMReq  = "87"
	MsgTypeGEMResp = "88"

	MsgTypeSHHKReq  = "89"
	MsgTypeSHHKResp = "90"

	MsgTypeSZHKReq  = "91"
	MsgTypeSZHKResp = "92"

	MsgTypeRiskReq  = "93"
	MsgTypeRiskResp = "94"
)

// 错误码
const (
	ErrSuccess         = "0"
	ErrNoLogin         = "10001001"
	ErrUserOrPassword  = "10001002"
	ErrClientExpired   = "10001004"
	ErrLoginCountLimit = "10001005"
	ErrNeedActivate    = "10001007"
	ErrSocketErr       = "10002001"
	ErrConnectFail     = "10002002"
	ErrRecvSockFail    = "10002007"
	ErrParseData       = "10004001"
	ErrParamErr        = "10004006"
	ErrStartDateErr    = "10004007"
	ErrEndDateErr      = "10004008"
	ErrStartBigThanEnd = "10004009"
	ErrCodeInvalid     = "10004011"
	ErrUnknown         = "10004003"
)
