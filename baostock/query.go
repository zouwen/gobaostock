// Package baostock - query.go
// 包含所有非K线的普通查询接口：
//   - 交易日历、股票列表、基本信息、行业分类
//   - 指数成分股（沪深300/上证50/中证500）
//   - 板块分类（ST/*ST/创业板/中小板/沪港通/深港通/概念/地域/风险警示板）
//   - 停退市股票列表
//   - 银行间同业拆放利率（Shibor）
package baostock

import (
	"fmt"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// 内部：构建消息体（所有参数用 \x01 分隔）
// ---------------------------------------------------------------------------

func buildMsgBody(params ...string) string {
	return strings.Join(params, MessageSplit)
}

func perPage() string { return fmt.Sprintf("%d", PerPageCount) }

// ---------------------------------------------------------------------------
// 通用结果类型（所有非K线接口共用）
// ---------------------------------------------------------------------------

// Result 通用查询结果（行列表形式）。
type Result struct {
	ErrorCode string
	ErrorMsg  string
	Fields    []string
	Rows      []map[string]string
}

func (r *Result) Success() bool { return r.ErrorCode == ErrSuccess }

// simpleQuery 发送非K线查询，自动读取一行响应（非压缩）。
// fieldsIdx: 响应 body_arr 中字段名列表所在的索引。
// dataIdx:   响应 body_arr 中 JSON 数据所在的索引（通常为 6）。
func (c *Client) simpleQuery(msgType, msgBody string, fieldsIdx, dataIdx int) (*Result, error) {
	resp, err := c.sendRecv(msgType, msgBody)
	if err != nil {
		return nil, err
	}
	return parseResult(resp, fieldsIdx, dataIdx)
}

func parseResult(resp string, fieldsIdx, dataIdx int) (*Result, error) {
	arr, err := parseBodyArr(resp)
	if err != nil {
		return nil, err
	}

	r := &Result{}
	if len(arr) < 2 {
		r.ErrorCode = ErrParseData
		r.ErrorMsg = "响应字段不足"
		return r, nil
	}
	r.ErrorCode = arr[0]
	r.ErrorMsg = arr[1]
	if !r.Success() {
		return r, nil
	}

	if len(arr) <= fieldsIdx || len(arr) <= dataIdx {
		return r, nil
	}

	// 解析字段名列表
	fieldParts := strings.Split(arr[fieldsIdx], AttributeSplit)
	for i, f := range fieldParts {
		fieldParts[i] = strings.TrimSpace(f)
	}
	r.Fields = fieldParts

	// 解析数据
	records, err := parseKRecords(arr[dataIdx], r.Fields)
	if err != nil {
		return nil, err
	}
	for _, rec := range records {
		r.Rows = append(r.Rows, map[string]string(rec))
	}
	return r, nil
}

// ---------------------------------------------------------------------------
// 交易日历
// ---------------------------------------------------------------------------

// TradeDateRecord 交易日历记录。
//
// 字段说明：
//
//	CalendarDate - 日期，格式 YYYY-MM-DD
//	IsTradingDay - 是否交易日："1"=交易日，"0"=非交易日（周末、节假日、休市日）
type TradeDateRecord struct {
	CalendarDate string // 日期，格式 YYYY-MM-DD
	IsTradingDay string // 是否交易日："1"=交易日，"0"=非交易日
}

// IsTradingDayBool 返回 bool 类型的交易日标志（便于程序判断）。
func (r *TradeDateRecord) IsTradingDayBool() bool {
	return r.IsTradingDay == "1"
}

// QueryTradeDates 查询指定日期范围内的交易日历。
//
// 参数说明：
//
//	startDate - 开始日期（含），格式 YYYY-MM-DD；为空时取 2015-01-01
//	endDate   - 结束日期（含），格式 YYYY-MM-DD；为空时取当日
//
// 返回结果包含范围内的每一个自然日（含非交易日），通过 IsTradingDay 字段区分。
func (c *Client) QueryTradeDates(startDate, endDate string) ([]*TradeDateRecord, error) {
	if startDate == "" {
		startDate = DefaultStartDate
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}
	msgBody := buildMsgBody("query_trade_dates", c.userID, "1", perPage(), startDate, endDate)
	r, err := c.simpleQuery(MsgTypeTradeDatesReq, msgBody, 9, 6)
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("QueryTradeDates: %s - %s", r.ErrorCode, r.ErrorMsg)
	}
	records := make([]*TradeDateRecord, 0, len(r.Rows))
	for _, row := range r.Rows {
		records = append(records, &TradeDateRecord{
			CalendarDate: row["calendar_date"],
			IsTradingDay: row["is_trading_day"],
		})
	}
	return records, nil
}

