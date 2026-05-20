// Package baostock - schedule.go
// 记录 BaoStock 各类数据的更新时间规则，便于调用方安排定时任务。
//
// 官方文档来源：https://www.baostock.com/mainContent?file=home.md
package baostock

import "time"

// ---------------------------------------------------------------------------
// 数据更新时间常量（每日）
// ---------------------------------------------------------------------------

// 当日 K 线 / 财务数据入库时间（Asia/Shanghai，UTC+8）。
const (
	// UpdateTimeDailyKLine 日K线数据入库时间：当前交易日 17:30。
	// 数据范围：1990-12-19 至今（股票）；2006-01-01 至今（指数）；2026-01-05 至今（ETF）。
	UpdateTimeDailyKLine = "17:30"

	// UpdateTimeAdjustFactor 复权因子数据入库时间：当前交易日 18:00。
	UpdateTimeAdjustFactor = "18:00"

	// UpdateTimeMinuteKLine 分钟K线数据（5/15/30/60分钟）入库时间：当前交易日 20:00。
	// 数据范围（近5年）：2020-01-03 至今（股票/ETF，不含指数）。
	UpdateTimeMinuteKLine = "20:00"

	// UpdateTimeFinancialReport 其他财务报告数据入库时间：第二自然日凌晨 1:30。
	// 包含：季频财务数据（资产负债、现金流量、利润、杜邦）、业绩快报、业绩预告。
	UpdateTimeFinancialReport = "01:30+1" // +1 表示第二自然日

	// UpdateTimeWeeklyKLine 周K线数据入库时间：每周六 17:30。
	UpdateTimeWeeklyKLine = "Saturday 17:30"

	// UpdateTimeMonthlyKLine 月K线数据入库时间：每月1号 17:30（上月数据）。
	UpdateTimeMonthlyKLine = "1st 17:30"
)

// ---------------------------------------------------------------------------
// 数据更新时间常量（每周）
// ---------------------------------------------------------------------------

const (
	// UpdateTimeIndexConstituents 指数成分股更新时间：每周一下午。
	// 覆盖：沪深300、上证50、中证500 成分股信息。
	UpdateTimeIndexConstituents = "Monday afternoon"
)

// ---------------------------------------------------------------------------
// 数据范围常量
// ---------------------------------------------------------------------------

const (
	// DataStartDateStock 股票日/周/月K线数据起始日期。
	DataStartDateStock = "1990-12-19"

	// DataStartDateIndex 指数K线数据起始日期。
	DataStartDateIndex = "2006-01-01"

	// DataStartDateETF ETF K线数据起始日期。
	DataStartDateETF = "2026-01-05"

	// DataStartDateMinute 分钟K线数据起始日期（近5年，股票/ETF，不含指数）。
	DataStartDateMinute = "2020-01-03"

	// DataStartDateFinancial 季频财务数据起始年份。
	DataStartDateFinancial = "2007-01-01"

	// DataStartDateForecast 业绩预告数据起始年份。
	DataStartDateForecast = "2003-01-01"

	// DataStartDatePerfExpress 业绩快报数据起始年份。
	DataStartDatePerfExpress = "2006-01-01"
)

// ---------------------------------------------------------------------------
// UpdateSchedule 数据更新计划（可直接被调用方用于设置 cron 任务）
// ---------------------------------------------------------------------------

// DataType 数据类型枚举。
type DataType int

const (
	DataTypeDailyKLine      DataType = iota // 日K线
	DataTypeMinuteKLine                     // 分钟K线（5/15/30/60分钟）
	DataTypeWeeklyKLine                     // 周K线
	DataTypeMonthlyKLine                    // 月K线
	DataTypeAdjustFactor                    // 复权因子
	DataTypeFinancialReport                 // 季频财务/业绩快报/业绩预告
	DataTypeIndexConstituents               // 指数成分股（沪深300/上证50/中证500）
)

// UpdateScheduleInfo 某类数据的更新计划信息。
type UpdateScheduleInfo struct {
	DataType    DataType
	Name        string // 数据名称
	CronExpr    string // 推荐的 cron 表达式（Asia/Shanghai）
	Description string // 更新规则说明
	DataRange   string // 历史数据范围
}

