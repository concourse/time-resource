package resource

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/concourse/time-resource/lord"
	"github.com/concourse/time-resource/models"
)

type CheckCommand struct {
}

// DescribeCron returns a human-readable explanation of a cron expression
func DescribeCron(expr string) string {
	// Handle common macros
	switch expr {
	case "@yearly", "@annually":
		return "triggers once a year at midnight on January 1st"
	case "@monthly":
		return "triggers at midnight on the 1st of every month"
	case "@weekly":
		return "triggers at midnight every Sunday"
	case "@daily", "@midnight":
		return "triggers once a day at midnight"
	case "@hourly":
		return "triggers at the start of every hour"
	}

	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return fmt.Sprintf("schedule: %s", expr)
	}

	minute, hour, dom, month, dow := fields[0], fields[1], fields[2], fields[3], fields[4]
	var parts []string
	var warnings []string

	// Day-of-week
	if dow != "*" && dow != "?" {
		dowDesc, dowWarnings := describeDOW(dow)
		parts = append(parts, "on "+dowDesc)
		warnings = append(warnings, dowWarnings...)
	}

	// Day-of-month
	domDesc, domWarnings := describeDOM(dom)
	if domDesc != "" {
		parts = append(parts, domDesc)
	}
	warnings = append(warnings, domWarnings...)

	// DOM + DOW = OR logic warning
	if dom != "*" && dom != "?" && dow != "*" && dow != "?" {
		warnings = append(warnings, "note: day-of-month AND day-of-week uses OR logic, not AND (triggers on EITHER match)")
	}

	// Month
	if month != "*" {
		parts = append(parts, "in "+describeMonth(month))
	}

	// Time description
	timeDesc := describeTime(minute, hour)
	if timeDesc != "" {
		parts = append(parts, timeDesc)
	}

	// DST warning for specific hours
	if hour != "*" && !strings.Contains(hour, "/") && !strings.Contains(hour, ",") {
		h, err := strconv.Atoi(hour)
		if err == nil && h >= 1 && h <= 3 {
			warnings = append(warnings, "note: may skip or double-trigger during DST transitions")
		}
	}

	if len(parts) == 0 {
		return fmt.Sprintf("schedule: %s", expr)
	}

	result := "triggers " + strings.Join(parts, ", ")
	if len(warnings) > 0 {
		result += "; " + strings.Join(warnings, "; ")
	}
	return result
}

func describeTime(minute, hour string) string {
	// Every N minutes
	if strings.HasPrefix(minute, "*/") {
		step := strings.TrimPrefix(minute, "*/")
		return fmt.Sprintf("every %s minutes", step)
	}

	// Every N hours
	if strings.HasPrefix(hour, "*/") {
		step := strings.TrimPrefix(hour, "*/")
		if minute == "0" {
			return fmt.Sprintf("every %s hours at minute 0", step)
		}
		return fmt.Sprintf("every %s hours at minute %s", step, minute)
	}

	// Specific time
	if hour != "*" && minute != "*" {
		h, _ := strconv.Atoi(hour)
		m, _ := strconv.Atoi(minute)
		return fmt.Sprintf("at %02d:%02d", h, m)
	}

	if hour != "*" {
		return fmt.Sprintf("during hour %s", hour)
	}

	if minute != "*" && !strings.Contains(minute, "/") {
		return fmt.Sprintf("at minute %s of every hour", minute)
	}

	return ""
}

func describeDOW(dow string) (string, []string) {
	days := map[string]string{
		"0": "Sunday", "1": "Monday", "2": "Tuesday", "3": "Wednesday",
		"4": "Thursday", "5": "Friday", "6": "Saturday",
		"SUN": "Sunday", "MON": "Monday", "TUE": "Tuesday", "WED": "Wednesday",
		"THU": "Thursday", "FRI": "Friday", "SAT": "Saturday",
	}

	var warnings []string
	dowUpper := strings.ToUpper(dow)

	if strings.Contains(dow, "#") {
		parts := strings.Split(dow, "#")
		if len(parts) == 2 {
			dayName := parts[0]
			if name, ok := days[strings.ToUpper(dayName)]; ok {
				dayName = name
			}
			nth := parts[1]
			ordinal := toOrdinal(nth)
			if nth == "5" {
				warnings = append(warnings, "note: 5th occurrence only exists in some months")
			}
			return fmt.Sprintf("%s %s of the month", ordinal, dayName), warnings
		}
	}

	if strings.HasSuffix(dowUpper, "L") {
		dayNum := strings.TrimSuffix(dow, "L")
		dayNum = strings.TrimSuffix(dayNum, "l")
		if name, ok := days[dayNum]; ok {
			return fmt.Sprintf("last %s of the month", name), warnings
		}
		return fmt.Sprintf("last %s of the month", dayNum), warnings
	}

	if strings.Contains(dow, "-") {
		parts := strings.Split(dow, "-")
		if len(parts) == 2 {
			startDay := parts[0]
			endDay := parts[1]
			if name, ok := days[strings.ToUpper(startDay)]; ok {
				startDay = name
			}
			if name, ok := days[strings.ToUpper(endDay)]; ok {
				endDay = name
			}
			return fmt.Sprintf("%s through %s", startDay, endDay), warnings
		}
	}

	if name, ok := days[dowUpper]; ok {
		return name, warnings
	}
	return dow, warnings
}