// ---------------------------------------------------------------------------
// 所有证券列表
// ---------------------------------------------------------------------------

// StockRecord 某日所有证券信息记录。
//
// 字段说明：
//
//	Code        - 证券代码，格式 sh.600000（sh=上海，sz=深圳）
//	TradeStatus - 交易状态："1"=正常交易，"0"=停牌
//	CodeName    - 证券名称（如 浦发银行、*ST柳化）
type StockRecord struct {
	Code        string // 证券代码，格式 sh.600519
	TradeStatus string // 交易状态："1"=正常交易，"0"=停牌
	CodeName    string // 证券名称
}

// IsTrading 返回该证券当日是否正常交易（非停牌）。
func (r *StockRecord) IsTrading() bool {
	return r.TradeStatus == "1"
}

// QueryAllStock 查询指定交易日沪深市场的所有证券信息。
//
// 参数说明：
//
//	date - 查询日期，格式 YYYY-MM-DD；为空时取当日
//	       注意：闭市后日K线数据更新后，该接口才会返回当天数据，否则返回空。
//
// 返回包含当日全部上市证券（含停牌），通过 TradeStatus 字段区分交易状态。
func (c *Client) QueryAllStock(date string) ([]*StockRecord, error) {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	msgBody := buildMsgBody("query_all_stock", c.userID, "1", perPage(), date)
	r, err := c.simpleQuery(MsgTypeAllStockReq, msgBody, 8, 6)
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("QueryAllStock: %s - %s", r.ErrorCode, r.ErrorMsg)
	}
	stocks := make([]*StockRecord, 0, len(r.Rows))
	for _, row := range r.Rows {
		stocks = append(stocks, &StockRecord{
			Code:        row["code"],
			TradeStatus: row["tradeStatus"],
			CodeName:    row["code_name"],
		})
	}
	return stocks, nil
}

// ---------------------------------------------------------------------------
// 证券基本资料
// ---------------------------------------------------------------------------

// StockBasic 证券基本资料。
//
// 字段说明：
//
//	Code      - 证券代码，格式 sh.600519
//	CodeName  - 证券名称
//	IPODate   - 上市日期，格式 YYYY-MM-DD
//	OutDate   - 退市日期，格式 YYYY-MM-DD；尚未退市则为空
//	StockType - 证券类型："1"=股票，"2"=指数，"3"=其他
//	Status    - 上市状态："1"=上市，"2"=退市，"3"=暂停上市
//	ExchSrq   - 是否为沪/深股通标的："sh"=沪股通，"sz"=深股通，"0"=非沪深股通
//	IsSt      - 是否ST："1"=是，"0"=否
type StockBasic struct {
	Code      string // 证券代码，格式 sh.600519
	CodeName  string // 证券名称
	IPODate   string // 上市日期，格式 YYYY-MM-DD
	OutDate   string // 退市日期，格式 YYYY-MM-DD；未退市为空
	StockType string // 证券类型："1"=股票，"2"=指数，"3"=其他
	Status    string // 上市状态："1"=上市，"2"=退市，"3"=暂停上市
	ExchSrq   string // 沪深股通："sh"=沪股通，"sz"=深股通，"0"=非
	IsSt      string // 是否ST："1"=是，"0"=否
}

