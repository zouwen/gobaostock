// Package baostock - history.go
// 历史 K 线数据查询接口：QueryHistoryKDataPlus
//
// 支持日线、周线、月线及 5/15/30/60 分钟线，支持前/后/不复权。
// 数据范围：1990-12-19 至当前日期。
package baostock

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// 参数常量
// ---------------------------------------------------------------------------

// K 线频率
const (
	FreqDay   = "d"  // 日线
	FreqWeek  = "w"  // 周线
	FreqMonth = "m"  // 月线
	FreqMin5  = "5"  // 5 分钟线
	FreqMin15 = "15" // 15 分钟线
	FreqMin30 = "30" // 30 分钟线
	FreqMin60 = "60" // 60 分钟线
)

// 复权类型
const (
	AdjustPost = "1" // 后复权（涨跌幅复权法）
	AdjustPre  = "2" // 前复权（涨跌幅复权法）
	AdjustNone = "3" // 不复权
)

// 日线可用的全量 fields（含估值指标）
const DailyFields = "date,code,open,high,low,close,preclose,volume,amount,adjustflag,turn,tradestatus,pctChg,peTTM,psTTM,pcfNcfTTM,pbMRQ,isST"

// 周/月线可用 fields
const WeekMonthFields = "date,code,open,high,low,close,volume,amount,adjustflag,turn,pctChg"

// 分钟线可用 fields（不含指数，不含 preclose/turn/pctChg 等）
const MinuteFields = "date,time,code,open,high,low,close,volume,amount,adjustflag"

// ---------------------------------------------------------------------------
// 强类型结构体：日线记录
// ---------------------------------------------------------------------------

// DailyBar 日线 K 线记录（含所有可查字段）。
//
// 字段说明：
//
//	Date        - 交易所行情日期，格式 YYYY-MM-DD
//	Code        - 证券代码，格式 sh.600519
//	Open        - 开盘价（精度4位小数，人民币元）
//	High        - 最高价
//	Low         - 最低价
//	Close       - 今收盘价
//	PreClose    - 昨收盘价（除权除息日为调整后价格）
//	Volume      - 成交数量（股）
//	Amount      - 成交金额（人民币元）
//	AdjustFlag  - 复权状态：1=后复权，2=前复权，3=不复权
//	Turn        - 换手率（%，精度6位小数）= 成交量/流通股 * 100%
//	TradeStatus - 交易状态：1=正常交易，0=停牌
//	PctChg      - 涨跌幅（%）= (收盘-前收盘)/前收盘 * 100%
//	PeTTM       - 滚动市盈率 = 收盘价 * 总股本 / 归母净利润TTM
//	PsTTM       - 滚动市销率 = 收盘价 * 总股本 / 营业总收入TTM
//	PcfNcfTTM   - 滚动市现率 = 收盘价 * 总股本 / 现金净增加额TTM
//	PbMRQ       - 市净率     = 总市值 / 归属股东权益（最近披露）
//	IsST        - 是否ST：1=是，0=否
type DailyBar struct {
	Date        string
	Code        string
	Open        string
	High        string
	Low         string
	Close       string
	PreClose    string
	Volume      string
	Amount      string
	AdjustFlag  string
	Turn        string
	TradeStatus string
	PctChg      string
	PeTTM       string
	PsTTM       string
	PcfNcfTTM   string
	PbMRQ       string
	IsST        string
}

// ---------------------------------------------------------------------------
// 强类型结构体：周/月线记录
// ---------------------------------------------------------------------------

