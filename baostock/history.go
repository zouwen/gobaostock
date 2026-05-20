package baostock

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// KRecord 一条 K 线记录，字段名对应 fields 参数中指定的列。
type KRecord map[string]string

// KDataResult 历史 K 线查询结果。
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

// QueryHistoryKDataPlus 查询历史 K 线数据（支持自动翻页，一次性返回所有数据）。
//
//	code       股票代码，格式 sh.600519 / sz.002475
//	fields     查询字段，逗号分隔，如 "date,open,high,low,close,volume,amount,pctChg"
//	startDate  开始日期，格式 2020-01-01
//	endDate    结束日期，格式 2026-05-18
//	frequency  d=日线, w=周线, m=月线, 5=5分钟, 15=15分钟, 30=30分钟, 60=60分钟
//	adjustFlag 1=前复权, 2=后复权, 3=不复权
func (c *Client) QueryHistoryKDataPlus(code, fields, startDate, endDate, frequency, adjustFlag string) (*KDataResult, error) {
	result := &KDataResult{}

	// 参数校验
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
		frequency = "d"
	}
	if adjustFlag == "" {
		adjustFlag = "3"
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

		// 首次：解析元信息
		if page == 1 {
			if len(arr) >= 13 {
				result.Code = arr[7]
				result.setFields(arr[8])
				result.StartDate = arr[9]
				result.EndDate = arr[10]
				result.Frequency = arr[11]
				result.AdjustFlag = arr[12]
			}
		}

		// 解析本页数据
		records, err := parseKRecords(arr[6], result.Fields)
		if err != nil {
			return nil, err
		}
		result.Records = append(result.Records, records...)

		// 判断是否还有下一页
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

// setFields 解析字段名列表（去掉空格）
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
	raw = strings.Join(strings.Fields(raw), "") // 去掉空白
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
