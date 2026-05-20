// Package baostock - evaluation.go
// 季频财务评估指标：
//   - 股息分红（QueryDividendData）
//   - 复权因子（QueryAdjustFactor）
//   - 盈利能力（QueryProfitData）
//   - 营运能力（QueryOperationData）
//   - 成长能力（QueryGrowthData）
//   - 杜邦指数（QueryDupontData）
//   - 偿债能力（QueryBalanceData）
//   - 现金流量（QueryCashFlowData）
package baostock

import (
	"fmt"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// 内部：季频查询（code + year + quarter）
// response: arr[7]=code, arr[8]=year, arr[9]=quarter, arr[10]=fields
// ---------------------------------------------------------------------------

func (c *Client) quarterQuery(msgType, method, code, year, quarter string) (*Result, error) {
	code = strings.ToLower(strings.TrimSpace(code))
	if year == "" {
		year = time.Now().Format("2006")[:4]
	}
	if quarter == "" {
		m := int(time.Now().Month())
		quarter = fmt.Sprintf("%d", (m+2)/3)
	}
	msgBody := buildMsgBody(method, c.userID, "1", perPage(), code, year, quarter)
	// fieldsIdx=10, dataIdx=6
	return c.simpleQuery(msgType, msgBody, 10, 6)
}

// ---------------------------------------------------------------------------
// 股息分红
// ---------------------------------------------------------------------------

// DividendRecord 股息分红记录。
type DividendRecord struct {
	Code           string
	DividPrenoticeDate string
	DividSchemDate string
	DividRegistDate string
	DividExDate    string
	DividPayDate   string
	DividStockMarketClose string
	DividCashPS    string // 每股股息（税前）
	DividTaxMarketClose string
	DividPrecondition string
	DividSurplusShareRate string
	DividReserveToShare string
	DividSharesRatio string
	Year           string
	YearType       string
}

// QueryDividendData 查询股息分红数据。
// yearType: "report"=预案公告年份（默认）, "operate"=除权除息年份。
func (c *Client) QueryDividendData(code, year, yearType string) ([]*DividendRecord, error) {
	code = strings.ToLower(strings.TrimSpace(code))
	if year == "" {
		year = time.Now().Format("2006")[:4]
	}
	if yearType == "" {
		yearType = "report"
	}
	// response: arr[7]=code, arr[8]=year, arr[9]=yearType, arr[10]=fields
	msgBody := buildMsgBody("query_dividend_data", c.userID, "1", perPage(), code, year, yearType)
	r, err := c.simpleQuery(MsgTypeDividendReq, msgBody, 10, 6)
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("QueryDividendData: %s - %s", r.ErrorCode, r.ErrorMsg)
	}
	list := make([]*DividendRecord, 0, len(r.Rows))
	for _, row := range r.Rows {
		list = append(list, &DividendRecord{
			Code:                  row["code"],
			DividPrenoticeDate:    row["dividPrenoticeDate"],
			DividSchemDate:        row["dividSchemDate"],
			DividRegistDate:       row["dividRegistDate"],
			DividExDate:           row["dividExDate"],
			DividPayDate:          row["dividPayDate"],
			DividStockMarketClose: row["dividStockMarketClose"],
			DividCashPS:           row["dividCashPs"],
			DividTaxMarketClose:   row["dividTaxMarketClose"],
			DividPrecondition:     row["dividPrecondition"],
			DividSurplusShareRate: row["dividSurplusShareRate"],
			DividReserveToShare:   row["dividReserveToShare"],
			DividSharesRatio:      row["dividSharesRatio"],
			Year:                  row["dividYear"],
			YearType:              row["dividYearType"],
		})
	}
	return list, nil
}

// ---------------------------------------------------------------------------
// 复权因子
// ---------------------------------------------------------------------------

// AdjustFactor 复权因子记录。
type AdjustFactor struct {
	DivDate       string
	FwdAdjFactor  string // 前复权因子
	BackAdjFactor string // 后复权因子
}

// QueryAdjustFactor 查询复权因子数据。
// response: arr[7]=code, arr[8]=startDate, arr[9]=endDate, arr[10]=fields
func (c *Client) QueryAdjustFactor(code, startDate, endDate string) ([]*AdjustFactor, error) {
	code = strings.ToLower(strings.TrimSpace(code))
	if startDate == "" {
		startDate = DefaultStartDate
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}
	msgBody := buildMsgBody("query_adjust_factor", c.userID, "1", perPage(), code, startDate, endDate)
	r, err := c.simpleQuery(MsgTypeAdjustFactorReq, msgBody, 10, 6)
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("QueryAdjustFactor: %s - %s", r.ErrorCode, r.ErrorMsg)
	}
	list := make([]*AdjustFactor, 0, len(r.Rows))
	for _, row := range r.Rows {
		list = append(list, &AdjustFactor{
			DivDate:       row["divDate"],
			FwdAdjFactor:  row["foreAdjustFactor"],
			BackAdjFactor: row["backAdjustFactor"],
		})
	}
	return list, nil
}

