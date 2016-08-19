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
	Interval *Interval  `json:"interval"`
	Start    *TimeOfDay `json:"start"`
	Stop     *TimeOfDay `json:"stop"`
	Days     []Weekday  `json:"days"`
}

func (source Source) Validate() error {
	if source.Interval == nil && source.Start == nil && source.Stop == nil {
		return errors.New("must configure either 'interval' or 'start' and 'stop'")
	}

	if source.Start != nil && source.Stop == nil {
		return errors.New("must configure 'stop' if 'start' is set")
	}

	if source.Start == nil && source.Stop != nil {
		return errors.New("must configure 'start' if 'stop' is set")
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

var timeFormats []string

func init() {
	timeFormats = append(timeFormats, "3:04 PM -0700")
	timeFormats = append(timeFormats, "3PM -0700")
	timeFormats = append(timeFormats, "3 PM -0700")
	timeFormats = append(timeFormats, "15:04 -0700")
	timeFormats = append(timeFormats, "1504 -0700")
}

type TimeOfDay time.Duration

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

	t = t.UTC()

	*tod = TimeOfDay(time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute)

	return nil
}

func (tod TimeOfDay) MarshalJSON() ([]byte, error) {
	hours := tod / TimeOfDay(time.Hour)
	minutes := tod % TimeOfDay(time.Hour) / TimeOfDay(time.Minute)
	return json.Marshal(fmt.Sprintf("%d:%02d -0000", hours, minutes))
}

type Weekday time.Weekday

func (x *Weekday) UnmarshalJSON(payload []byte) error {
	var wdStr string
	err := json.Unmarshal(payload, &wdStr)
	if err != nil {
		return err
	}

	var wd time.Weekday
	switch wdStr {
	case "Sunday":
		wd = time.Sunday
	case "Monday":
		wd = time.Monday
	case "Tuesday":
		wd = time.Tuesday
	case "Wednesday":
		wd = time.Wednesday
	case "Thursday":
		wd = time.Thursday
	case "Friday":
		wd = time.Friday
	case "Saturday":
		wd = time.Saturday
	default:
		return fmt.Errorf("unknown weekday: %s", payload)
	}

	*x = Weekday(wd)

	return nil
}

func (wd Weekday) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Weekday(wd).String())
}