func toOrdinal(n string) string {
	num, err := strconv.Atoi(n)
	if err != nil {
		return n
	}
	suffix := "th"
	if num%100 < 11 || num%100 > 13 {
		switch num % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		}
	}
	return fmt.Sprintf("%d%s", num, suffix)
}

func describeDOM(dom string) (string, []string) {
	var warnings []string

	if dom == "*" || dom == "?" {
		return "", warnings
	}

	if strings.HasPrefix(dom, "*/") {
		step := strings.TrimPrefix(dom, "*/")
		desc := fmt.Sprintf("every %s days from 1st of month", step)
		stepNum, err := strconv.Atoi(step)
		if err == nil && stepNum > 0 {
			for day := 1; day <= 31; day += stepNum {
				if day == 31 {
					warnings = append(warnings, "note: 31st then 1st = back-to-back triggers")
					break
				}
			}
		}
		return desc, warnings
	}

	domUpper := strings.ToUpper(dom)

	if domUpper == "L" {
		return "on the last day of the month", warnings
	}

	if strings.HasSuffix(domUpper, "W") {
		dayNum := strings.TrimSuffix(dom, "W")
		dayNum = strings.TrimSuffix(dayNum, "w")
		desc := fmt.Sprintf("on the nearest weekday to the %s", toOrdinal(dayNum))
		if dayNum == "31" {
			warnings = append(warnings, "note: only triggers in months with 31 days")
		} else if dayNum == "30" {
			warnings = append(warnings, "note: skips February")
		} else if dayNum == "29" {
			warnings = append(warnings, "note: only triggers in leap years for February")
		}
		return desc, warnings
	}

	if dom == "31" {
		warnings = append(warnings, "note: only triggers in months with 31 days (Jan, Mar, May, Jul, Aug, Oct, Dec)")
	} else if dom == "30" {
		warnings = append(warnings, "note: skips February")
	} else if dom == "29" {
		warnings = append(warnings, "note: only triggers in leap years for February")
	}

	return "on day " + dom + " of the month", warnings
}

func describeMonth(month string) string {
	months := map[string]string{
		"1": "January", "2": "February", "3": "March", "4": "April",
		"5": "May", "6": "June", "7": "July", "8": "August",
		"9": "September", "10": "October", "11": "November", "12": "December",
	}
	if name, ok := months[month]; ok {
		return name
	}
	return "month " + month
}

func (*CheckCommand) Run(request models.CheckRequest) ([]models.Version, error) {
	err := request.Source.Validate()
	if err != nil {
		return nil, err
	}

	previousTime := request.Version.Time
	currentTime := time.Now().UTC()

	specifiedLocation := request.Source.Location
	if specifiedLocation != nil {
		currentTime = currentTime.In((*time.Location)(specifiedLocation))
	}

	tl := lord.TimeLord{
		PreviousTime: previousTime,
		Location:     specifiedLocation,
		Start:        request.Source.Start,
		Stop:         request.Source.Stop,
		Interval:     request.Source.Interval,
		Days:         request.Source.Days,
		StartAfter:   request.Source.StartAfter,
		Cron:         request.Source.Cron,
	}

	var versions []models.Version

	if !previousTime.IsZero() {
		versions = append(versions, models.Version{Time: previousTime})
	} else if request.Source.InitialVersion {
		// For cron with initial_version, use the cron boundary time for consistency.
		// For non-cron, use currentTime
		versionTime := currentTime
		if request.Source.Cron != nil {
			cronTime := tl.Latest(currentTime)
			if !cronTime.IsZero() {
				versionTime = cronTime
				fmt.Fprintf(os.Stderr, "cron: emitting initial version at %s\n  %s\n",
					versionTime.Format(time.RFC3339), DescribeCron(request.Source.Cron.Expression))
			}
		}
		versions = append(versions, models.Version{Time: versionTime})
		return versions, nil
	} else if request.Source.Cron != nil {
		// Cron with initial_version:false (or unset) and no previous version:
		// don't emit any version until after the first cron trigger is observed.
		return versions, nil
	}

	if tl.Check(currentTime) {
		var versionTime time.Time

		// For cron expressions, use the actual scheduled cron time
		// instead of the check time. This ensures versions are at cron boundaries.
		// Example: cron @5minutes, check at 3:07pm → version time = 3:05pm
		if request.Source.Cron != nil {
			versionTime = tl.Latest(currentTime)
			if !versionTime.IsZero() {
				fmt.Fprintf(os.Stderr, "cron: emitting version at %s (previous: %s)\n  %s\n",
					versionTime.Format(time.RFC3339), previousTime.Format(time.RFC3339),
					DescribeCron(request.Source.Cron.Expression))
			}
		}

		// For non-cron (interval, start/stop ranges), use currentTime
		if versionTime.IsZero() {
			versionTime = currentTime
		}

		versions = append(versions, models.Version{Time: versionTime})
	}

	return versions, nil
}