// ---------------------------------------------------------------------------
// 季频财务指标（盈利/营运/成长/杜邦/偿债/现金流）
// 所有接口响应格式相同: arr[7]=code, arr[8]=year, arr[9]=quarter, arr[10]=fields
// ---------------------------------------------------------------------------

// FinancialRecord 季频财务指标记录（通用，字段名随 API 不同而异）。
type FinancialRecord struct {
	Code    string
	Year    string
	Quarter string
	Fields  map[string]string // 原始字段名 -> 值
}

func (c *Client) quarterFinancial(msgType, method, code, year, quarter string) ([]*FinancialRecord, error) {
	r, err := c.quarterQuery(msgType, method, code, year, quarter)
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("%s: %s - %s", method, r.ErrorCode, r.ErrorMsg)
	}
	list := make([]*FinancialRecord, 0, len(r.Rows))
	for _, row := range r.Rows {
		rec := &FinancialRecord{
			Code:    row["code"],
			Year:    row["pubDate"],
			Quarter: row["statDate"],
			Fields:  row,
		}
		list = append(list, rec)
	}
	return list, nil
}

// QueryProfitData 查询季频盈利能力数据。
// 字段：roeAvg(净资产收益率), npMargin(销售净利率), gpMargin(销售毛利率),
//
//	netProfit(净利润), epsTTM(每股收益TTM), MBRevenue(主营业务收入), totalShare(总股本),
//	liqaShare(流通股本)
func (c *Client) QueryProfitData(code, year, quarter string) ([]*FinancialRecord, error) {
	return c.quarterFinancial(MsgTypeProfitReq, "query_profit_data", code, year, quarter)
}

// QueryOperationData 查询季频营运能力数据。
// 字段：NRTurnRatio(应收账款周转率), NRTurnDays(应收账款周转天数),
//
//	INVTurnRatio(存货周转率), INVTurnDays(存货周转天数),
//	CATurnRatio(流动资产周转率), AssetTurnRatio(总资产周转率)
func (c *Client) QueryOperationData(code, year, quarter string) ([]*FinancialRecord, error) {
	return c.quarterFinancial(MsgTypeOperationReq, "query_operation_data", code, year, quarter)
}

// QueryGrowthData 查询季频成长能力数据。
// 字段：YOYEquity(净资产同比增长率), YOYAsset(总资产同比增长率),
//
//	YOYNI(净利润同比增长率), YOYEPSBasic(基本每股收益同比增长率),
//	YOYPNI(归母净利润同比增长率)
func (c *Client) QueryGrowthData(code, year, quarter string) ([]*FinancialRecord, error) {
	return c.quarterFinancial(MsgTypeGrowthReq, "query_growth_data", code, year, quarter)
}

// QueryDupontData 查询季频杜邦指数数据。
// 字段：dupontROE(净资产收益率), dupontAssetStoequity(权益乘数),
//
//	dupontEquityMultiplier(权益乘数倒数), dupontStockholdersEquity(归属母公司股东权益),
//	dupontNetprofit(净利润), dupontNI(净利润/净收入), dupontTax(净利润/税前利润),
//	dupontInterestBurden(税前利润/EBIT), dupontEbos(EBIT/净收入),
//	dupontAssetTurnover(净收入/总资产), dupontNETA(总资产/归属母公司股东权益)
func (c *Client) QueryDupontData(code, year, quarter string) ([]*FinancialRecord, error) {
	return c.quarterFinancial(MsgTypeDupontReq, "query_dupont_data", code, year, quarter)
}

// QueryBalanceData 查询季频偿债能力数据。
// 字段：currentRatio(流动比率), quickRatio(速动比率),
//
//	cashRatio(现金比率), YOYLiability(总负债同比增长率),
//	liabilityToAsset(资产负债率), assetToEquity(权益乘数)
func (c *Client) QueryBalanceData(code, year, quarter string) ([]*FinancialRecord, error) {
	return c.quarterFinancial(MsgTypeBalanceReq, "query_balance_data", code, year, quarter)
}

// QueryCashFlowData 查询季频现金流量数据。
// 字段：CAToAsset(流动资产除以总资产), NCAToAsset(非流动资产除以总资产),
//
//	tangibleAssetToAsset(有形资产除以总资产), ebitToInterest(EBIT/利息费用),
//	operatingNIToTotalLiability(经营活动净收益/负债合计),
//	endCash(期末现金及现金等价物余额)
func (c *Client) QueryCashFlowData(code, year, quarter string) ([]*FinancialRecord, error) {
	return c.quarterFinancial(MsgTypeCashFlowReq, "query_cash_flow_data", code, year, quarter)
}
