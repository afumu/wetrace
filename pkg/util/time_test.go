package util

import (
	"strings"
	"testing"
	"time"
)

func TestTimeOf(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantTime time.Time
		wantOk   bool
	}{
		// 空输入
		{
			name:     "empty string",
			input:    "",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "whitespace only",
			input:    "   ",
			wantTime: time.Time{},
			wantOk:   false,
		},

		// 自然语言时间
		{
			name:   "now",
			input:  "now",
			wantOk: true, // 不检查具体时间，因为会随运行时间变化
		},
		{
			name:   "today",
			input:  "today",
			wantOk: true,
		},
		{
			name:   "yesterday",
			input:  "yesterday",
			wantOk: true,
		},
		{
			name:   "this-week",
			input:  "this-week",
			wantOk: true,
		},
		{
			name:   "last-week",
			input:  "last-week",
			wantOk: true,
		},
		{
			name:   "this-month",
			input:  "this-month",
			wantOk: true,
		},
		{
			name:   "last-month",
			input:  "last-month",
			wantOk: true,
		},
		{
			name:   "this-year",
			input:  "this-year",
			wantOk: true,
		},
		{
			name:   "last-year",
			input:  "last-year",
			wantOk: true,
		},
		{
			name:   "all",
			input:  "all",
			wantOk: true,
		},

		// 相对时间
		{
			name:   "1h-ago",
			input:  "1h-ago",
			wantOk: true,
		},
		{
			name:   "24h-ago",
			input:  "24h-ago",
			wantOk: true,
		},
		{
			name:   "1d-ago",
			input:  "1d-ago",
			wantOk: true,
		},
		{
			name:   "7d-ago",
			input:  "7d-ago",
			wantOk: true,
		},
		{
			name:   "1w-ago",
			input:  "1w-ago",
			wantOk: true,
		},
		{
			name:   "1m-ago",
			input:  "1m-ago",
			wantOk: true,
		},
		{
			name:   "1y-ago",
			input:  "1y-ago",
			wantOk: true,
		},
		{
			name:     "invalid-ago",
			input:    "invalid-ago",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:   "0d-ago",
			input:  "0d-ago",
			wantOk: true, // 应该是今天
		},
		{
			name:     "-1d-ago",
			input:    "-1d-ago",
			wantTime: time.Time{},
			wantOk:   false, // 负数应该无效
		},

		// 季度
		{
			name:     "2020Q1",
			input:    "2020Q1",
			wantTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
			wantOk:   true,
		},
		{
			name:     "2020Q2",
			input:    "2020Q2",
			wantTime: time.Date(2020, 4, 1, 0, 0, 0, 0, time.Local),
			wantOk:   true,
		},
		{
			name:     "2020Q3",
			input:    "2020Q3",
			wantTime: time.Date(2020, 7, 1, 0, 0, 0, 0, time.Local),
			wantOk:   true,
		},
		{
			name:     "2020Q4",
			input:    "2020Q4",
			wantTime: time.Date(2020, 10, 1, 0, 0, 0, 0, time.Local),
			wantOk:   true,
		},
		{
			name:     "2020Q5", // 无效季度
			input:    "2020Q5",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "1969Q1", // 早于1970年
			input:    "1969Q1",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "10000Q1", // 超过9999年
			input:    "10000Q1",
			wantTime: time.Time{},
			wantOk:   false,
		},

		// 年份
		{
			name:     "2020",
			input:    "2020",
			wantTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
			wantOk:   true,
		},
		{
			name:     "1970", // 最早有效年份
			input:    "1970",
			wantTime: time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local),
			wantOk:   true,
		},
		{
			name:     "9999", // 最晚有效年份
			input:    "9999",
			wantTime: time.Date(9999, 1, 1, 0, 0, 0, 0, time.Local),
			wantOk:   true,
		},
		{
			name:     "1969", // 早于1970年
			input:    "1969",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "10000", // 超过9999年
			input:    "10000",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "202", // 不是4位数字
			input:    "202",
			wantTime: time.Time{},
			wantOk:   false,
		},

		// 月份
		{
			name:     "202001",
			input:    "202001",
			wantTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
			wantOk:   true,
		},
		{
			name:     "202012",
			input:    "202012",
			wantTime: time.Date(2020, 12, 1, 0, 0, 0, 0, time.Local),
			wantOk:   true,
		},
		{
			name:     "2020-01",
			input:    "2020-01",
			wantTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
			wantOk:   true,
		},
		{
			name:     "2020-12",
			input:    "2020-12",
			wantTime: time.Date(2020, 12, 1, 0, 0, 0, 0, time.Local),
			wantOk:   true,
		},
		{
			name:     "202013", // 无效月份
			input:    "202013",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "2020-13", // 无效月份
			input:    "2020-13",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "2020-00", // 无效月份
			input:    "2020-00",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "196912", // 早于1970年
			input:    "196912",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "1969-12", // 早于1970年
			input:    "1969-12",
			wantTime: time.Time{},
			wantOk:   false,
		},

		// 日期格式
		{
			name:     "20200101",
			input:    "20200101",
			wantTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
			wantOk:   true,
		},
		{
			name:     "20201231",
			input:    "20201231",
			wantTime: time.Date(2020, 12, 31, 0, 0, 0, 0, time.Local),
			wantOk:   true,
		},
		{
			name:     "2020-01-01",
			input:    "2020-01-01",
			wantTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
			wantOk:   true,
		},
		{
			name:     "2020-12-31",
			input:    "2020-12-31",
			wantTime: time.Date(2020, 12, 31, 0, 0, 0, 0, time.Local),
			wantOk:   true,
		},
		{
			name:     "20200229", // 闰年2月29日
			input:    "20200229",
			wantTime: time.Date(2020, 2, 29, 0, 0, 0, 0, time.Local),
			wantOk:   true,
		},
		{
			name:     "20190229", // 非闰年2月29日
			input:    "20190229",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "20200230", // 无效日期
			input:    "20200230",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "2020-02-30", // 无效日期
			input:    "2020-02-30",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "20200000", // 无效日期
			input:    "20200000",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "20200132", // 无效日期
			input:    "20200132",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "19691231", // 早于1970年
			input:    "19691231",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "1969-12-31", // 早于1970年
			input:    "1969-12-31",
			wantTime: time.Time{},
			wantOk:   false,
		},

		// 带时间的日期
		{
			name:     "20200101/12:34",
			input:    "20200101/12:34",
			wantTime: time.Date(2020, 1, 1, 12, 34, 0, 0, time.Local),
			wantOk:   true,
		},
		{
			name:     "2020-01-01/12:34",
			input:    "2020-01-01/12:34",
			wantTime: time.Date(2020, 1, 1, 12, 34, 0, 0, time.Local),
			wantOk:   true,
		},
		{
			name:     "20200101/24:00", // 无效时间
			input:    "20200101/24:00",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "20200101/12:60", // 无效时间
			input:    "20200101/12:60",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "20200101/12:34:56", // 不支持的格式
			input:    "20200101/12:34:56",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "20200101/12-34", // 不支持的格式
			input:    "20200101/12-34",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "19691231/12:34", // 早于1970年
			input:    "19691231/12:34",
			wantTime: time.Time{},
			wantOk:   false,
		},

		// 完整时间
		{
			name:     "20200101120000",
			input:    "20200101120000",
			wantTime: time.Date(2020, 1, 1, 12, 0, 0, 0, time.Local),
			wantOk:   true,
		},
		{
			name:     "20201231235959",
			input:    "20201231235959",
			wantTime: time.Date(2020, 12, 31, 23, 59, 59, 0, time.Local),
			wantOk:   true,
		},
		{
			name:     "20200101240000", // 无效时间
			input:    "20200101240000",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "20200101126000", // 无效时间
			input:    "20200101126000",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "20200101120060", // 无效时间
			input:    "20200101120060",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "2020010112000", // 长度不对
			input:    "2020010112000",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "202001011200000", // 长度不对
			input:    "202001011200000",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "19691231235959", // 早于1970年
			input:    "19691231235959",
			wantTime: time.Time{},
			wantOk:   false,
		},

		// 时间戳(秒)
		{
			name:     "1577836800", // 2020-01-01 00:00:00
			input:    "1577836800",
			wantTime: time.Unix(1577836800, 0),
			wantOk:   true,
		},
		{
			name:     "1609459199", // 2020-12-31 23:59:59
			input:    "1609459199",
			wantTime: time.Unix(1609459199, 0),
			wantOk:   true,
		},
		{
			name:     "999999999", // 小于1000000000
			input:    "999999999",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "253402300800", // 大于253402300799
			input:    "253402300800",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "abc", // 非数字
			input:    "abc",
			wantTime: time.Time{},
			wantOk:   false,
		},

		// RFC3339
		{
			name:     "2020-01-01T12:00:00Z",
			input:    "2020-01-01T12:00:00Z",
			wantTime: time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
			wantOk:   true,
		},
		{
			name:     "2020-01-01T12:00:00+08:00",
			input:    "2020-01-01T12:00:00+08:00",
			wantTime: time.Date(2020, 1, 1, 12, 0, 0, 0, time.FixedZone("", 8*60*60)),
			wantOk:   true,
		},
		{
			name:     "2020-01-01T12:00Z",
			input:    "2020-01-01T12:00Z",
			wantTime: time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
			wantOk:   true,
		},
		{
			name:     "2020-01-01T12:00:00", // 缺少时区
			input:    "2020-01-01T12:00:00",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "2020-01-01 12:00:00Z", // 格式不对
			input:    "2020-01-01 12:00:00Z",
			wantTime: time.Time{},
			wantOk:   false,
		},

		// 边界情况和特殊情况
		{
			name:     "99999", // 不是有效的时间戳
			input:    "99999",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "2020/01/01", // 不支持的格式
			input:    "2020/01/01",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "01/01/2020", // 不支持的格式
			input:    "01/01/2020",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "2020-1-1", // 不支持的格式
			input:    "2020-1-1",
			wantTime: time.Time{},
			wantOk:   false,
		},
		{
			name:     "20201-01-01", // 不支持的格式
			input:    "20201-01-01",
			wantTime: time.Time{},
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTime, gotOk := TimeOf(tt.input)

			if tt.wantOk != gotOk {
				t.Errorf("TimeOf() ok = %v, want %v", gotOk, tt.wantOk)
				return
			}

			if !tt.wantOk {
				return // 不需要检查时间值
			}

			if tt.input == "now" || strings.HasSuffix(tt.input, "-ago") ||
				tt.input == "today" || tt.input == "yesterday" ||
				tt.input == "this-week" || tt.input == "last-week" ||
				tt.input == "this-month" || tt.input == "last-month" ||
				tt.input == "this-year" || tt.input == "last-year" {
				// 对于相对时间，不检查具体值
				return
			}

			if !tt.wantTime.Equal(gotTime) {
				t.Errorf("TimeOf() = %v, want %v", gotTime, tt.wantTime)
			}
		})
	}
}