// WeekMonthBar 周线/月线 K 线记录。
//
// 字段说明：
//
//	Date       - 行情日期（每周/月最后一个交易日）
//	Code       - 证券代码
//	Open       - 区间开盘价（第一个交易日开盘价）
//	High       - 区间最高价
//	Low        - 区间最低价
//	Close      - 区间收盘价（最后一个交易日收盘价）
//	Volume     - 区间成交数量（股）
//	Amount     - 区间成交金额（元）
//	AdjustFlag - 复权状态：1=后复权，2=前复权，3=不复权
//	Turn       - 区间换手率（%）
//	PctChg     - 区间涨跌幅（%）= (末收盘 - 首前收盘) / 首前收盘 * 100%
type WeekMonthBar struct {
	Date       string
	Code       string
	Open       string
	High       string
	Low        string
	Close      string
	Volume     string
	Amount     string
	AdjustFlag string
	Turn       string
	PctChg     string
}

// ---------------------------------------------------------------------------
// 强类型结构体：分钟线记录
// ---------------------------------------------------------------------------

// MinuteBar 分钟线 K 线记录（5/15/30/60 分钟线，不含指数）。
//
// 字段说明：
//
//	Date       - 交易所行情日期，格式 YYYY-MM-DD
//	Time       - 交易所行情时间，格式 YYYYMMDDHHMMSSsss
//	Code       - 证券代码
//	Open       - 开盘价
//	High       - 最高价
//	Low        - 最低价
//	Close      - 收盘价
//	Volume     - 时间范围内累计成交数量（股）
//	Amount     - 时间范围内累计成交金额（元）
//	AdjustFlag - 复权状态：1=后复权，2=前复权，3=不复权
type MinuteBar struct {
	Date       string
	Time       string
	Code       string
	Open       string
	High       string
	Low        string
	Close      string
	Volume     string
	Amount     string
	AdjustFlag string
}

// ---------------------------------------------------------------------------
// 通用结果容器（向后兼容，保留原 KRecord map 形式）
// ---------------------------------------------------------------------------

// KRecord 一条 K 线记录，字段名对应 fields 参数中指定的列。
type KRecord map[string]string

// KDataResult 历史 K 线查询结果（通用，含原始字段映射）。
type KDataResult struct {
	ErrorCode  string
	ErrorMsg   string
	Code       string
	Fields     []string
	StartDate  string
	EndDate    string
	Frequency  string
	AdjustFlag string
	Records    []KRecord
}

func (r *KDataResult) Success() bool { return r.ErrorCode == ErrSuccess }

// ---------------------------------------------------------------------------
// 强类型查询方法
// ---------------------------------------------------------------------------

// QueryDailyBars 查询日线 K 线，返回强类型 []*DailyBar。
//
//	fields 可以是 DailyFields 常量，也可以自定义子集（如 "date,open,high,low,close,volume"）
//	当自定义 fields 时，缺失的字段在结构体中为空字符串。
func (c *Client) QueryDailyBars(code, fields, startDate, endDate, adjustFlag string) ([]*DailyBar, error) {
	if fields == "" {
		fields = DailyFields
	}
	res, err := c.QueryHistoryKDataPlus(code, fields, startDate, endDate, FreqDay, adjustFlag)
	if err != nil {
		return nil, err
	}
	if !res.Success() {
		return nil, fmt.Errorf("QueryDailyBars: %s - %s", res.ErrorCode, res.ErrorMsg)
	}
	bars := make([]*DailyBar, 0, len(res.Records))
	for _, rec := range res.Records {
		bars = append(bars, &DailyBar{
			Date:        rec["date"],
			Code:        rec["code"],
			Open:        rec["open"],
			High:        rec["high"],
			Low:         rec["low"],
			Close:       rec["close"],
			PreClose:    rec["preclose"],
			Volume:      rec["volume"],
			Amount:      rec["amount"],
			AdjustFlag:  rec["adjustflag"],
			Turn:        rec["turn"],
			TradeStatus: rec["tradestatus"],
			PctChg:      rec["pctChg"],
			PeTTM:       rec["peTTM"],
			PsTTM:       rec["psTTM"],
			PcfNcfTTM:   rec["pcfNcfTTM"],
			PbMRQ:       rec["pbMRQ"],
			IsST:        rec["isST"],
		})
	}
	return bars, nil
}

