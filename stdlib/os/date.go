package os

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hirochachacha/plua/object"
)

var formatFuncs = map[byte]func(time.Time) string{
	'a': func(t time.Time) string {
		return t.Format("Mon")
	},
	'A': func(t time.Time) string {
		return t.Format("Monday")
	},
	'b': func(t time.Time) string {
		return t.Format("Jan")
	},
	'B': func(t time.Time) string {
		return t.Format("January")
	},
	'c': func(t time.Time) string {
		return t.Format("Mon Jan 2 15:04:05 MST 2006")
	},
	'd': func(t time.Time) string {
		return t.Format("02")
	},
	'F': func(t time.Time) string {
		return t.Format("2006-01-02")
	},
	'H': func(t time.Time) string {
		return t.Format("15")
	},
	'I': func(t time.Time) string {
		return t.Format("03")
	},
	'j': func(t time.Time) string {
		return fmt.Sprintf("%03d", t.YearDay())
	},
	'm': func(t time.Time) string {
		return t.Format("01")
	},
	'M': func(t time.Time) string {
		return t.Format("04")
	},
	'p': func(t time.Time) string {
		return t.Format("PM")
	},
	'S': func(t time.Time) string {
		return t.Format("05")
	},
	// 'U': func(t time.Time) string {
	// },
	'w': func(t time.Time) string {
		return strconv.Itoa(int(t.Weekday()))
	},
	'W': func(t time.Time) string {
		_, week := t.ISOWeek()
		return fmt.Sprintf("%02d", week)
	},
	'x': func(t time.Time) string {
		return t.Format("Mon Jan 2")
	},
	'X': func(t time.Time) string {
		return t.Format("15:04:05")
	},
	'y': func(t time.Time) string {
		return t.Format("06")
	},
	'Y': func(t time.Time) string {
		return t.Format("2006")
	},
	'Z': func(t time.Time) string {
		return t.Format("MST")
	},
	'%': func(t time.Time) string {
		return "%"
	},
}

func dateFormat(th object.Thread, format string, t time.Time) (string, *object.RuntimeError) {
	var buf bytes.Buffer

	var start int
	var end int
	for {
		end = strings.IndexRune(format[start:], '%')
		if end == -1 {
			buf.WriteString(format[start:])
			break
		}

		end += start

		if end == len(format)-1 {
			return "", object.NewRuntimeError("invalid conversion specifier '%'")
		}

		buf.WriteString(format[start:end])

		if fn, ok := formatFuncs[format[end+1]]; ok {
			buf.WriteString(fn(t))
		} else {
			return "", object.NewRuntimeError("invalid conversion specifier '%" + string(format[end+1]) + "'")
		}

		if end == len(format)-2 {
			break
		}

		start = end + 2
	}

	return buf.String(), nil
}

func updateTable(th object.Thread, tab object.Table, t time.Time) {
	// suggested algorithm at http://play.golang.org/p/JvHUk1NjO5

	_, n := t.Zone()
	_, w := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location()).Zone()
	_, s := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location()).Zone()

	if w > s {
		w, s = s, w
	}

	isdst := w != s && n != w

	tab.Set(object.String("isdst"), object.Boolean(isdst))

	tab.Set(object.String("wday"), object.Integer(int(t.Weekday()+1)))
	tab.Set(object.String("year"), object.Integer(t.Year()))
	tab.Set(object.String("sec"), object.Integer(t.Second()))
	tab.Set(object.String("month"), object.Integer(int(t.Month())))
	tab.Set(object.String("day"), object.Integer(t.Day()))
	tab.Set(object.String("hour"), object.Integer(t.Hour()))
	tab.Set(object.String("yday"), object.Integer(int(t.YearDay())))
	tab.Set(object.String("min"), object.Integer(t.Minute()))
}
