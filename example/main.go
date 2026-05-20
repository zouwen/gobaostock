package main

import (
	"fmt"
	"log"

	"github.com/zouwen/gobaostock/baostock"
)

func main() {
	client := baostock.New()
	lg, err := client.Login("", "")
	if err != nil {
		log.Fatalf("登录失败: %v", err)
	}
	if !lg.Success() {
		log.Fatalf("登录错误: %s - %s", lg.ErrorCode, lg.ErrorMsg)
	}
	defer client.Logout()

	sep := func(title string) { fmt.Printf("\n===== %s =====\n", title) }

	// ── 1a. 强类型日线 QueryDailyBars ────────────────────────────────────
	sep("立讯精密 日K线-强类型（后复权，最近5条）")
	dailyBars, _ := client.QueryDailyBars("sz.002475",
		"date,open,high,low,close,preclose,volume,amount,turn,tradestatus,pctChg,isST",
		"2026-05-01", "2026-05-19", baostock.AdjustPost)
	fmt.Printf("共 %d 条\n", len(dailyBars))
	for i, bar := range dailyBars {
		if i >= 5 {
			break
		}
		fmt.Printf("  %s  O=%-8s H=%-8s L=%-8s C=%-8s  pctChg=%s%%  ST=%s\n",
			bar.Date, bar.Open, bar.High, bar.Low, bar.Close, bar.PctChg, bar.IsST)
	}

	// ── 1b. 强类型周线 QueryWeekBars ─────────────────────────────────────
	sep("立讯精密 周K线（不复权，近4周）")
	weekBars, _ := client.QueryWeekBars("sz.002475", "2026-04-01", "2026-05-19", baostock.AdjustNone)
	for _, bar := range weekBars {
		fmt.Printf("  %s  C=%s  pctChg=%s%%\n", bar.Date, bar.Close, bar.PctChg)
	}

	// ── 1c. 兼容旧接口 QueryHistoryKDataPlus（map 形式）────────────────────
	sep("立讯精密 日K线-原始map（后复权，最近5条）")
	kResult, _ := client.QueryHistoryKDataPlus("sz.002475",
		"date,open,high,low,close,volume,pctChg", "2026-05-01", "2026-05-19",
		baostock.FreqDay, baostock.AdjustPost)
	fmt.Printf("共 %d 条，字段: %v\n", len(kResult.Records), kResult.Fields)
	for i, rec := range kResult.Records {
		if i >= 5 {
			break
		}
		fmt.Printf("  %s  收盘=%-10s 涨跌幅=%s%%\n", rec["date"], rec["close"], rec["pctChg"])
	}

	// ── 2. 交易日历 ──────────────────────────────────────────────────────
	sep("交易日历（2026-05-10 ~ 2026-05-19）")
	dates, _ := client.QueryTradeDates("2026-05-10", "2026-05-19")
	for _, d := range dates {
		flag := "休市"
		if d.IsTradingDay == "1" {
			flag = "✓ 交易日"
		}
		fmt.Printf("  %s  %s\n", d.CalendarDate, flag)
	}

	// ── 3. 证券基本信息 ───────────────────────────────────────────────────
	sep("立讯精密基本资料")
	basics, _ := client.QueryStockBasic("sz.002475", "")
	for _, b := range basics {
		fmt.Printf("  %s(%s)  上市:%s  状态:%s  ST:%s\n",
			b.Code, b.CodeName, b.IPODate, b.Status, b.IsSt)
	}

	// ── 4. 行业分类 ───────────────────────────────────────────────────────
	sep("立讯精密行业分类")
	inds, _ := client.QueryStockIndustry("sz.002475", "")
	for _, ind := range inds {
		fmt.Printf("  %s | 行业: %s(%s)\n", ind.UpdateDate, ind.Industry, ind.IndustryClassification)
	}

	// ── 5. 指数成分股 ─────────────────────────────────────────────────────
	sep("沪深300成分股（前3条）")
	hs300, _ := client.QueryHS300Stocks("")
	for i, s := range hs300 {
		if i >= 3 {
			break
		}
		fmt.Printf("  %s  %s\n", s.Code, s.DisplayName)
	}

	// ── 6. 复权因子 ───────────────────────────────────────────────────────
	sep("立讯精密复权因子（2024-01-01 ~ 2024-12-31）")
	factors, _ := client.QueryAdjustFactor("sz.002475", "2024-01-01", "2024-12-31")
	for i, f := range factors {
		if i >= 3 {
			break
		}
		fmt.Printf("  %s  前复权=%-18s 后复权=%s\n", f.DivDate, f.FwdAdjFactor, f.BackAdjFactor)
	}

	// ── 7. 股息分红 ───────────────────────────────────────────────────────
	sep("立讯精密股息分红（2023年）")
	divs, _ := client.QueryDividendData("sz.002475", "2023", "report")
	for _, d := range divs {
		fmt.Printf("   除权日:%s  每股股息:%s\n", d.DividExDate, d.DividCashPS)
	}
	if len(divs) == 0 {
		fmt.Println("  暂无数据")
	}

	// ── 8. 季频盈利能力 ───────────────────────────────────────────────────
	sep("立讯精密盈利能力（2024Q4）")
	profits, _ := client.QueryProfitData("sz.002475", "2024", "4")
	for _, p := range profits {
		fmt.Printf("  代码:%s  字段数:%d\n", p.Code, len(p.Fields))
		for k, v := range p.Fields {
			if v != "" && k != "code" {
				fmt.Printf("    %-30s = %s\n", k, v)
			}
		}
	}
	if len(profits) == 0 {
		fmt.Println("  暂无数据")
	}

	// ── 9. 业绩快报 ───────────────────────────────────────────────────────
	sep("立讯精密业绩快报（2024年）")
	perfs, _ := client.QueryPerformanceExpressReport("sz.002475", "2024-01-01", "2024-12-31")
	for _, p := range perfs {
		fmt.Printf("  公告日:%s  净利润:%s  营收:%s\n", p.PubDate, p.NetProfit, p.OperatingIncome)
	}
	if len(perfs) == 0 {
		fmt.Println("  暂无数据")
	}

	// ── 10. 业绩预告 ──────────────────────────────────────────────────────
	sep("立讯精密业绩预告（2024年）")
	forecasts, _ := client.QueryForecastReport("sz.002475", "2024-01-01", "2024-12-31")
	for _, f := range forecasts {
		fmt.Printf("  公告日:%s  类型:%s  净利润区间:[%s, %s]\n",
			f.PubDate, f.Type, f.NetProfitMin, f.NetProfitMax)
	}
	if len(forecasts) == 0 {
		fmt.Println("  暂无数据")
	}

	// ── 11. 存款利率 ──────────────────────────────────────────────────────
	sep("存款利率（2023-01-01 ~ 2024-12-31）")
	deposits, _ := client.QueryDepositRateData("2023-01-01", "2024-12-31")
	for _, d := range deposits {
		fmt.Printf("  %s  活期=%s  1年定期=%s\n", d.Date, d.DepositRateShort, d.DepositRateMidLong)
	}
	if len(deposits) == 0 {
		fmt.Println("  暂无数据")
	}

	// ── 12. 存款准备金率 ──────────────────────────────────────────────────
	sep("存款准备金率（2023-01-01 ~ 2024-12-31）")
	reserves, _ := client.QueryRequiredReserveRatioData("2023-01-01", "2024-12-31", "0")
	for _, r := range reserves {
		fmt.Printf("  %s  大行=%s  小行=%s\n", r.Date, r.BusinessBankRate, r.OtherBankRate)
	}
	if len(reserves) == 0 {
		fmt.Println("  暂无数据")
	}

	// ── 13. 货币供应量（月度）───────────────────────────────────────────
	sep("货币供应量月度（2024-01 ~ 2024-03）")
	moneyMonth, _ := client.QueryMoneySupplyDataMonth("2024-01", "2024-03")
	for _, m := range moneyMonth {
		fmt.Printf("  %s  M0=%s(YOY:%s)  M2=%s(YOY:%s)\n",
			m.StatDate, m.M0, m.M0YOY, m.M2, m.M2YOY)
	}
	if len(moneyMonth) == 0 {
		fmt.Println("  暂无数据")
	}

	// ── 14. CPI ───────────────────────────────────────────────────────────
	sep("CPI（2024-01-01 ~ 2024-03-31）")
	cpis, _ := client.QueryCPIData("2024-01-01", "2024-03-31")
	for _, c := range cpis {
		fmt.Printf("  %s  YOY=%s  MOM=%s\n", c.Date, c.CPIYOY, c.CPIMOM)
	}
	if len(cpis) == 0 {
		fmt.Println("  暂无数据")
	}

	// ── 15. ST 股票列表 ───────────────────────────────────────────────────
	sep("ST股票（最新，前5条）")
	stStocks, _ := client.QueryStStocks("")
	for i, s := range stStocks {
		if i >= 5 {
			break
		}
		fmt.Printf("  %s  %s\n", s.Code, s.CodeName)
	}

	// ── 16. 创业板成分股 ──────────────────────────────────────────────────
	sep("创业板（GEM）成分股（前5条）")
	gem, _ := client.QueryGEMStocks("")
	for i, s := range gem {
		if i >= 5 {
			break
		}
		fmt.Printf("  %s  %s\n", s.Code, s.CodeName)
	}

	// ── 17. 沪港通 ────────────────────────────────────────────────────────
	sep("沪港通成分股（前3条）")
	shhk, _ := client.QuerySHHKStocks("")
	for i, s := range shhk {
		if i >= 3 {
			break
		}
		fmt.Printf("  %s  %s\n", s.Code, s.CodeName)
	}

	// ── 18. 概念板块 ──────────────────────────────────────────────────────
	sep("概念板块（前5条）")
	concepts, _ := client.QueryStockConcept("")
	for i, ct := range concepts {
		if i >= 5 {
			break
		}
		fmt.Printf("  概念:%s(%s)  股票:%s %s\n", ct.ConceptName, ct.ConceptCode, ct.Code, ct.CodeName)
	}
	if len(concepts) == 0 {
		fmt.Println("  暂无数据")
	}

	// ── 19. 地域板块 ──────────────────────────────────────────────────────
	sep("地域板块（前5条）")
	areas, _ := client.QueryStockArea("")
	for i, a := range areas {
		if i >= 5 {
			break
		}
		fmt.Printf("  地域:%s(%s)  股票:%s %s\n", a.AreaName, a.AreaCode, a.Code, a.CodeName)
	}
	if len(areas) == 0 {
		fmt.Println("  暂无数据")
	}

	// ── 20. SetAPIKey 用法示意（免费用户跳过实际调用） ────────────────────
	sep("SetAPIKey 用法示意（付费用户使用）")
	fmt.Println("  client2 := baostock.New()")
	fmt.Println("  client2.SetAPIKey(\"your-api-key\")  // 登录前设置")
	fmt.Println("  client2.Login(\"\", \"\")")
}
