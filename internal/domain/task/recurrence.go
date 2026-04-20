package task

import (
	"fmt"
	"time"
)

type RecurrenceType string

const (
	RecurrenceNone         RecurrenceType = "none"
	RecurrenceEveryDays    RecurrenceType = "every_n_days"
	RecurrenceMonthlyOnDay RecurrenceType = "monthly_on_day"
	RecurrenceOnceOnDate   RecurrenceType = "once_on_date"
	RecurrenceEvenDays     RecurrenceType = "even_days"
	RecurrenceOddDays      RecurrenceType = "odd_days"
)

type Recurrence struct {
	Type       RecurrenceType `json:"type"`
	DaysCount  int            `json:"days_count,omitempty"`
	DayOfMonth int            `json:"day_of_month,omitempty"`
	StartDate  *time.Time     `json:"start_date,omitempty"`
	EndDate    *time.Time     `json:"end_date,omitempty"`
}

func (r Recurrence) Valid() bool {
	if r.EndDate != nil && r.StartDate != nil && r.EndDate.Before(*r.StartDate) {
		return false
	}

	switch r.Type {
	case RecurrenceNone:
		return true
	case RecurrenceEveryDays:
		return r.DaysCount > 0
	case RecurrenceMonthlyOnDay:
		return r.DayOfMonth >= 1 && r.DayOfMonth <= 31
	case RecurrenceOnceOnDate:
		return r.StartDate != nil
	case RecurrenceEvenDays, RecurrenceOddDays:
		return r.StartDate != nil
	default:
		return false
	}
}

// DaysInMonth returns the number of days in a given month/year.
func DaysInMonth(month time.Month, year int) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// IsValidDayOfMonth checks if the given day exists in all 12 months.
// Returns a list of months where the day does NOT exist.
func (r Recurrence) InvalidMonths() []time.Month {
	if r.Type != RecurrenceMonthlyOnDay || r.DayOfMonth <= 28 {
		return nil
	}

	invalid := make([]time.Month, 0)
	for m := time.January; m <= time.December; m++ {
		days := DaysInMonth(m, 2025) // non-leap year is the strictest
		if r.DayOfMonth > days {
			invalid = append(invalid, m)
		}
	}
	return invalid
}

// DefaultHolidaysRU is a base set of Russian public holidays (month, day).
// Users can extend this via AddHoliday / RemoveHoliday.
var DefaultHolidaysRU = map[time.Month]map[int]bool{
	time.January:   {1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true, 8: true},
	time.February:  {23: true},
	time.March:     {8: true},
	time.May:       {1: true, 9: true},
	time.June:      {12: true},
	time.November:  {4: true},
	time.December:  {31: true},
}

// IsHoliday checks if a date is a holiday.
func IsHoliday(t time.Time) bool {
	if days, ok := DefaultHolidaysRU[t.Month()]; ok {
		return days[t.Day()]
	}
	return false
}

// IsWeekend checks if a date falls on a weekend.
func IsWeekend(t time.Time) bool {
	dow := t.Weekday()
	return dow == time.Saturday || dow == time.Sunday
}

// IsNonWorkingDay checks if a date is a weekend or holiday.
func IsNonWorkingDay(t time.Time) bool {
	return IsWeekend(t) || IsHoliday(t)
}

// NextWorkingDay returns the next working day after the given date.
func NextWorkingDay(from time.Time) time.Time {
	next := from.AddDate(0, 0, 1)
	for IsNonWorkingDay(next) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

// PrevWorkingDay returns the previous working day before the given date.
func PrevWorkingDay(from time.Time) time.Time {
	prev := from.AddDate(0, 0, -1)
	for IsNonWorkingDay(prev) {
		prev = prev.AddDate(0, 0, -1)
	}
	return prev
}

// SuggestAlternativeDates returns the previous and next 2-3 working days
// around the given non-working date.
func SuggestAlternativeDates(t time.Time) []time.Time {
	suggestions := make([]time.Time, 0, 5)

	// Previous working days
	prev := PrevWorkingDay(t)
	suggestions = append(suggestions, prev)
	prev2 := PrevWorkingDay(prev)
	suggestions = append(suggestions, prev2)
	prev3 := PrevWorkingDay(prev2)
	suggestions = append(suggestions, prev3)

	// Next working days
	next := NextWorkingDay(t)
	suggestions = append(suggestions, next)
	next2 := NextWorkingDay(next)
	suggestions = append(suggestions, next2)

	return suggestions
}

// NonWorkingDayConflict is returned when the start_date falls on a weekend or holiday.
type NonWorkingDayConflict struct {
	Date       time.Time
	Suggestion string
	Alts       []time.Time
}

func (e NonWorkingDayConflict) Error() string {
	return fmt.Sprintf("start_date %s falls on a non-working day (weekend/holiday). Please choose another date. Suggestions: %s",
		e.Date.Format("2006-01-02 15:04"),
		formatTimes(e.Alts))
}

func formatTimes(times []time.Time) string {
	result := ""
	for i, t := range times {
		if i > 0 {
			result += ", "
		}
		result += t.Format("2006-01-02")
	}
	return result
}

// ValidateStartDate checks if the start date falls on a non-working day.
// Returns NonWorkingDayConflict if it does, nil otherwise.
func ValidateStartDate(startDate time.Time) error {
	if !IsNonWorkingDay(startDate) {
		return nil
	}

	alts := SuggestAlternativeDates(startDate)
	return NonWorkingDayConflict{
		Date: startDate,
		Suggestion: fmt.Sprintf("falls on non-working day, suggested alternatives: %s",
			formatTimes(alts)),
		Alts: alts,
	}
}

// ParseStartDate parses a date string in "2006-01-02 15:04" format.
// Returns nil if the input is empty.
func ParseStartDate(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}

	t, err := time.Parse("2006-01-02 15:04", raw)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format, expected 'YYYY-MM-DD HH:MM': %w", err)
	}

	return &t, nil
}

