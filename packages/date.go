// name: date
// description: Date and time manipulation utilities
// author: roturbot
// requires: time

type Date struct{}

func (Date) now() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func (Date) nowUnix() int64 {
	return time.Now().Unix()
}

func (Date) nowUnixMilli() int64 {
	return time.Now().UnixMilli()
}

func (Date) nowUnixMicro() int64 {
	return time.Now().UnixMicro()
}

func (Date) nowUnixNano() int64 {
	return time.Now().UnixNano()
}

func (Date) fromUnix(timestamp any) string {
	timestampNum := int64(OSLcastNumber(timestamp))
	t := time.Unix(timestampNum, 0)
	return t.Format("2006-01-02 15:04:05")
}

func (Date) format(format any, unixTime any) string {
	formatStr := OSLtoString(format)
	unixTimeNum := int64(OSLcastNumber(unixTime))
	t := time.Unix(unixTimeNum, 0)
	return t.Format(formatStr)
}

func (Date) parse(format any, timeStr any) int64 {
	formatStr := OSLtoString(format)
	timeStrVal := OSLtoString(timeStr)

	t, err := time.Parse(formatStr, timeStrVal)
	if err != nil {
		return 0
	}
	return t.Unix()
}

func (Date) addDays(baseTime any, days any) int64 {
	baseTimeNum := int64(OSLcastNumber(baseTime))
	daysNum := int(OSLcastNumber(days))
	t := time.Unix(baseTimeNum, 0)
	newT := t.AddDate(0, 0, daysNum)
	return newT.Unix()
}

func (Date) addHours(baseTime any, hours any) int64 {
	baseTimeNum := int64(OSLcastNumber(baseTime))
	hoursNum := OSLcastNumber(hours)
	t := time.Unix(baseTimeNum, 0)
	newT := t.Add(time.Duration(hoursNum) * time.Hour)
	return newT.Unix()
}

func (Date) addMinutes(baseTime any, minutes any) int64 {
	baseTimeNum := int64(OSLcastNumber(baseTime))
	minutesNum := OSLcastNumber(minutes)
	t := time.Unix(baseTimeNum, 0)
	newT := t.Add(time.Duration(minutesNum) * time.Minute)
	return newT.Unix()
}

func (Date) addSeconds(baseTime any, seconds any) int64 {
	baseTimeNum := int64(OSLcastNumber(baseTime))
	secondsNum := int64(OSLcastNumber(seconds))
	t := time.Unix(baseTimeNum, 0)
	newT := t.Add(time.Duration(secondsNum) * time.Second)
	return newT.Unix()
}

func (Date) diff(time1 any, time2 any, unit any) float64 {
	time1Num := int64(OSLcastNumber(time1))
	time2Num := int64(OSLcastNumber(time2))
	unitStr := strings.ToLower(OSLtoString(unit))

	t1 := time.Unix(time1Num, 0)
	t2 := time.Unix(time2Num, 0)

	duration := t1.Sub(t2)

	switch unitStr {
	case "ns", "nanosecond", "nanoseconds":
		return float64(duration.Nanoseconds())
	case "us", "microsecond", "microseconds":
		return float64(duration.Microseconds())
	case "ms", "millisecond", "milliseconds":
		return float64(duration.Milliseconds())
	case "s", "second", "seconds":
		return duration.Seconds()
	case "m", "minute", "minutes":
		return duration.Minutes()
	case "h", "hour", "hours":
		return duration.Hours()
	case "d", "day", "days":
		return duration.Hours() / 24
	default:
		return duration.Seconds()
	}
}

func (Date) between(start any, end any, check any) bool {
	startTime := int64(OSLcastNumber(start))
	endTime := int64(OSLcastNumber(end))
	checkTime := int64(OSLcastNumber(check))

	return checkTime >= startTime && checkTime <= endTime
}

func (Date) year(unixTime any) int {
	t := time.Unix(int64(OSLcastNumber(unixTime)), 0)
	return t.Year()
}

func (Date) month(unixTime any) int {
	t := time.Unix(int64(OSLcastNumber(unixTime)), 0)
	return int(t.Month())
}

func (Date) day(unixTime any) int {
	t := time.Unix(int64(OSLcastNumber(unixTime)), 0)
	return t.Day()
}

func (Date) hour(unixTime any) int {
	t := time.Unix(int64(OSLcastNumber(unixTime)), 0)
	return t.Hour()
}

