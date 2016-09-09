package os

import (
	"bytes"
	"strings"
	"time"

	"github.com/hirochachacha/plua/internal/strconv"
	"github.com/hirochachacha/plua/object"
)

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

		switch format[start+1] {
		case 'a':
			buf.WriteString(t.Weekday().String()[:3])
		case 'A':
			buf.WriteString(t.Weekday().String())
		case 'b':
			buf.WriteString(t.Month().String()[:3])
		case 'B':
			buf.WriteString(t.Month().String())
		case 'c':
			buf.WriteString(t.Weekday().String()[:3])

			buf.WriteRune(' ')

			buf.WriteString(t.Month().String()[:3])

			buf.WriteRune(' ')

			day := t.Day()
			if day < 10 {
				buf.WriteRune(' ')
			}
			buf.WriteString(strconv.Itoa(day))

			buf.WriteRune(' ')

			hour := t.Hour()
			if hour < 10 {
				buf.WriteRune('0')
			}
			buf.WriteString(strconv.Itoa(hour))

			buf.WriteRune(':')

			minute := t.Minute()
			if minute < 10 {
				buf.WriteRune('0')
			}
			buf.WriteString(strconv.Itoa(minute))

			buf.WriteRune(':')

			second := t.Second()
			if second < 10 {
				buf.WriteRune('0')
			}
			buf.WriteString(strconv.Itoa(second))

			buf.WriteRune(' ')

			buf.WriteString(strconv.Itoa(t.Year()))
		case 'd':
			day := t.Day()
			if day < 10 {
				buf.WriteRune('0')
			}
			buf.WriteString(strconv.Itoa(day))
		case 'H':
			hour := t.Hour()
			if hour < 10 {
				buf.WriteRune('0')
			}
			buf.WriteString(strconv.Itoa(hour))
		case 'I':
			hour := t.Hour()
			if hour > 12 {
				hour -= 12
			}
			if hour < 10 {
				buf.WriteRune('0')
			}
			buf.WriteString(strconv.Itoa(hour))
		case 'M':
			minute := t.Minute()
			if minute < 10 {
				buf.WriteRune('0')
			}
			buf.WriteString(strconv.Itoa(minute))
		case 'm':
			month := int(t.Month())
			if month < 10 {
				buf.WriteRune('0')
			}
			buf.WriteString(strconv.Itoa(month))
		case 'p':
			hour := t.Hour()
			if hour > 12 {
				buf.WriteString("pm")
			} else {
				buf.WriteString("am")
			}
		case 'S':
			second := t.Second()
			if second < 10 {
				buf.WriteRune('0')
			}
			buf.WriteString(strconv.Itoa(second))
		case 'w':
			weekday := int(t.Weekday())
			buf.WriteString(strconv.Itoa(weekday))
		case 'x':
			month := int(t.Month())
			if month < 10 {
				buf.WriteRune('0')
			}
			buf.WriteString(strconv.Itoa(month))

			buf.WriteRune('/')

			day := t.Day()
			if day < 10 {
				buf.WriteRune('0')
			}
			buf.WriteString(strconv.Itoa(day))

			buf.WriteRune('/')

			year := t.Year()
			year_s := strconv.Itoa(year)

			if year < 10 {
				buf.WriteRune('0')
				buf.WriteString(year_s)
			} else {
				buf.WriteString(year_s[len(year_s)-2:])
			}
		case 'X':
			hour := t.Hour()
			if hour < 10 {
				buf.WriteRune('0')
			}
			buf.WriteString(strconv.Itoa(hour))

			buf.WriteRune(':')

			minute := t.Minute()
			if minute < 10 {
				buf.WriteRune('0')
			}
			buf.WriteString(strconv.Itoa(minute))

			buf.WriteRune(':')

			second := t.Second()
			if second < 10 {
				buf.WriteRune('0')
			}
			buf.WriteString(strconv.Itoa(second))
		case 'Y':
			buf.WriteString(strconv.Itoa(t.Year()))
		case 'y':
			year := t.Year()
			year_s := strconv.Itoa(year)

			if year < 10 {
				buf.WriteRune('0')
				buf.WriteString(year_s)
			} else {
				buf.WriteString(year_s[len(year_s)-2:])
			}
		case '%':
			buf.WriteRune('%')
		default:
			return "", object.NewRuntimeError("invalid conversion specifier '%" + string(format[start+1]) + "'")
		}

		if end == len(format)-2 {
			break
		}

		start = end + 2
	}

	return buf.String(), nil
}

func dateTable(th object.Thread, t time.Time) object.Table {
	table := th.NewTableSize(0, 8)

	// suggested algorithm at http://play.golang.org/p/JvHUk1NjO5

	_, n := t.Zone()
	_, w := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location()).Zone()
	_, s := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location()).Zone()

	if w > s {
		w, s = s, w
	}

	isdst := object.Boolean(w != s && n != w)

	table.Set(object.String("isdst"), isdst)

	table.Set(object.String("wday"), object.Integer(int(t.Weekday())))
	table.Set(object.String("year"), object.Integer(t.Year()))
	table.Set(object.String("sec"), object.Integer(t.Second()))
	table.Set(object.String("month"), object.Integer(int(t.Month())))
	table.Set(object.String("day"), object.Integer(t.Day()))
	table.Set(object.String("hour"), object.Integer(t.Hour()))
	table.Set(object.String("yday"), object.Integer(int(t.Weekday())))
	table.Set(object.String("min"), object.Integer(t.Minute()))

	return table
}
