package util

import (
	"math"
	"time"
)

// ToFixed rounds passed num to target precision.
func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

// GetToday returns today timeTime object.
func GetToday(location *time.Location) time.Time {
	var now time.Time
	now = time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	return now
}

// GetYesterday returns today timeTime object for yesterday 00:00:00
func GetYesterday(location *time.Location) time.Time {
	return GetToday(location).Add(-24 * time.Hour)
}

// Contains return true if x in a.
func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// GetDateRangeFunc returns a date range function over start date to end date inclusive.
// After the end of the range, the range function returns a zero date,
// date.IsZero() is true.
func GetDateRangeFunc(start, end time.Time) func() time.Time {
	y, m, d := start.Date()
	start = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	y, m, d = end.Date()
	end = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	return func() time.Time {
		if start.After(end) {
			return time.Time{}
		}
		date := start
		start = start.AddDate(0, 0, 1)
		return date
	}
}

// GetDateRangeArr returns array of dates in range
func GetDateRangeArr(startDate, endDate time.Time) []time.Time {
	var dateRange []time.Time
	for rd := GetDateRangeFunc(startDate, endDate); ; {
		date := rd()
		if date.IsZero() {
			break
		}
		dateRange = append(dateRange, date)
	}
	return dateRange
}

// round float to integer
func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

// IsDateEquals returns true if dates are equal (no Timezone compared)
func IsDateEquals(date1 time.Time, date2 time.Time) bool {
	return date1.Year() == date2.Year() && date1.Month() == date2.Month() && date1.Day() == date2.Day()
}

// UniqueStringSlice return slice with unique members of passed slice
func UniqueStringSlice(stringSlice []string) []string {
	var (
		list []string
		keys = make(map[string]bool)
	)
	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
