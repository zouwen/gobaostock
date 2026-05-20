// Package baostock - macroscopic.go
// 宏观经济数据：
//   - 存款利率（QueryDepositRateData）
//   - 贷款利率（QueryLoanRateData）
//   - 存款准备金率（QueryRequiredReserveRatioData）
//   - 货币供应量月度（QueryMoneySupplyDataMonth）
//   - 货币供应量年度（QueryMoneySupplyDataYear）
//   - 居民消费价格指数 CPI（QueryCPIData）
//   - 工业品出厂价格指数 PPI（QueryPPIData）
//   - 采购经理人指数 PMI（QueryPMIData）
package baostock

import "fmt"

// ---------------------------------------------------------------------------
// 内部：宏观数据通用查询（startDate + endDate）
// response: arr[7]=startDate, arr[8]=endDate, arr[9]=fields
// ---------------------------------------------------------------------------

func (c *Client) macroDateQuery(msgType, method, startDate, endDate string) (*Result, error) {
	msgBody := buildMsgBody(method, c.userID, "1", perPage(), startDate, endDate)
	return c.simpleQuery(msgType, msgBody, 9, 6)
}

// ---------------------------------------------------------------------------
// 存款利率
// ---------------------------------------------------------------------------

// DepositRateRecord 存款利率记录。
type DepositRateRecord struct {
	Date         string // 利率调整日期
	DepositRateShort string // 活期存款利率
	DepositRateMidShort string // 三个月整存整取
	DepositRateMid      string // 半年整存整取
	DepositRateMidLong  string // 一年整存整取
	DepositRateLong     string // 二年整存整取
}

// QueryDepositRateData 查询存款利率数据。
// 日期格式：2024-01-01；为空则查全部。
func (c *Client) QueryDepositRateData(startDate, endDate string) ([]*DepositRateRecord, error) {
	r, err := c.macroDateQuery(MsgTypeDepositRateReq, "query_deposit_rate_data", startDate, endDate)
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("QueryDepositRateData: %s - %s", r.ErrorCode, r.ErrorMsg)
	}
	list := make([]*DepositRateRecord, 0, len(r.Rows))
	for _, row := range r.Rows {
		list = append(list, &DepositRateRecord{
			Date:                row["date"],
			DepositRateShort:    row["depositRateShort"],
			DepositRateMidShort: row["depositRateMidShort"],
			DepositRateMid:      row["depositRateMid"],
			DepositRateMidLong:  row["depositRateMidLong"],
			DepositRateLong:     row["depositRateLong"],
		})
	}
	return list, nil
}

// ---------------------------------------------------------------------------
// 贷款利率
// ---------------------------------------------------------------------------

// LoanRateRecord 贷款利率记录。
type LoanRateRecord struct {
	Date              string
	LoanRateShort     string // 六个月以内贷款利率
	LoanRateMidShort  string // 六个月至一年贷款利率
	LoanRateMid       string // 一至三年贷款利率
	LoanRateMidLong   string // 三至五年贷款利率
	LoanRateLong      string // 五年以上贷款利率
}

// QueryLoanRateData 查询贷款利率数据。
func (c *Client) QueryLoanRateData(startDate, endDate string) ([]*LoanRateRecord, error) {
	r, err := c.macroDateQuery(MsgTypeLoanRateReq, "query_loan_rate_data", startDate, endDate)
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("QueryLoanRateData: %s - %s", r.ErrorCode, r.ErrorMsg)
	}
	list := make([]*LoanRateRecord, 0, len(r.Rows))
	for _, row := range r.Rows {
		list = append(list, &LoanRateRecord{
			Date:             row["date"],
			LoanRateShort:    row["loanRateShort"],
			LoanRateMidShort: row["loanRateMidShort"],
			LoanRateMid:      row["loanRateMid"],
			LoanRateMidLong:  row["loanRateMidLong"],
			LoanRateLong:     row["loanRateLong"],
		})
	}
	return list, nil
}

// ---------------------------------------------------------------------------
// 存款准备金率
// ---------------------------------------------------------------------------

// RequiredReserveRecord 存款准备金率记录。
type RequiredReserveRecord struct {
	Date              string
	BusinessBankRate  string // 大型金融机构存款准备金率
	OtherBankRate     string // 小型金融机构存款准备金率
}