// QueryStockBasic 查询证券基本资料。
// code 格式 sh.600519，为空则查全部；codeName 支持模糊匹配。
func (c *Client) QueryStockBasic(code, codeName string) ([]*StockBasic, error) {
	code = strings.ToLower(strings.TrimSpace(code))
	// response: arr[9]=fields
	msgBody := buildMsgBody("query_stock_basic", c.userID, "1", perPage(), code, codeName)
	r, err := c.simpleQuery(MsgTypeStockBasicReq, msgBody, 9, 6)
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("QueryStockBasic: %s - %s", r.ErrorCode, r.ErrorMsg)
	}
	basics := make([]*StockBasic, 0, len(r.Rows))
	for _, row := range r.Rows {
		basics = append(basics, &StockBasic{
			Code:      row["code"],
			CodeName:  row["code_name"],
			IPODate:   row["ipoDate"],
			OutDate:   row["outDate"],
			StockType: row["stock_type"],
			Status:    row["status"],
			ExchSrq:   row["exchSrq"],
			IsSt:      row["is_st"],
		})
	}
	return basics, nil
}

// ---------------------------------------------------------------------------
// 行业分类
// ---------------------------------------------------------------------------

// StockIndustry 行业分类记录。
type StockIndustry struct {
	UpdateDate             string
	Code                   string
	CodeName               string
	Industry               string
	IndustryClassification string
}

// QueryStockIndustry 查询股票行业分类信息（申万一级行业）。
// code 格式 sh.600519；date 格式 2024-01-01，为空查最新。
func (c *Client) QueryStockIndustry(code, date string) ([]*StockIndustry, error) {
	code = strings.ToLower(strings.TrimSpace(code))
	// response: arr[9]=fields（code + date）
	msgBody := buildMsgBody("query_stock_industry", c.userID, "1", perPage(), code, date)
	r, err := c.simpleQuery(MsgTypeStockIndustryReq, msgBody, 9, 6)
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("QueryStockIndustry: %s - %s", r.ErrorCode, r.ErrorMsg)
	}
	list := make([]*StockIndustry, 0, len(r.Rows))
	for _, row := range r.Rows {
		list = append(list, &StockIndustry{
			UpdateDate:             row["updateDate"],
			Code:                   row["code"],
			CodeName:               row["code_name"],
			Industry:               row["industry"],
			IndustryClassification: row["industryClassification"],
		})
	}
	return list, nil
}

// ---------------------------------------------------------------------------
// 指数成分股（内部通用）
// ---------------------------------------------------------------------------

// IndexStock 指数成分股记录。
type IndexStock struct {
	Date        string
	Code        string
	DisplayName string
	CodeName    string
}

// queryIndexStocks 内部通用：查询某个指数成分股。
// response: arr[8]=fields, arr[7]=date
func (c *Client) queryIndexStocks(msgType, methodName, date string) ([]*IndexStock, error) {
	msgBody := buildMsgBody(methodName, c.userID, "1", perPage(), date)
	r, err := c.simpleQuery(msgType, msgBody, 8, 6)
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("%s: %s - %s", methodName, r.ErrorCode, r.ErrorMsg)
	}
	stocks := make([]*IndexStock, 0, len(r.Rows))
	for _, row := range r.Rows {
		stocks = append(stocks, &IndexStock{
			Date:        row["date"],
			Code:        row["code"],
			DisplayName: row["display_name"],
			CodeName:    row["code_name"],
		})
	}
	return stocks, nil
}

// QueryHS300Stocks 查询沪深300成分股。date 为空则查最新。
func (c *Client) QueryHS300Stocks(date string) ([]*IndexStock, error) {
	return c.queryIndexStocks(MsgTypeHS300Req, "query_hs300_stocks", date)
}

// QuerySZ50Stocks 查询上证50成分股。
func (c *Client) QuerySZ50Stocks(date string) ([]*IndexStock, error) {
	return c.queryIndexStocks(MsgTypeSZ50Req, "query_sz50_stocks", date)
}

// QueryZZ500Stocks 查询中证500成分股。
func (c *Client) QueryZZ500Stocks(date string) ([]*IndexStock, error) {
	return c.queryIndexStocks(MsgTypeZZ500Req, "query_zz500_stocks", date)
}