// QueryWeekBars 查询周线 K 线，返回强类型 []*WeekMonthBar。
func (c *Client) QueryWeekBars(code, startDate, endDate, adjustFlag string) ([]*WeekMonthBar, error) {
	return c.queryWeekMonthBars(code, FreqWeek, startDate, endDate, adjustFlag)
}

// QueryMonthBars 查询月线 K 线，返回强类型 []*WeekMonthBar。
func (c *Client) QueryMonthBars(code, startDate, endDate, adjustFlag string) ([]*WeekMonthBar, error) {
	return c.queryWeekMonthBars(code, FreqMonth, startDate, endDate, adjustFlag)
}

func (c *Client) queryWeekMonthBars(code, freq, startDate, endDate, adjustFlag string) ([]*WeekMonthBar, error) {
	res, err := c.QueryHistoryKDataPlus(code, WeekMonthFields, startDate, endDate, freq, adjustFlag)
	if err != nil {
		return nil, err
	}
	if !res.Success() {
		return nil, fmt.Errorf("queryWeekMonthBars(%s): %s - %s", freq, res.ErrorCode, res.ErrorMsg)
	}
	bars := make([]*WeekMonthBar, 0, len(res.Records))
	for _, rec := range res.Records {
		bars = append(bars, &WeekMonthBar{
			Date:       rec["date"],
			Code:       rec["code"],
			Open:       rec["open"],
			High:       rec["high"],
			Low:        rec["low"],
			Close:      rec["close"],
			Volume:     rec["volume"],
			Amount:     rec["amount"],
			AdjustFlag: rec["adjustflag"],
			Turn:       rec["turn"],
			PctChg:     rec["pctChg"],
		})
	}
	return bars, nil
}

// QueryMinuteBars 查询分钟线 K 线，返回强类型 []*MinuteBar。
//
//	freq 只能是 FreqMin5 / FreqMin15 / FreqMin30 / FreqMin60
//	注意：分钟线不包含指数。
func (c *Client) QueryMinuteBars(code, freq, startDate, endDate, adjustFlag string) ([]*MinuteBar, error) {
	res, err := c.QueryHistoryKDataPlus(code, MinuteFields, startDate, endDate, freq, adjustFlag)
	if err != nil {
		return nil, err
	}
	if !res.Success() {
		return nil, fmt.Errorf("QueryMinuteBars(%s): %s - %s", freq, res.ErrorCode, res.ErrorMsg)
	}
	bars := make([]*MinuteBar, 0, len(res.Records))
	for _, rec := range res.Records {
		bars = append(bars, &MinuteBar{
			Date:       rec["date"],
			Time:       rec["time"],
			Code:       rec["code"],
			Open:       rec["open"],
			High:       rec["high"],
			Low:        rec["low"],
			Close:      rec["close"],
			Volume:     rec["volume"],
			Amount:     rec["amount"],
			AdjustFlag: rec["adjustflag"],
		})
	}
	return bars, nil
}

// ---------------------------------------------------------------------------
// 底层通用接口（兼容旧代码，fields 自由指定）
// ---------------------------------------------------------------------------