func (Date) minute(unixTime any) int {
	t := time.Unix(int64(OSLcastNumber(unixTime)), 0)
	return t.Minute()
}

func (Date) second(unixTime any) int {
	t := time.Unix(int64(OSLcastNumber(unixTime)), 0)
	return t.Second()
}

func (Date) weekday(unixTime any) string {
	t := time.Unix(int64(OSLcastNumber(unixTime)), 0)
	weekday := t.Weekday()
	return weekday.String()
}

func (Date) weekdayNumber(unixTime any) int {
	t := time.Unix(int64(OSLcastNumber(unixTime)), 0)
	weekday := t.Weekday()
	return int(weekday)
}

func (Date) yearday(unixTime any) int {
	t := time.Unix(int64(OSLcastNumber(unixTime)), 0)
	return t.YearDay()
}

func (Date) isLeap(year any) bool {
	yearNum := int(OSLcastNumber(year))
	return time.Date(yearNum, time.December, 31, 23, 59, 59, 0, time.UTC).YearDay() == 366
}

func (Date) daysInMonth(year any, month any) int {
	yearNum := int(OSLcastNumber(year))
	monthNum := int(OSLcastNumber(month))
	if monthNum < 1 || monthNum > 12 {
		return 0
	}

	t := time.Date(yearNum, time.Month(monthNum+1), 0, 23, 59, 59, 0, time.UTC)
	return t.Day()
}

func (Date) startOfDay(unixTime any) int64 {
	t := time.Unix(int64(OSLcastNumber(unixTime)), 0)
	startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return startOfDay.Unix()
}

func (Date) endOfDay(unixTime any) int64 {
	t := time.Unix(int64(OSLcastNumber(unixTime)), 0)
	endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
	return endOfDay.Unix()
}

func (Date) startOfWeek(unixTime any) int64 {
	t := time.Unix(int64(OSLcastNumber(unixTime)), 0)
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	startOfWeek := t.AddDate(0, 0, -weekday+1)
	startOfDay := time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())
	return startOfDay.Unix()
}

func (Date) startOfMonth(unixTime any) int64 {
	t := time.Unix(int64(OSLcastNumber(unixTime)), 0)
	startOfMonth := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	return startOfMonth.Unix()
}

func (Date) endOfMonth(unixTime any) int64 {
	t := time.Unix(int64(OSLcastNumber(unixTime)), 0)
	endOfMonth := t.AddDate(0, 1, 0)
	endOfMonth = time.Date(endOfMonth.Year(), endOfMonth.Month(), 0, 23, 59, 59, 999999999, endOfMonth.Location())
	return endOfMonth.Unix()
}

func (Date) age(birthDate any, currentDate any) int {
	birthTime := time.Unix(int64(OSLcastNumber(birthDate)), 0)
	currentTime := time.Unix(int64(OSLcastNumber(currentDate)), 0)

	years := currentTime.Year() - birthTime.Year()

	if currentTime.Month() < birthTime.Month() {
		years--
	} else if currentTime.Month() == birthTime.Month() && currentTime.Day() < birthTime.Day() {
		years--
	}

	return years
}

func (Date) timezone() string {
	return time.Now().Format("MST")
}

func (Date) utcOffset() int {
	_, offset := time.Now().Zone()
	return offset / 3600
}

func (Date) inTimezone(unixTime any, timezone any) int64 {
	timestampNum := int64(OSLcastNumber(unixTime))
	tzStr := OSLtoString(timezone)

	location, err := time.LoadLocation(tzStr)
	if err != nil {
		location, _ = time.LoadLocation("UTC")
	}

	t := time.Unix(timestampNum, 0).In(location)
	return t.Unix()
}

func (Date) isoString(unixTime any) string {
	t := time.Unix(int64(OSLcastNumber(unixTime)), 0)
	return t.Format(time.RFC3339)
}

func (Date) fromIso(isoString any) int64 {
	isoStr := OSLtoString(isoString)
	t, err := time.Parse(time.RFC3339, isoStr)
	if err != nil {
		return 0
	}
	return t.Unix()
}

func (Date) dateString(unixTime any) string {
	t := time.Unix(int64(OSLcastNumber(unixTime)), 0)
	return t.Format("2006-01-02")
}

func (Date) timeString(unixTime any) string {
	t := time.Unix(int64(OSLcastNumber(unixTime)), 0)
	return t.Format("15:04:05")
}

var date = Date{}
