// Package baostock - corpreport.go
// 公司公告类数据：
//   - 业绩快报（QueryPerformanceExpressReport）
//   - 业绩预告（QueryForecastReport）
package baostock

import (
	"fmt"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// 业绩快报
// ---------------------------------------------------------------------------

// PerformanceExpressRecord 业绩快报记录。
type PerformanceExpressRecord struct {
	Code                  string
	PubDate               string // 公告日期
	StatDate              string // 统计截止日期
	TotalAssets           string // 总资产
	TotalShareHolderEquity string // 净资产
	OperatingIncome       string // 营业收入
	NetProfit             string // 净利润
	EPS                   string // 每股收益
	WeightedAvgROE        string // 净资产收益率（加权）
	YOYNetProfit          string // 净利润同比增长率
	YOYOperIncome         string // 营业收入同比增长率
}

// QueryPerformanceExpressReport 查询公司业绩快报。
// response: arr[7]=code, arr[8]=startDate, arr[9]=endDate, arr[10]=fields
func (c *Client) QueryPerformanceExpressReport(code, startDate, endDate string) ([]*PerformanceExpressRecord, error) {
	code = strings.ToLower(strings.TrimSpace(code))
	if startDate == "" {
		startDate = DefaultStartDate
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}
	msgBody := buildMsgBody("query_performance_express_report", c.userID, "1", perPage(), code, startDate, endDate)
	r, err := c.simpleQuery(MsgTypePerfExpressReq, msgBody, 10, 6)
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("QueryPerformanceExpressReport: %s - %s", r.ErrorCode, r.ErrorMsg)
	}
	list := make([]*PerformanceExpressRecord, 0, len(r.Rows))
	for _, row := range r.Rows {
		list = append(list, &PerformanceExpressRecord{
			Code:                   row["code"],
			PubDate:                row["pubDate"],
			StatDate:               row["statDate"],
			TotalAssets:            row["toTalAssets"],
			TotalShareHolderEquity: row["totalShareHolderEquity"],
			OperatingIncome:        row["operatingIncome"],
			NetProfit:              row["netProfit"],
			EPS:                    row["EPS"],
			WeightedAvgROE:         row["weightedAvgROE"],
			YOYNetProfit:           row["YOYNI"],
			YOYOperIncome:          row["YOYOperIncome"],
		})
	}
	return list, nil
}

// ---------------------------------------------------------------------------
// 业绩预告
// ---------------------------------------------------------------------------

// ForecastReport 业绩预告记录。
type ForecastReport struct {
	Code                string
	PubDate             string // 公告日期
	StatDate            string // 统计截止日期（季度末）
	Type                string // 预告类型
	ReportPeriod        string // 报告期
	NetProfitMin        string // 净利润变动幅度下限
	NetProfitMax        string // 净利润变动幅度上限
	LastYearNetProfit   string // 上年同期净利润
	FirstHalfNetProfit  string // 本报告期上半年净利润
	NetProfitMinOfYear  string // 全年净利润变动幅度下限
	NetProfitMaxOfYear  string // 全年净利润变动幅度上限
	LastYearNetProfitOfYear string // 上年全年净利润
}

// QueryForecastReport 查询公司业绩预告。
// response: arr[7]=code, arr[8]=startDate, arr[9]=endDate, arr[10]=fields
func (c *Client) QueryForecastReport(code, startDate, endDate string) ([]*ForecastReport, error) {
	code = strings.ToLower(strings.TrimSpace(code))
	if startDate == "" {
		startDate = DefaultStartDate
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}
	msgBody := buildMsgBody("query_forecast_report", c.userID, "1", perPage(), code, startDate, endDate)
	r, err := c.simpleQuery(MsgTypeForecastReq, msgBody, 10, 6)
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("QueryForecastReport: %s - %s", r.ErrorCode, r.ErrorMsg)
	}
	list := make([]*ForecastReport, 0, len(r.Rows))
	for _, row := range r.Rows {
		list = append(list, &ForecastReport{
			Code:                    row["code"],
			PubDate:                 row["pubDate"],
			StatDate:                row["statDate"],
			Type:                    row["type"],
			ReportPeriod:            row["reportPeriod"],
			NetProfitMin:            row["netProfitMin"],
			NetProfitMax:            row["netProfitMax"],
			LastYearNetProfit:       row["lastYearNetProfit"],
			FirstHalfNetProfit:      row["firstHalfNetProfit"],
			NetProfitMinOfYear:      row["netProfitMinOfYear"],
			NetProfitMaxOfYear:      row["netProfitMaxOfYear"],
			LastYearNetProfitOfYear: row["lastYearNetProfitOfYear"],
		})
	}
	return list, nil
}
