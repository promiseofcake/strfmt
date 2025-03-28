// Copyright 2015 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package strfmt

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	// UnixZero sets the zero unix UTC timestamp we want to compare against.
	//
	// Unix 0 for an EST timezone is not equivalent to a UTC timezone.
	UnixZero = time.Unix(0, 0).UTC()
)

func init() {
	dt := DateTime{}
	Default.Add("datetime", &dt, IsDateTime)
}

// IsDateTime returns true when the string is a valid date-time.
//
// JSON datetime format consist of a date and a time separated by a "T", e.g. 2012-04-23T18:25:43.511Z.
func IsDateTime(str string) bool {
	const (
		minDateTimeLength = 4
		minParts          = 2
	)
	if len(str) < minDateTimeLength {
		return false
	}
	s := strings.Split(strings.ToLower(str), "t")
	if len(s) < minParts || !IsDate(s[0]) {
		return false
	}

	matches := rxDateTime.FindAllStringSubmatch(s[1], -1)
	if len(matches) == 0 || len(matches[0]) == 0 {
		return false
	}
	m := matches[0]
	res := m[1] <= "23" && m[2] <= "59" && m[3] <= "59"
	return res
}

const (
	// RFC3339Millis represents a ISO8601 format to millis instead of to nanos
	RFC3339Millis = "2006-01-02T15:04:05.000Z07:00"
	// RFC3339MillisNoColon represents a ISO8601 format to millis instead of to nanos
	RFC3339MillisNoColon = "2006-01-02T15:04:05.000Z0700"
	// RFC3339Micro represents a ISO8601 format to micro instead of to nano
	RFC3339Micro = "2006-01-02T15:04:05.000000Z07:00"
	// RFC3339MicroNoColon represents a ISO8601 format to micro instead of to nano
	RFC3339MicroNoColon = "2006-01-02T15:04:05.000000Z0700"
	// ISO8601LocalTime represents a ISO8601 format to ISO8601 in local time (no timezone)
	ISO8601LocalTime = "2006-01-02T15:04:05"
	// ISO8601TimeWithReducedPrecision represents a ISO8601 format with reduced precision (dropped secs)
	ISO8601TimeWithReducedPrecision = "2006-01-02T15:04Z"
	// ISO8601TimeWithReducedPrecisionLocaltime represents a ISO8601 format with reduced precision and no timezone (dropped seconds + no timezone)
	ISO8601TimeWithReducedPrecisionLocaltime = "2006-01-02T15:04"
	// ISO8601TimeUniversalSortableDateTimePattern represents a ISO8601 universal sortable date time pattern.
	ISO8601TimeUniversalSortableDateTimePattern = "2006-01-02 15:04:05"
	// short form of ISO8601TimeUniversalSortableDateTimePattern
	ISO8601TimeUniversalSortableDateTimePatternShortForm = "2006-01-02"
	// DateTimePattern pattern to match for the date-time format from http://tools.ietf.org/html/rfc3339#section-5.6
	DateTimePattern = `^([0-9]{2}):([0-9]{2}):([0-9]{2})(.[0-9]+)?(z|([+-][0-9]{2}:[0-9]{2}))$`
)

var (
	rxDateTime = regexp.MustCompile(DateTimePattern)

	// DateTimeFormats is the collection of formats used by ParseDateTime()
	DateTimeFormats = []string{RFC3339Micro, RFC3339MicroNoColon, RFC3339Millis, RFC3339MillisNoColon, time.RFC3339, time.RFC3339Nano, ISO8601LocalTime, ISO8601TimeWithReducedPrecision, ISO8601TimeWithReducedPrecisionLocaltime, ISO8601TimeUniversalSortableDateTimePattern, ISO8601TimeUniversalSortableDateTimePatternShortForm}

	// MarshalFormat sets the time resolution format used for marshaling time (set to milliseconds)
	MarshalFormat = RFC3339Millis

	// NormalizeTimeForMarshal provides a normalization function on time before marshalling (e.g. time.UTC).
	// By default, the time value is not changed.
	NormalizeTimeForMarshal = func(t time.Time) time.Time { return t }

	// DefaultTimeLocation provides a location for a time when the time zone is not encoded in the string (ex: ISO8601 Local variants).
	DefaultTimeLocation = time.UTC
)

