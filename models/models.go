package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/adhocore/gronx"
)

type Version struct {
	Time time.Time `json:"time"`
}

type InRequest struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
}

type InResponse struct {
	Version  Version  `json:"version"`
	Metadata Metadata `json:"metadata"`
}

type OutRequest struct {
	Source Source `json:"source"`
}

type OutResponse struct {
	Version  Version  `json:"version"`
	Metadata Metadata `json:"metadata"`
}

type CheckRequest struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
}

type CheckResponse []Version

type Source struct {
	InitialVersion bool       `json:"initial_version"`
	Interval       *Interval  `json:"interval"`
	Start          *TimeOfDay `json:"start"`
	Stop           *TimeOfDay `json:"stop"`
	Days           []Weekday  `json:"days"`
	Cron           *Cron      `json:"cron"`
	Location       *Location  `json:"location"`
}

func (source Source) Validate() error {
	// Validate interval/start/stop/days and cron are not specified together
	if source.Cron != nil {
		if source.Interval != nil || source.Start != nil || source.Stop != nil || len(source.Days) > 0 {
			return errors.New("cannot configure 'interval' or 'start'/'stop' or 'days' with 'cron'")
		}
	}

	// Validate start and stop are both set or both unset
	if (source.Start != nil) != (source.Stop != nil) {
		if source.Start != nil {
			return errors.New("must configure 'stop' if 'start' is set")
		}
		return errors.New("must configure 'start' if 'stop' is set")
	}

	// Validate days if specified
	for _, day := range source.Days {
		if day < 0 || day > 6 {
			return fmt.Errorf("invalid day: %v", day)
		}
	}

	// Validate cron expression if provided
	if source.Cron != nil {
		if err := source.Cron.Validate(); err != nil {
			return fmt.Errorf("invalid cron expression: %v", err)
		}
	}

	return nil
}

type Metadata []MetadataField

type MetadataField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Interval time.Duration

func (i *Interval) UnmarshalJSON(payload []byte) error {
	var durStr string
	err := json.Unmarshal(payload, &durStr)
	if err != nil {
		return err
	}

	duration, err := time.ParseDuration(durStr)
	if err != nil {
		return err
	}

	*i = Interval(duration)

	return nil
}

func (i Interval) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(i).String())
}

type Location time.Location

func (l *Location) UnmarshalJSON(payload []byte) error {
	var locStr string
	err := json.Unmarshal(payload, &locStr)
	if err != nil {
		return err
	}

	location, err := time.LoadLocation(locStr)
	if err != nil {
		return err
	}

	*l = Location(*location)

	return nil
}

func (l Location) MarshalJSON() ([]byte, error) {
	return json.Marshal((*time.Location)(&l).String())
}

var timeFormats []string

func init() {
	timeFormats = append(timeFormats, "3:04 PM -0700")
	timeFormats = append(timeFormats, "3PM -0700")
	timeFormats = append(timeFormats, "3 PM -0700")
	timeFormats = append(timeFormats, "15:04 -0700")
	timeFormats = append(timeFormats, "1504 -0700")
	timeFormats = append(timeFormats, "3:04 PM")
	timeFormats = append(timeFormats, "3PM")
	timeFormats = append(timeFormats, "3 PM")
	timeFormats = append(timeFormats, "15:04")
	timeFormats = append(timeFormats, "1504")
}

type TimeOfDay time.Duration

func NewTimeOfDay(t time.Time) TimeOfDay {
	return TimeOfDay(time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute)
}

func (tod *TimeOfDay) UnmarshalJSON(payload []byte) error {
	var timeStr string
	err := json.Unmarshal(payload, &timeStr)
	if err != nil {
		return err
	}

	var t time.Time
	for _, format := range timeFormats {
		t, err = time.Parse(format, strings.ToUpper(timeStr))
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("invalid time format: %s, must be one of: %s", timeStr, strings.Join(timeFormats, ", "))
	}

	*tod = NewTimeOfDay(t.UTC())

	return nil
}

func (tod TimeOfDay) MarshalJSON() ([]byte, error) {
	return json.Marshal(tod.String())
}