// AllUpdateSchedules 所有数据类型的更新计划，可用于配置 cron 任务。
//
// 示例——在交易日 17:35 拉取日K线：
//
//	schedule := baostock.AllUpdateSchedules[baostock.DataTypeDailyKLine]
//	fmt.Println(schedule.CronExpr)  // "35 17 * * 1-5"
var AllUpdateSchedules = map[DataType]*UpdateScheduleInfo{
	DataTypeDailyKLine: {
		DataType:    DataTypeDailyKLine,
		Name:        "日K线数据",
		CronExpr:    "35 17 * * 1-5", // 交易日 17:35（入库后5分钟）
		Description: "当前交易日 17:30 完成入库，建议 17:35 后拉取",
		DataRange:   "1990-12-19 至今（股票）；2006-01-01 至今（指数）；2026-01-05 至今（ETF）",
	},
	DataTypeAdjustFactor: {
		DataType:    DataTypeAdjustFactor,
		Name:        "复权因子数据",
		CronExpr:    "5 18 * * 1-5", // 交易日 18:05
		Description: "当前交易日 18:00 完成入库，建议 18:05 后拉取",
		DataRange:   "配合日K线使用，前复权/后复权",
	},
	DataTypeMinuteKLine: {
		DataType:    DataTypeMinuteKLine,
		Name:        "分钟K线数据（5/15/30/60分钟）",
		CronExpr:    "5 20 * * 1-5", // 交易日 20:05
		Description: "当前交易日 20:00 完成入库，建议 20:05 后拉取",
		DataRange:   "2020-01-03 至今（近5年，股票/ETF，不含指数）",
	},
	DataTypeFinancialReport: {
		DataType:    DataTypeFinancialReport,
		Name:        "财务报告数据（季频/业绩快报/业绩预告）",
		CronExpr:    "0 2 * * 2-6", // 周二至周六凌晨 02:00（入库后30分钟）
		Description: "前一交易日数据于第二自然日 01:30 完成入库，建议 02:00 后拉取",
		DataRange:   "季频财务: 2007-至今；业绩预告: 2003-至今；业绩快报: 2006-至今",
	},
	DataTypeWeeklyKLine: {
		DataType:    DataTypeWeeklyKLine,
		Name:        "周K线数据",
		CronExpr:    "0 18 * * 6", // 周六 18:00
		Description: "每周六 17:30 完成入库，建议 18:00 后拉取",
		DataRange:   "1990-12-19 至今（股票）；2006-01-01 至今（指数）",
	},
	DataTypeMonthlyKLine: {
		DataType:    DataTypeMonthlyKLine,
		Name:        "月K线数据",
		CronExpr:    "0 18 1 * *", // 每月1号 18:00
		Description: "每月1号 17:30 完成上月数据入库，建议 18:00 后拉取",
		DataRange:   "1990-12-19 至今（股票）；2006-01-01 至今（指数）",
	},
	DataTypeIndexConstituents: {
		DataType:    DataTypeIndexConstituents,
		Name:        "指数成分股（沪深300/上证50/中证500）",
		CronExpr:    "0 18 * * 1", // 每周一 18:00
		Description: "每周一下午完成入库，建议周一 18:00 后拉取",
		DataRange:   "沪深300、上证50、中证500 成分股信息",
	},
}

// ---------------------------------------------------------------------------
// 工具函数：判断当前时间数据是否已就绪
// ---------------------------------------------------------------------------

// IsDailyKLineReady 判断当前时间日K线数据是否已就绪（交易日 17:30 后）。
// loc 传 time.LoadLocation("Asia/Shanghai")，为 nil 时使用本地时区。
func IsDailyKLineReady(t time.Time, loc *time.Location) bool {
	if loc != nil {
		t = t.In(loc)
	}
	wd := t.Weekday()
	if wd == time.Saturday || wd == time.Sunday {
		return false // 非交易日
	}
	ready := time.Date(t.Year(), t.Month(), t.Day(), 17, 30, 0, 0, t.Location())
	return t.After(ready)
}

// IsMinuteKLineReady 判断当前时间分钟K线数据是否已就绪（交易日 20:00 后）。
func IsMinuteKLineReady(t time.Time, loc *time.Location) bool {
	if loc != nil {
		t = t.In(loc)
	}
	wd := t.Weekday()
	if wd == time.Saturday || wd == time.Sunday {
		return false
	}
	ready := time.Date(t.Year(), t.Month(), t.Day(), 20, 0, 0, 0, t.Location())
	return t.After(ready)
}

// IsAdjustFactorReady 判断复权因子数据是否已就绪（交易日 18:00 后）。
func IsAdjustFactorReady(t time.Time, loc *time.Location) bool {
	if loc != nil {
		t = t.In(loc)
	}
	wd := t.Weekday()
	if wd == time.Saturday || wd == time.Sunday {
		return false
	}
	ready := time.Date(t.Year(), t.Month(), t.Day(), 18, 0, 0, 0, t.Location())
	return t.After(ready)
}
