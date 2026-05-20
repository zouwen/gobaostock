# gobaostock

> A complete Go rewrite of the [baostock](http://baostock.com) Python library — access China A-share market data via the baostock TCP protocol.

## Features

- ✅ Full TCP protocol implementation (binary framing + CRC32 + zlib)
- ✅ All 33 APIs ported from the original Python library
- ✅ Auto-pagination for K-line queries
- ✅ Zero external dependencies (pure Go stdlib)

## Installation

```bash
go get github.com/zouwen/gobaostock@latest
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/zouwen/gobaostock/baostock"
)

func main() {
    client := baostock.New()
    lg, _ := client.Login("", "")
    if !lg.Success() {
        panic(lg.ErrorMsg)
    }
    defer client.Logout()

    // K-line data (post-adjusted, daily)
    kdata, _ := client.QueryHistoryKDataPlus(
        "sz.002475",
        "date,open,high,low,close,volume,pctChg",
        "2026-01-01", "2026-05-19", "d", "2",
    )
    fmt.Printf("Got %d candles\n", len(kdata.Records))
    for _, row := range kdata.Records[:3] {
        fmt.Printf("  %s  close=%s  chg=%s%%\n", row["date"], row["close"], row["pctChg"])
    }
}
```

## API Reference

### Authentication
| Method | Description |
|--------|-------------|
| `New()` | Create a new client |
| `SetAPIKey(key)` | Set API key (paid users only, call before Login) |
| `Login(userID, password)` | Login (anonymous: pass empty strings) |
| `Logout()` | Logout and close connection |

### Market Data
| Method | Description |
|--------|-------------|
| `QueryHistoryKDataPlus(code, fields, start, end, freq, adjustFlag)` | K-line (OHLCV), auto-pagination |
| `QueryTradeDates(start, end)` | Trading calendar |
| `QueryAllStock(date)` | All listed stocks on a given date |
| `QueryStockBasic(code, codeName)` | Stock basic info (IPO date, status, ST flag) |
| `QueryStockIndustry(code, date)` | CSRC industry classification |

### Index Constituents
| Method | Description |
|--------|-------------|
| `QueryHS300Stocks(date)` | CSI 300 constituents |
| `QuerySZ50Stocks(date)` | SSE 50 constituents |
| `QueryZZ500Stocks(date)` | CSI 500 constituents |

### Sector Classification
| Method | Description |
|--------|-------------|
| `QueryGEMStocks(date)` | Growth Enterprise Market (创业板) |
| `QuerySMEStocks(date)` | SME Board (中小板) |
| `QuerySHHKStocks(date)` | Shanghai-HK Connect (沪港通) |
| `QuerySZHKStocks(date)` | Shenzhen-HK Connect (深港通) |
| `QueryStockInRisk(date)` | Risk Warning Board (风险警示板) |
| `QueryStStocks(date)` | ST stocks |
| `QueryStarStStocks(date)` | \*ST stocks |
| `QueryTerminatedStocks(date)` | Delisted stocks |
| `QuerySuspendedStocks(date)` | Suspended stocks |
| `QueryStockConcept(date)` | Concept sector (概念板块) |
| `QueryStockArea(date)` | Geographic sector (地域板块) |

### Corporate Reports
| Method | Description |
|--------|-------------|
| `QueryPerformanceExpressReport(code, start, end)` | Performance express (业绩快报) |
| `QueryForecastReport(code, start, end)` | Earnings forecast (业绩预告) |

### Quarterly Financial Metrics
| Method | Description |
|--------|-------------|
| `QueryDividendData(code, year, yearType)` | Dividend data |
| `QueryAdjustFactor(code, start, end)` | Price adjustment factors |
| `QueryProfitData(code, year, quarter)` | Profitability (ROE, net margin, EPS…) |
| `QueryOperationData(code, year, quarter)` | Operating efficiency (turnover ratios) |
| `QueryGrowthData(code, year, quarter)` | Growth capability (YOY growth rates) |
| `QueryDupontData(code, year, quarter)` | DuPont analysis |
| `QueryBalanceData(code, year, quarter)` | Solvency (current ratio, debt ratio…) |
| `QueryCashFlowData(code, year, quarter)` | Cash flow indicators |

### Macroeconomic Data
| Method | Description |
|--------|-------------|
| `QueryDepositRateData(start, end)` | Bank deposit rates |
| `QueryLoanRateData(start, end)` | Bank loan rates |
| `QueryRequiredReserveRatioData(start, end, yearType)` | Required reserve ratio |
| `QueryMoneySupplyDataMonth(start, end)` | Money supply M0/M1/M2 (monthly) |
| `QueryMoneySupplyDataYear(start, end)` | Money supply M0/M1/M2 (annual) |
| `QueryCPIData(start, end)` | Consumer Price Index |
| `QueryPPIData(start, end)` | Producer Price Index |
| `QueryPMIData(start, end)` | Purchasing Managers' Index |
| `QueryShiborData(date)` | Shanghai Interbank Offered Rate |

## K-line Fields

`frequency`: `d`=daily, `w`=weekly, `m`=monthly, `5`/`15`/`30`/`60`=minute

`adjustFlag`: `1`=pre-adjusted (前复权), `2`=post-adjusted (后复权), `3`=none

Available fields: `date`, `code`, `open`, `high`, `low`, `close`, `preclose`, `volume`, `amount`, `adjustflag`, `turn`, `tradestatus`, `pctChg`, `peTTM`, `psTTM`, `pcfNcfTTM`, `pbMRQ`, `isST`

## License

MIT