func TestTimeRangeOf(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantStart time.Time
		wantEnd   time.Time
		wantOk    bool
	}{
		// 空输入
		{
			name:      "empty string",
			input:     "",
			wantStart: time.Time{},
			wantEnd:   time.Time{},
			wantOk:    false,
		},
		{
			name:      "whitespace only",
			input:     "   ",
			wantStart: time.Time{},
			wantEnd:   time.Time{},
			wantOk:    false,
		},

		// all 特殊情况
		{
			name:      "all",
			input:     "all",
			wantStart: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC),
			wantOk:    true,
		},
		{
			name:      "ALL (uppercase)",
			input:     "ALL",
			wantStart: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC),
			wantOk:    true,
		},

		// 相对时间范围
		{
			name:   "last-1d",
			input:  "last-1d",
			wantOk: true, // 不检查具体时间，因为会随运行时间变化
		},
		{
			name:   "last-7d",
			input:  "last-7d",
			wantOk: true,
		},
		{
			name:   "last-30d",
			input:  "last-30d",
			wantOk: true,
		},
		{
			name:   "last-1w",
			input:  "last-1w",
			wantOk: true,
		},
		{
			name:   "last-4w",
			input:  "last-4w",
			wantOk: true,
		},
		{
			name:   "last-1m",
			input:  "last-1m",
			wantOk: true,
		},
		{
			name:   "last-3m",
			input:  "last-3m",
			wantOk: true,
		},
		{
			name:   "last-1y",
			input:  "last-1y",
			wantOk: true,
		},
		{
			name:      "last-0d", // 无效输入
			input:     "last-0d",
			wantStart: time.Time{},
			wantEnd:   time.Time{},
			wantOk:    false,
		},
		{
			name:      "last--1d", // 无效输入
			input:     "last--1d",
			wantStart: time.Time{},
			wantEnd:   time.Time{},
			wantOk:    false,
		},
		{
			name:      "last-1x", // 无效输入
			input:     "last-1x",
			wantStart: time.Time{},
			wantEnd:   time.Time{},
			wantOk:    false,
		},

		// 时间区间
		{
			name:      "2020-01-01~2020-01-31",
			input:     "2020-01-01~2020-01-31",
			wantStart: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2020, 1, 31, 23, 59, 59, 999999999, time.Local),
			wantOk:    true,
		},
		{
			name:      "2020-01-01,2020-01-31",
			input:     "2020-01-01,2020-01-31",
			wantStart: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2020, 1, 31, 23, 59, 59, 999999999, time.Local),
			wantOk:    true,
		},
		{
			name:      "2020-01-01 to 2020-01-31",
			input:     "2020-01-01 to 2020-01-31",
			wantStart: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2020, 1, 31, 23, 59, 59, 999999999, time.Local),
			wantOk:    true,
		},
		{
			name:      "20200101~20200131",
			input:     "20200101~20200131",
			wantStart: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2020, 1, 31, 23, 59, 59, 999999999, time.Local),
			wantOk:    true,
		},
		{
			name:      "2020-01-31~2020-01-01", // 开始时间晚于结束时间
			input:     "2020-01-31~2020-01-01",
			wantStart: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2020, 1, 31, 23, 59, 59, 999999999, time.Local),
			wantOk:    true, // 应该自动交换
		},
		{
			name:      "2020-01-01~invalid", // 结束时间无效
			input:     "2020-01-01~invalid",
			wantStart: time.Time{},
			wantEnd:   time.Time{},
			wantOk:    false,
		},
		{
			name:      "invalid~2020-01-31", // 开始时间无效
			input:     "invalid~2020-01-31",
			wantStart: time.Time{},
			wantEnd:   time.Time{},
			wantOk:    false,
		},
		{
			name:      "2020-01-01~2020-02-30", // 结束时间无效日期
			input:     "2020-01-01~2020-02-30",
			wantStart: time.Time{},
			wantEnd:   time.Time{},
			wantOk:    false,
		},
		{
			name:      "2020-01-01~", // 缺少结束时间
			input:     "2020-01-01~",
			wantStart: time.Time{},
			wantEnd:   time.Time{},
			wantOk:    false,
		},
		{
			name:      "~2020-01-31", // 缺少开始时间
			input:     "~2020-01-31",
			wantStart: time.Time{},
			wantEnd:   time.Time{},
			wantOk:    false,
		},

		// 单个时间点，根据粒度确定范围
		{
			name:      "2020-01-01", // 精确到天
			input:     "2020-01-01",
			wantStart: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2020, 1, 1, 23, 59, 59, 999999999, time.Local),
			wantOk:    true,
		},
		{
			name:      "20200101", // 精确到天
			input:     "20200101",
			wantStart: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2020, 1, 1, 23, 59, 59, 999999999, time.Local),
			wantOk:    true,
		},
		{
			name:      "2020-01", // 精确到月
			input:     "2020-01",
			wantStart: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2020, 1, 31, 23, 59, 59, 999999999, time.Local),
			wantOk:    true,
		},
		{
			name:      "202001", // 精确到月
			input:     "202001",
			wantStart: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2020, 1, 31, 23, 59, 59, 999999999, time.Local),
			wantOk:    true,
		},
		{
			name:      "2020Q1", // 精确到季度
			input:     "2020Q1",
			wantStart: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2020, 3, 31, 23, 59, 59, 999999999, time.Local),
			wantOk:    true,
		},
		{
			name:      "2020", // 精确到年
			input:     "2020",
			wantStart: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2020, 12, 31, 23, 59, 59, 999999999, time.Local),
			wantOk:    true,
		},
		{
			name:      "2020-01-01/12:34", // 精确到分钟
			input:     "2020-01-01/12:34",
			wantStart: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2020, 1, 1, 23, 59, 59, 999999999, time.Local),
			wantOk:    true,
		},
		{
			name:      "20200101120000", // 精确到秒
			input:     "20200101120000",
			wantStart: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2020, 1, 1, 23, 59, 59, 999999999, time.Local),
			wantOk:    true,
		},
		{
			name:      "1577836800", // 时间戳 2020-01-01 00:00:00
			input:     "1577836800",
			wantStart: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2020, 1, 1, 23, 59, 59, 999999999, time.Local),
			wantOk:    true,
		},
		{
			name:      "2020-01-01T12:00:00Z", // RFC3339
			input:     "2020-01-01T12:00:00Z",
			wantStart: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2020, 1, 1, 23, 59, 59, 999999999, time.UTC),
			wantOk:    true,
		},

		// 自然语言时间
		{
			name:   "today",
			input:  "today",
			wantOk: true,
		},
		{
			name:   "yesterday",
			input:  "yesterday",
			wantOk: true,
		},
		{
			name:   "this-week",
			input:  "this-week",
			wantOk: true,
		},
		{
			name:   "last-week",
			input:  "last-week",
			wantOk: true,
		},
		{
			name:   "this-month",
			input:  "this-month",
			wantOk: true,
		},
		{
			name:   "last-month",
			input:  "last-month",
			wantOk: true,
		},
		{
			name:   "this-year",
			input:  "this-year",
			wantOk: true,
		},
		{
			name:   "last-year",
			input:  "last-year",
			wantOk: true,
		},

		// 边界情况和特殊情况
		{
			name:      "invalid", // 无效输入
			input:     "invalid",
			wantStart: time.Time{},
			wantEnd:   time.Time{},
			wantOk:    false,
		},
		{
			name:      "2020-02-29", // 闰年2月29日
			input:     "2020-02-29",
			wantStart: time.Date(2020, 2, 29, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2020, 2, 29, 23, 59, 59, 999999999, time.Local),
			wantOk:    true,
		},
		{
			name:      "2019-02-29", // 非闰年2月29日
			input:     "2019-02-29",
			wantStart: time.Time{},
			wantEnd:   time.Time{},
			wantOk:    false,
		},
		{
			name:      "2020-04-31", // 无效日期
			input:     "2020-04-31",
			wantStart: time.Time{},
			wantEnd:   time.Time{},
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd, gotOk := TimeRangeOf(tt.input)

			if tt.wantOk != gotOk {
				t.Errorf("TimeRangeOf() ok = %v, want %v", gotOk, tt.wantOk)
				return
			}

			if !tt.wantOk {
				return // 不需要检查时间值
			}

			if tt.input == "today" || tt.input == "yesterday" ||
				tt.input == "this-week" || tt.input == "last-week" ||
				tt.input == "this-month" || tt.input == "last-month" ||
				tt.input == "this-year" || tt.input == "last-year" ||
				strings.HasPrefix(tt.input, "last-") {
				// 对于相对时间，不检查具体值
				return
			}

			if !tt.wantStart.Equal(gotStart) {
				t.Errorf("TimeRangeOf() start = %v, want %v", gotStart, tt.wantStart)
			}

			if !tt.wantEnd.Equal(gotEnd) {
				t.Errorf("TimeRangeOf() end = %v, want %v", gotEnd, tt.wantEnd)
			}
		})
	}
}