// QueryHistoryKDataPlus 查询历史 K 线数据（支持自动翻页，一次性返回所有数据）。
//
//	code       股票代码，格式 sh.600519 / sz.002475
//	fields     查询字段，逗号分隔，如 "date,open,high,low,close,volume,amount,pctChg"
//	           日线可用字段：见 DailyFields 常量
//	           周月线可用字段：见 WeekMonthFields 常量
//	           分钟线可用字段：见 MinuteFields 常量
//	startDate  开始日期，格式 2020-01-01，为空默认 2015-01-01
//	endDate    结束日期，格式 2026-05-18，为空默认当日
//	frequency  FreqDay/FreqWeek/FreqMonth/FreqMin5/FreqMin15/FreqMin30/FreqMin60
//	adjustFlag AdjustPost(后复权)/AdjustPre(前复权)/AdjustNone(不复权，默认)
//
// 注意：
//   - 停牌日：开高低收相同，等于前一交易日收盘价；成交量/额为 0；换手率为空字符串。
//   - 前收盘价（preclose）：除权除息日为交易所计算的除权除息价，非前日实际收盘价。
//   - 复权采用"涨跌幅复权法"，与同花顺/通达信算法不同。
func (c *Client) QueryHistoryKDataPlus(code, fields, startDate, endDate, frequency, adjustFlag string) (*KDataResult, error) {
	result := &KDataResult{}

	code = strings.ToLower(strings.TrimSpace(code))
	if len(code) != 9 {
		result.ErrorCode = ErrParamErr
		result.ErrorMsg = "股票代码应为9位，格式如 sh.600519"
		return result, nil
	}
	if fields == "" {
		result.ErrorCode = ErrParamErr
		result.ErrorMsg = "fields 不能为空"
		return result, nil
	}
	if startDate == "" {
		startDate = DefaultStartDate
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}
	if frequency == "" {
		frequency = FreqDay
	}
	if adjustFlag == "" {
		adjustFlag = AdjustNone
	}

	page := 1
	for {
		arr, err := c.queryKDataPage(page, code, fields, startDate, endDate, frequency, adjustFlag)
		if err != nil {
			return nil, err
		}

		if len(arr) < 2 {
			result.ErrorCode = ErrParseData
			result.ErrorMsg = "响应字段不足"
			return result, nil
		}

		result.ErrorCode = arr[0]
		result.ErrorMsg = arr[1]
		if !result.Success() {
			return result, nil
		}

		if page == 1 && len(arr) >= 13 {
			result.Code = arr[7]
			result.setFields(arr[8])
			result.StartDate = arr[9]
			result.EndDate = arr[10]
			result.Frequency = arr[11]
			result.AdjustFlag = arr[12]
		}

		records, err := parseKRecords(arr[6], result.Fields)
		if err != nil {
			return nil, err
		}
		result.Records = append(result.Records, records...)

		if len(records) < PerPageCount {
			break
		}
		page++
	}

	return result, nil
}

// queryKDataPage 发送单页 K 线查询请求，返回 body 字段数组。
func (c *Client) queryKDataPage(page int, code, fields, startDate, endDate, frequency, adjustFlag string) ([]string, error) {
	msgBody := "query_history_k_data_plus" + MessageSplit +
		c.userID + MessageSplit +
		strconv.Itoa(page) + MessageSplit +
		strconv.Itoa(PerPageCount) + MessageSplit +
		code + MessageSplit +
		fields + MessageSplit +
		startDate + MessageSplit +
		endDate + MessageSplit +
		frequency + MessageSplit +
		adjustFlag

	resp, err := c.sendRecv(MsgTypeKDataPlusReq, msgBody)
	if err != nil {
		return nil, err
	}
	return parseBodyArr(resp)
}

// setFields 解析字段名列表（去掉空格）。
func (r *KDataResult) setFields(raw string) {
	parts := strings.Split(raw, AttributeSplit)
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	r.Fields = parts
}

// parseKRecords 从 JSON 字符串解析 K 线记录并映射到字段名。
// JSON 格式：{"record":[["2020-01-02","977.31","992.09",...], ...]}
func parseKRecords(raw string, fields []string) ([]KRecord, error) {
	raw = strings.Join(strings.Fields(raw), "")
	if raw == "" {
		return nil, nil
	}

	var js struct {
		Record [][]string `json:"record"`
	}
	if err := json.Unmarshal([]byte(raw), &js); err != nil {
		return nil, fmt.Errorf("baostock: parse kdata json: %w", err)
	}

	records := make([]KRecord, 0, len(js.Record))
	for _, row := range js.Record {
		rec := make(KRecord, len(fields))
		for i, f := range fields {
			if i < len(row) {
				rec[f] = row[i]
			}
		}
		records = append(records, rec)
	}
	return records, nil
}
