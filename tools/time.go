package tools

import "time"
import "fmt"

var OneDay, _ = time.ParseDuration("24h")

func CalcDayNum(start *time.Time, end *time.Time) int {
	s := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	e := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())

	return int(e.Sub(s).Hours() / 24)
}

func DateFromTime(input *time.Time) time.Time {
	return time.Date(input.Year(), input.Month(), input.Day(), 0, 0, 0, 0, time.Local)
}

func Adday(input *time.Time, dayNum int) {
	d, _ := time.ParseDuration(fmt.Sprintf("%dh", 24*dayNum))
	input.Add(d)
}
