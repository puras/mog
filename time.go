package mog

import (
	"database/sql/driver"
	"time"
)

/**
* @project momo-backend
* @author <a href="mailto:he@puras.cn">Puras.He</a>
* @date 2021-10-03 16:08
 */

type Time time.Time

const TimeFormat = "2006-01-02 15:04:05"

func (t Time) String() string {
	return time.Time(t).Format(TimeFormat)
}

// UnmarshalJSON 转换成时间戳
func (t *Time) UnmarshalJSON(data []byte) (err error) {
	if len(data) == 2 {
		*t = Time(time.Time{})
		return
	}
	now, err := time.ParseInLocation(`"`+TimeFormat+`"`, string(data), time.Local)
	*t = Time(now)
	return
}

// MarshalJSON 转换成自定义格式
func (t Time) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(TimeFormat)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, TimeFormat)
	b = append(b, '"')
	return b, nil
}

func (t Time) Value() (driver.Value, error) {
	if t.String() == "0001-01-01 00:00:00" {
		return nil, nil
	}
	return []byte(time.Time(t).Format(TimeFormat)), nil
}

func (t *Time) Scan(v interface{}) error {
	time, _ := time.Parse("2006-01-02 15:04:05 +0800 CST", v.(time.Time).String())
	*t = Time(time)
	return nil
}