// QueryRequiredReserveRatioData 查询存款准备金率数据。
// yearType: "0"=公告日期（默认）, "1"=生效日期
func (c *Client) QueryRequiredReserveRatioData(startDate, endDate, yearType string) ([]*RequiredReserveRecord, error) {
	if yearType == "" {
		yearType = "0"
	}
	// response: arr[7]=startDate, arr[8]=endDate, arr[9]=yearType, arr[10]=fields
	msgBody := buildMsgBody("query_required_reserve_ratio_data", c.userID, "1", perPage(), startDate, endDate, yearType)
	r, err := c.simpleQuery(MsgTypeRequiredReserveReq, msgBody, 10, 6)
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("QueryRequiredReserveRatioData: %s - %s", r.ErrorCode, r.ErrorMsg)
	}
	list := make([]*RequiredReserveRecord, 0, len(r.Rows))
	for _, row := range r.Rows {
		list = append(list, &RequiredReserveRecord{
			Date:             row["date"],
			BusinessBankRate: row["businessBankRate"],
			OtherBankRate:    row["otherBankRate"],
		})
	}
	return list, nil
}

// ---------------------------------------------------------------------------
// 货币供应量（月度）
// ---------------------------------------------------------------------------

// MoneySupplyMonthRecord 货币供应量月度记录。
type MoneySupplyMonthRecord struct {
	StatDate string // 统计月份 (yyyy-mm)
	M0       string // 流通中现金 M0（亿元）
	M0YOY    string // M0 同比增长率
	M1       string // 狭义货币 M1
	M1YOY    string
	M2       string // 广义货币 M2
	M2YOY    string
}

// QueryMoneySupplyDataMonth 查询货币供应量（月度）。
// 日期格式：yyyy-mm（如 2024-01）
func (c *Client) QueryMoneySupplyDataMonth(startDate, endDate string) ([]*MoneySupplyMonthRecord, error) {
	r, err := c.macroDateQuery(MsgTypeMoneySupplyMonthReq, "query_money_supply_data_month", startDate, endDate)
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("QueryMoneySupplyDataMonth: %s - %s", r.ErrorCode, r.ErrorMsg)
	}
	list := make([]*MoneySupplyMonthRecord, 0, len(r.Rows))
	for _, row := range r.Rows {
		list = append(list, &MoneySupplyMonthRecord{
			StatDate: row["statDate"],
			M0:       row["m0"],
			M0YOY:    row["m0YOY"],
			M1:       row["m1"],
			M1YOY:    row["m1YOY"],
			M2:       row["m2"],
			M2YOY:    row["m2YOY"],
		})
	}
	return list, nil
}

// ---------------------------------------------------------------------------
// 货币供应量（年度）
// ---------------------------------------------------------------------------

// MoneySupplyYearRecord 货币供应量年底余额记录。
type MoneySupplyYearRecord struct {
	Year  string
	M0    string
	M0YOY string
	M1    string
	M1YOY string
	M2    string
	M2YOY string
}

// QueryMoneySupplyDataYear 查询货币供应量（年底余额）。
// 日期格式：yyyy（如 2024）
func (c *Client) QueryMoneySupplyDataYear(startDate, endDate string) ([]*MoneySupplyYearRecord, error) {
	r, err := c.macroDateQuery(MsgTypeMoneySupplyYearReq, "query_money_supply_data_year", startDate, endDate)
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("QueryMoneySupplyDataYear: %s - %s", r.ErrorCode, r.ErrorMsg)
	}
	list := make([]*MoneySupplyYearRecord, 0, len(r.Rows))
	for _, row := range r.Rows {
		list = append(list, &MoneySupplyYearRecord{
			Year:  row["year"],
			M0:    row["m0"],
			M0YOY: row["m0YOY"],
			M1:    row["m1"],
			M1YOY: row["m1YOY"],
			M2:    row["m2"],
			M2YOY: row["m2YOY"],
		})
	}
	return list, nil
}

// ---------------------------------------------------------------------------
// CPI（居民消费价格指数）
// ---------------------------------------------------------------------------

// CPIRecord CPI 记录。
type CPIRecord struct {
	Date     string
	CPIYOY   string // 居民消费价格指数（同比）
	CPIMOM   string // 居民消费价格指数（环比）
	CPIAccumulated string // 居民消费价格指数（累计）
}

// QueryCPIData 查询居民消费价格指数（CPI）。
func (c *Client) QueryCPIData(startDate, endDate string) ([]*CPIRecord, error) {
	r, err := c.macroDateQuery(MsgTypeCPIReq, "query_cpi_data", startDate, endDate)
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("QueryCPIData: %s - %s", r.ErrorCode, r.ErrorMsg)
	}
	list := make([]*CPIRecord, 0, len(r.Rows))
	for _, row := range r.Rows {
		list = append(list, &CPIRecord{
			Date:           row["date"],
			CPIYOY:         row["cpiYOY"],
			CPIMOM:         row["cpiMOM"],
			CPIAccumulated: row["cpiAccumulated"],
		})
	}
	return list, nil
}

