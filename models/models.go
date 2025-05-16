package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
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
	InitialVersion bool        `json:"initial_version"`
	Interval       *Interval   `json:"interval"`
	Start          *TimeOfDay  `json:"start"`
	Stop           *TimeOfDay  `json:"stop"`
	Days           []Weekday   `json:"days"`
	Location       *Location   `json:"location"`
	StartAfter     *StartAfter `json:"start_after"`
}

func (source Source) Validate() error {
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
		t, err = time.Parse(format, timeStr)
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

var dateTimeFormats = []string{
	"2006-01-02T15:04:05",
	"2006-01-02T15:04",
	"2006-01-02T15",
	time.DateOnly,
	time.DateTime,
}

type StartAfter time.Time

func (sa *StartAfter) UnmarshalJSON(payload []byte) error {
	var dateTimeStr string

	err := json.Unmarshal(payload, &dateTimeStr)
	if err != nil {
		return err
	}

	var startAfter time.Time
	for _, format := range dateTimeFormats {
		startAfter, err = time.Parse(format, dateTimeStr)
		if err == nil {
			break
		}
	}

	if err != nil {
		return fmt.Errorf("invalid date format: %s, must be one of: %s", dateTimeStr, strings.Join(dateTimeFormats, ", "))
	}
	*sa = StartAfter(startAfter)

	return nil
}

func (sa StartAfter) MarshalJSON() ([]byte, error) {
	StartAfterStr := time.Time(sa).Format("2006-01-02T15:04:05")
	return json.Marshal(StartAfterStr)
}