// ---------------------------------------------------------------------------
// 停退市股票
// ---------------------------------------------------------------------------

// SuspendedStock 停退市股票记录（通用）。
type SuspendedStock struct {
	Code     string
	CodeName string
	Date     string
}

// queryStockList 内部：查询各类股票名单（停市/退市/ST/*ST 等）。
// response: arr[8]=fields
func (c *Client) queryStockList(msgType, methodName, date string) ([]*SuspendedStock, error) {
	msgBody := buildMsgBody(methodName, c.userID, "1", perPage(), date)
	r, err := c.simpleQuery(msgType, msgBody, 8, 6)
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("%s: %s - %s", methodName, r.ErrorCode, r.ErrorMsg)
	}
	list := make([]*SuspendedStock, 0, len(r.Rows))
	for _, row := range r.Rows {
		list = append(list, &SuspendedStock{
			Code:     row["code"],
			CodeName: row["code_name"],
			Date:     row["date"],
		})
	}
	return list, nil
}

// QueryTerminatedStocks 查询退市股票列表。
func (c *Client) QueryTerminatedStocks(date string) ([]*SuspendedStock, error) {
	return c.queryStockList(MsgTypeTerminatedReq, "query_terminated_stocks", date)
}

// QuerySuspendedStocks 查询暂停上市股票列表。
func (c *Client) QuerySuspendedStocks(date string) ([]*SuspendedStock, error) {
	return c.queryStockList(MsgTypeSuspendedReq, "query_suspended_stocks", date)
}

// QueryStStocks 查询ST股票列表。
func (c *Client) QueryStStocks(date string) ([]*SuspendedStock, error) {
	return c.queryStockList(MsgTypeSTReq, "query_st_stocks", date)
}

// QueryStarStStocks 查询*ST股票列表。
func (c *Client) QueryStarStStocks(date string) ([]*SuspendedStock, error) {
	return c.queryStockList(MsgTypeStarSTReq, "query_starst_stocks", date)
}

// ---------------------------------------------------------------------------
// 板块分类
// ---------------------------------------------------------------------------

// SectorStock 板块成分股记录。
type SectorStock struct {
	Date        string
	Code        string
	CodeName    string
	DisplayName string
}

// querySectorStocks 内部：查询各类板块成分股。
// response: arr[8]=fields
func (c *Client) querySectorStocks(msgType, methodName, date string) ([]*SectorStock, error) {
	msgBody := buildMsgBody(methodName, c.userID, "1", perPage(), date)
	r, err := c.simpleQuery(msgType, msgBody, 8, 6)
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("%s: %s - %s", methodName, r.ErrorCode, r.ErrorMsg)
	}
	list := make([]*SectorStock, 0, len(r.Rows))
	for _, row := range r.Rows {
		list = append(list, &SectorStock{
			Date:        row["date"],
			Code:        row["code"],
			CodeName:    row["code_name"],
			DisplayName: row["display_name"],
		})
	}
	return list, nil
}

// QueryGEMStocks 查询创业板（GEM）成分股。
func (c *Client) QueryGEMStocks(date string) ([]*SectorStock, error) {
	return c.querySectorStocks(MsgTypeGEMReq, "query_gem_stocks", date)
}

// QuerySMEStocks 查询中小板（SME）成分股。
func (c *Client) QuerySMEStocks(date string) ([]*SectorStock, error) {
	return c.querySectorStocks(MsgTypeSMEReq, "query_sme_stocks", date)
}

// QuerySHHKStocks 查询沪港通成分股。
func (c *Client) QuerySHHKStocks(date string) ([]*SectorStock, error) {
	return c.querySectorStocks(MsgTypeSHHKReq, "query_shhk_stocks", date)
}

// QuerySZHKStocks 查询深港通成分股。
func (c *Client) QuerySZHKStocks(date string) ([]*SectorStock, error) {
	return c.querySectorStocks(MsgTypeSZHKReq, "query_szhk_stocks", date)
}

// QueryStockInRisk 查询风险警示板股票。
func (c *Client) QueryStockInRisk(date string) ([]*SectorStock, error) {
	return c.querySectorStocks(MsgTypeRiskReq, "query_stock_in_risk", date)
}