// ---------------------------------------------------------------------------
// PPI（工业品出厂价格指数）
// ---------------------------------------------------------------------------

// PPIRecord PPI 记录。
type PPIRecord struct {
	Date           string
	PPIYOYIndustry string // 工业品出厂价格指数
	PPIYOYMining   string // 采掘工业价格指数
	PPIYOYRaw      string // 原材料工业价格指数
	PPIYOYProcessing string // 加工工业价格指数
	PPIYOYElectricity string // 生产资料价格指数
	PPIYOYLiving   string // 生活资料价格指数
	PPIAccumulated string // 工业品出厂价格指数（累计）
}

// QueryPPIData 查询工业品出厂价格指数（PPI）。
func (c *Client) QueryPPIData(startDate, endDate string) ([]*PPIRecord, error) {
	r, err := c.macroDateQuery(MsgTypePPIReq, "query_ppi_data", startDate, endDate)
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("QueryPPIData: %s - %s", r.ErrorCode, r.ErrorMsg)
	}
	list := make([]*PPIRecord, 0, len(r.Rows))
	for _, row := range r.Rows {
		list = append(list, &PPIRecord{
			Date:              row["date"],
			PPIYOYIndustry:    row["ppiYOYIndustry"],
			PPIYOYMining:      row["ppiYOYMining"],
			PPIYOYRaw:         row["ppiYOYRaw"],
			PPIYOYProcessing:  row["ppiYOYProcessing"],
			PPIYOYElectricity: row["ppiYOYElectricity"],
			PPIYOYLiving:      row["ppiYOYLiving"],
			PPIAccumulated:    row["ppiAccumulated"],
		})
	}
	return list, nil
}

// ---------------------------------------------------------------------------
// PMI（采购经理人指数）
// ---------------------------------------------------------------------------

// PMIRecord PMI 记录。
type PMIRecord struct {
	Date                string
	ManufacturingPMI    string // 制造业采购经理指数
	PMIProductionIndex  string // 生产指数
	PMISubOrdersIndex   string // 新订单指数
	PMISubExportOrders  string // 新出口订单指数
	PMISubInProcessGoods string // 在手订单指数
	PMISubFinishedGoods string // 产成品库存指数
	PMISubPurchaseGoods string // 采购量指数
	PMISubImports       string // 进口指数
	PMISubPurchasePrice string // 购进价格指数
	PMISubSellPrice     string // 出厂价格指数
	PMISubEmployees     string // 从业人员指数
	PMISubSupplierDeliveryTime string // 供应商配送时间指数
	PMISubRawMaterials  string // 原材料库存指数
	PMINonManufacturing string // 非制造业PMI
}

// QueryPMIData 查询采购经理人指数（PMI）。
func (c *Client) QueryPMIData(startDate, endDate string) ([]*PMIRecord, error) {
	r, err := c.macroDateQuery(MsgTypePMIReq, "query_pmi_data", startDate, endDate)
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("QueryPMIData: %s - %s", r.ErrorCode, r.ErrorMsg)
	}
	list := make([]*PMIRecord, 0, len(r.Rows))
	for _, row := range r.Rows {
		list = append(list, &PMIRecord{
			Date:                 row["date"],
			ManufacturingPMI:     row["manufacturingPMI"],
			PMIProductionIndex:   row["PMIProductionIndex"],
			PMISubOrdersIndex:    row["PMISubOrdersIndex"],
			PMISubExportOrders:   row["PMISubExportOrders"],
			PMISubInProcessGoods: row["PMISubInProcessGoods"],
			PMISubFinishedGoods:  row["PMISubFinishedGoods"],
			PMISubPurchaseGoods:  row["PMISubPurchaseGoods"],
			PMISubImports:        row["PMISubImports"],
			PMISubPurchasePrice:  row["PMISubPurchasePrice"],
			PMISubSellPrice:      row["PMISubSellPrice"],
			PMISubEmployees:      row["PMISubEmployees"],
			PMISubSupplierDeliveryTime: row["PMISubSupplierDeliveryTime"],
			PMISubRawMaterials:   row["PMISubRawMaterials"],
			PMINonManufacturing:  row["nonmanufacturingPMI"],
		})
	}
	return list, nil
}