// ParseDateTime parses a string that represents an ISO8601 time or a unix epoch
func ParseDateTime(data string) (DateTime, error) {
	if data == "" {
		return NewDateTime(), nil
	}
	var lastError error
	for _, layout := range DateTimeFormats {
		dd, err := time.ParseInLocation(layout, data, DefaultTimeLocation)
		if err != nil {
			lastError = err
			continue
		}
		return DateTime(dd), nil
	}
	return DateTime{}, lastError
}

// DateTime is a time but it serializes to ISO8601 format with millis.
//
// It knows how to read 3 different variations of a RFC3339 date time.
// Most APIs we encounter want either millisecond or second precision times.
// This just tries to make it worry-free.
//
// swagger:strfmt date-time
type DateTime time.Time

// NewDateTime is a representation of the UNIX epoch (January 1, 1970 00:00:00 UTC) for the [DateTime] type.
//
// Notice that this is not the zero value of the [DateTime] type.
//
// You may use [DateTime.IsUNIXZero] to check against this value.
func NewDateTime() DateTime {
	return DateTime(time.Unix(0, 0).UTC())
}

// MakeDateTime is a representation of the zero value of the [DateTime] type (January 1, year 1, 00:00:00 UTC).
//
// You may use [Datetime.IsZero] to check against this value.
func MakeDateTime() DateTime {
	return DateTime(time.Time{})
}

// String converts this time to a string
func (t DateTime) String() string {
	return NormalizeTimeForMarshal(time.Time(t)).Format(MarshalFormat)
}

// IsZero returns whether the date time is a zero value
func (t DateTime) IsZero() bool {
	return time.Time(t).IsZero()
}

// IsUnixZerom returns whether the date time is equivalent to time.Unix(0, 0).UTC().
func (t DateTime) IsUnixZero() bool {
	return time.Time(t) == UnixZero
}

// MarshalText implements the text marshaller interface
func (t DateTime) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText implements the text unmarshaller interface
func (t *DateTime) UnmarshalText(text []byte) error {
	tt, err := ParseDateTime(string(text))
	if err != nil {
		return err
	}
	*t = tt
	return nil
}

// Scan scans a DateTime value from database driver type.
func (t *DateTime) Scan(raw interface{}) error {
	// TODO: case int64: and case float64: ?
	switch v := raw.(type) {
	case []byte:
		return t.UnmarshalText(v)
	case string:
		return t.UnmarshalText([]byte(v))
	case time.Time:
		*t = DateTime(v)
	case nil:
		*t = DateTime{}
	default:
		return fmt.Errorf("cannot sql.Scan() strfmt.DateTime from: %#v: %w", v, ErrFormat)
	}

	return nil
}

// Value converts DateTime to a primitive value ready to written to a database.
func (t DateTime) Value() (driver.Value, error) {
	return driver.Value(t.String()), nil
}

// MarshalJSON returns the DateTime as JSON
func (t DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(NormalizeTimeForMarshal(time.Time(t)).Format(MarshalFormat))
}

// UnmarshalJSON sets the DateTime from JSON
func (t *DateTime) UnmarshalJSON(data []byte) error {
	if string(data) == jsonNull {
		return nil
	}

	var tstr string
	if err := json.Unmarshal(data, &tstr); err != nil {
		return err
	}
	tt, err := ParseDateTime(tstr)
	if err != nil {
		return err
	}
	*t = tt
	return nil
}

// DeepCopyInto copies the receiver and writes its value into out.
func (t *DateTime) DeepCopyInto(out *DateTime) {
	*out = *t
}

// DeepCopy copies the receiver into a new DateTime.
func (t *DateTime) DeepCopy() *DateTime {
	if t == nil {
		return nil
	}
	out := new(DateTime)
	t.DeepCopyInto(out)
	return out
}

// GobEncode implements the gob.GobEncoder interface.
func (t DateTime) GobEncode() ([]byte, error) {
	return t.MarshalBinary()
}

// GobDecode implements the gob.GobDecoder interface.
func (t *DateTime) GobDecode(data []byte) error {
	return t.UnmarshalBinary(data)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (t DateTime) MarshalBinary() ([]byte, error) {
	return NormalizeTimeForMarshal(time.Time(t)).MarshalBinary()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (t *DateTime) UnmarshalBinary(data []byte) error {
	var original time.Time

	err := original.UnmarshalBinary(data)
	if err != nil {
		return err
	}

	*t = DateTime(original)

	return nil
}

// Equal checks if two DateTime instances are equal using time.Time's Equal method
func (t DateTime) Equal(t2 DateTime) bool {
	return time.Time(t).Equal(time.Time(t2))
}