// GenerateUpcomingDates generates the next N upcoming occurrence dates
// for this recurrence starting from StartDate.
// Returns dates with warnings if they fall on non-working days.
func (r Recurrence) GenerateUpcomingDates(from time.Time, count int) []DateInfo {
	if r.Type == RecurrenceNone || count <= 0 {
		return nil
	}

	startDate := from
	if r.StartDate != nil {
		startDate = *r.StartDate
	}

	endDate := r.EndDate
	results := make([]DateInfo, 0, count)

	switch r.Type {
	case RecurrenceEveryDays:
		current := startDate
		for len(results) < count {
			current = current.AddDate(0, 0, r.DaysCount)
			if endDate != nil && current.After(*endDate) {
				break
			}
			info := DateInfo{Date: current}
			if IsNonWorkingDay(current) {
				alts := SuggestAlternativeDates(current)
				info.Warning = fmt.Sprintf("falls on non-working day, suggested alternatives: %s", formatTimes(alts))
				if len(alts) > 0 {
					info.AlternativeDates = alts
				}
			}
			results = append(results, info)
		}

	case RecurrenceMonthlyOnDay:
		year, month, _ := from.Date()
		currentMonth := month
		currentYear := year

		for len(results) < count {
			daysInMonth := DaysInMonth(currentMonth, currentYear)
			day := r.DayOfMonth
			if day > daysInMonth {
				// Skip months where the day doesn't exist
				currentMonth++
				if currentMonth > time.December {
					currentMonth = time.January
					currentYear++
				}
				continue
			}

			date := time.Date(currentYear, currentMonth, day, startDate.Hour(), startDate.Minute(), startDate.Second(), startDate.Nanosecond(), startDate.Location())
			if !date.Before(from) {
				if endDate != nil && date.After(*endDate) {
					break
				}
				info := DateInfo{Date: date}
				if IsNonWorkingDay(date) {
					alts := SuggestAlternativeDates(date)
					info.Warning = fmt.Sprintf("falls on non-working day, suggested alternatives: %s", formatTimes(alts))
					if len(alts) > 0 {
						info.AlternativeDates = alts
					}
				}
				results = append(results, info)
			}

			currentMonth++
			if currentMonth > time.December {
				currentMonth = time.January
				currentYear++
			}
		}

	case RecurrenceOnceOnDate:
		if r.StartDate != nil {
			if endDate == nil || !r.StartDate.After(*endDate) {
				results = append(results, DateInfo{Date: *r.StartDate})
			}
		}

	case RecurrenceEvenDays, RecurrenceOddDays:
		current := startDate.AddDate(0, 0, 1)
		isEven := r.Type == RecurrenceEvenDays

		for len(results) < count {
			if endDate != nil && current.After(*endDate) {
				break
			}
			day := current.Day()
			if (isEven && day%2 == 0) || (!isEven && day%2 != 0) {
				info := DateInfo{Date: current}
				if IsNonWorkingDay(current) {
					alts := SuggestAlternativeDates(current)
					info.Warning = fmt.Sprintf("falls on non-working day, suggested alternatives: %s", formatTimes(alts))
					if len(alts) > 0 {
						info.AlternativeDates = alts
					}
				}
				results = append(results, info)
			}
			current = current.AddDate(0, 0, 1)
		}
	}

	if len(results) > count {
		results = results[:count]
	}

	return results
}

// DateInfo holds information about a scheduled date.
type DateInfo struct {
	Date            time.Time   `json:"date"`
	Warning         string      `json:"warning,omitempty"`
	AlternativeDate *time.Time  `json:"alternative_date,omitempty"`
	AlternativeDates []time.Time `json:"alternative_dates,omitempty"`
}

// AddHoliday adds a holiday to the holiday calendar.
func AddHoliday(month time.Month, day int) {
	if _, ok := DefaultHolidaysRU[month]; !ok {
		DefaultHolidaysRU[month] = make(map[int]bool)
	}
	DefaultHolidaysRU[month][day] = true
}

// RemoveHoliday removes a holiday from the holiday calendar.
func RemoveHoliday(month time.Month, day int) {
	if days, ok := DefaultHolidaysRU[month]; ok {
		delete(days, day)
	}
}