// 测试边界情况
func TestTimeOfEdgeCases(t *testing.T) {
	// 测试非常长的数字字符串
	longDigits := "99999999999999999999999999999999999999"
	_, ok := TimeOf(longDigits)
	if ok {
		t.Errorf("TimeOf(%s) should return false for very long digit string", longDigits)
	}

	// 测试非常长的字符串
	longString := strings.Repeat("a", 10000)
	_, ok = TimeOf(longString)
	if ok {
		t.Errorf("TimeOf(%s) should return false for very long string", "very_long_string")
	}

	// 测试特殊字符
	specialChars := "!@#$%^&*()"
	_, ok = TimeOf(specialChars)
	if ok {
		t.Errorf("TimeOf(%s) should return false for special characters", specialChars)
	}

	// 测试SQL注入类字符串
	sqlInjection := "2020-01-01' OR '1'='1"
	_, ok = TimeOf(sqlInjection)
	if ok {
		t.Errorf("TimeOf(%s) should return false for SQL injection attempt", sqlInjection)
	}
}

// 测试时区处理
func TestTimeOfTimezones(t *testing.T) {
	// RFC3339格式的时区处理
	utcTime, ok := TimeOf("2020-01-01T12:00:00Z")
	if !ok {
		t.Fatalf("TimeOf(2020-01-01T12:00:00Z) failed")
	}

	estTime, ok := TimeOf("2020-01-01T12:00:00-05:00")
	if !ok {
		t.Fatalf("TimeOf(2020-01-01T12:00:00-05:00) failed")
	}

	// UTC比EST时区快5小时，所以相同时钟时间的UTC应该比EST早5小时
	// 转换为UTC后比较
	utcInUTC := utcTime.UTC()
	estInUTC := estTime.UTC()

	// EST时区是-5小时，所以相同时钟时间的EST转为UTC后应该比UTC的时钟时间多5小时
	hourDiff := estInUTC.Hour() - utcInUTC.Hour()
	if hourDiff != 5 {
		t.Errorf("Expected 5 hour difference between UTC and EST, got %v", hourDiff)
	}
}

// 测试闰年处理
func TestLeapYearHandling(t *testing.T) {
	// 测试闰年2月29日
	leapDay, ok := TimeOf("20200229")
	if !ok {
		t.Fatalf("TimeOf(20200229) failed for leap year")
	}
	if leapDay.Day() != 29 || leapDay.Month() != 2 || leapDay.Year() != 2020 {
		t.Errorf("Expected 2020-02-29, got %v", leapDay)
	}

	// 测试非闰年2月29日
	_, ok = TimeOf("20190229")
	if ok {
		t.Errorf("TimeOf(20190229) should fail for non-leap year")
	}

	// 测试世纪闰年规则 (2000是闰年，2100不是)
	_, ok = TimeOf("20000229")
	if !ok {
		t.Errorf("TimeOf(20000229) should succeed for century leap year")
	}

	_, ok = TimeOf("21000229")
	if ok {
		t.Errorf("TimeOf(21000229) should fail for non-leap century year")
	}
}