func (tod TimeOfDay) Hour() int {
	return int(time.Duration(tod) / time.Hour)
}

func (tod TimeOfDay) Minute() int {
	return int(time.Duration(tod) % time.Hour / time.Minute)
}

func (tod TimeOfDay) String() string {
	return fmt.Sprintf("%d:%02d", tod.Hour(), tod.Minute())
}

type Weekday time.Weekday

func ParseWeekday(wdStr string) (time.Weekday, error) {
	switch strings.ToLower(wdStr) {
	case "sun", "sunday":
		return time.Sunday, nil
	case "mon", "monday":
		return time.Monday, nil
	case "tue", "tuesday":
		return time.Tuesday, nil
	case "wed", "wednesday":
		return time.Wednesday, nil
	case "thu", "thursday":
		return time.Thursday, nil
	case "fri", "friday":
		return time.Friday, nil
	case "sat", "saturday":
		return time.Saturday, nil
	}

	return 0, fmt.Errorf("unknown weekday: %s", wdStr)
}

func (x *Weekday) UnmarshalJSON(payload []byte) error {
	var wdStr string
	err := json.Unmarshal(payload, &wdStr)
	if err != nil {
		return err
	}

	wd, err := ParseWeekday(wdStr)
	if err != nil {
		return err
	}

	*x = Weekday(wd)

	return nil
}

func (wd Weekday) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Weekday(wd).String())
}

type Cron struct {
	Expression string
}

func (c *Cron) UnmarshalJSON(payload []byte) error {
	var cronStr string
	err := json.Unmarshal(payload, &cronStr)
	if err != nil {
		return err
	}

	g := gronx.New()
	if !g.IsValid(cronStr) {
		return fmt.Errorf("invalid cron expression: %s", cronStr)
	}

	c.Expression = cronStr
	return nil
}

func (c Cron) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Expression)
}

func (c Cron) Validate() error {
	g := gronx.New()
	if !g.IsValid(c.Expression) {
		return fmt.Errorf("invalid cron expression: %s", c.Expression)
	}

	// Check if the cron expression would run too frequently (less than 60 seconds)
	if err := c.validateMinimumInterval(); err != nil {
		return err
	}

	return nil
}

// validateMinimumInterval ensures no specific seconds are specified (to always run on minute boundaries)
func (c Cron) validateMinimumInterval() error {
	if strings.HasPrefix(c.Expression, "@") {
		if c.Expression == "@everysecond" {
			return errors.New("@everysecond is not supported: cron expressions must not specify seconds")
		}

		// Unknown macro - let it pass if gronx accepted it, assuming it doesn't operate at the second level
		return nil
	}

	// Check if the expression is a 6-field cron format (includes seconds)
	fields := strings.Fields(c.Expression)

	// If there are 6 fields, it's a cron expression with seconds, which we want to disallow
	if len(fields) == 6 {
		return errors.New("cron expressions with seconds field are not supported: use 5-field format")
	}

	return nil
}

// parsePositiveInt parses a string to an integer and ensures it's positive
func parsePositiveInt(s string) (int, error) {
	var value int
	_, err := fmt.Sscanf(s, "%d", &value)
	if err != nil {
		return 0, err
	}

	if value <= 0 {
		return 0, errors.New("value must be positive")
	}

	return value, nil
}

func (c *Cron) Next(t time.Time) (time.Time, error) {
	return gronx.NextTickAfter(c.Expression, t, false)
}

// NextIncludingCurrent returns the next time including the reference time if it matches
func (c *Cron) NextIncludingCurrent(t time.Time) (time.Time, error) {
	return gronx.NextTickAfter(c.Expression, t, true)
}

// NextN returns all next cron times between after and before
func (c *Cron) NextN(after time.Time, before time.Time, n int) []time.Time {
	var times []time.Time
	next, err := c.Next(after)
	if err != nil {
		return nil
	}

	for len(times) < n && next.Before(before) {
		times = append(times, next)
		next, err = c.Next(next)
		if err != nil {
			break
		}
	}
	return times
}