// ---------------------------------------------------------------------------
// 银行间同业拆放利率（Shibor）
// ---------------------------------------------------------------------------

// ShiborRecord Shibor 记录。
type ShiborRecord struct {
	Date        string
	Overnight   string // 隔夜
	OneWeek     string // 一周
	TwoWeeks    string // 两周
	OneMonth    string // 一个月
	ThreeMonths string // 三个月
	SixMonths   string // 六个月
	NineMonths  string // 九个月
	OneYear     string // 一年
}

// QueryShiborData 查询银行间同业拆放利率（Shibor）。
// date 格式 2024-01-01，为空则查最新。
func (c *Client) QueryShiborData(date string) ([]*ShiborRecord, error) {
	msgBody := buildMsgBody("query_shibor_data", c.userID, "1", perPage(), date)
	r, err := c.simpleQuery(MsgTypeShiborReq, msgBody, 8, 6)
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("QueryShiborData: %s - %s", r.ErrorCode, r.ErrorMsg)
	}
	list := make([]*ShiborRecord, 0, len(r.Rows))
	for _, row := range r.Rows {
		list = append(list, &ShiborRecord{
			Date:        row["date"],
			Overnight:   row["ON"],
			OneWeek:     row["1W"],
			TwoWeeks:    row["2W"],
			OneMonth:    row["1M"],
			ThreeMonths: row["3M"],
			SixMonths:   row["6M"],
			NineMonths:  row["9M"],
			OneYear:     row["1Y"],
		})
	}
	return list, nil
}

// ---------------------------------------------------------------------------
// 概念板块分类
// ---------------------------------------------------------------------------

// ConceptStock 概念板块成分股记录。
type ConceptStock struct {
	UpdateDate  string
	ConceptCode string // 概念编码
	ConceptName string // 概念名称
	Code        string // 股票代码
	CodeName    string // 股票名称
}

// QueryStockConcept 查询概念板块成分股信息。
// date 格式 2024-01-01，为空则查最新。
func (c *Client) QueryStockConcept(date string) ([]*ConceptStock, error) {
	msgBody := buildMsgBody("query_stock_concept", c.userID, "1", perPage(), date)
	// response: arr[7]=date, arr[8]=fields
	r, err := c.simpleQuery(MsgTypeConceptReq, msgBody, 8, 6)
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("QueryStockConcept: %s - %s", r.ErrorCode, r.ErrorMsg)
	}
	list := make([]*ConceptStock, 0, len(r.Rows))
	for _, row := range r.Rows {
		list = append(list, &ConceptStock{
			UpdateDate:  row["updateDate"],
			ConceptCode: row["conceptCode"],
			ConceptName: row["conceptName"],
			Code:        row["code"],
			CodeName:    row["codeName"],
		})
	}
	return list, nil
}

// ---------------------------------------------------------------------------
// 地域板块分类
// ---------------------------------------------------------------------------

// AreaStock 地域板块成分股记录。
type AreaStock struct {
	UpdateDate string
	AreaCode   string // 地域编码
	AreaName   string // 地域名称
	Code       string // 股票代码
	CodeName   string // 股票名称
}

// QueryStockArea 查询地域板块成分股信息。
// date 格式 2024-01-01，为空则查最新。
func (c *Client) QueryStockArea(date string) ([]*AreaStock, error) {
	msgBody := buildMsgBody("query_stock_area", c.userID, "1", perPage(), date)
	// response: arr[7]=date, arr[8]=fields
	r, err := c.simpleQuery(MsgTypeAreaReq, msgBody, 8, 6)
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("QueryStockArea: %s - %s", r.ErrorCode, r.ErrorMsg)
	}
	list := make([]*AreaStock, 0, len(r.Rows))
	for _, row := range r.Rows {
		list = append(list, &AreaStock{
			UpdateDate: row["updateDate"],
			AreaCode:   row["areaCode"],
			AreaName:   row["areaName"],
			Code:       row["code"],
			CodeName:   row["codeName"],
		})
	}
	return list, nil
}
