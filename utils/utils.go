package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"time"
)

func GetTimeHHMMSS(t time.Time) string {
	return t.Format("150405")
}

var CET, _ = time.LoadLocation("Europe/Zurich")

func TimeIsOnSameDay(date time.Time, refDate time.Time) (bool, error) {
	// get time for reference date 0:00
	sameDay := GetZeroTimeForDay(refDate)

	// get time for next day 0:00
	nextDay := sameDay.Add(time.Hour * 24)

	if (date.Equal(sameDay) || date.After(sameDay)) && date.Before(nextDay) {
		return true, nil
	}

	return false, nil
}

func TimeIsWithinLifeTime(date time.Time, start time.Time, end time.Time) bool {
	if date.Equal(start) || date.Equal(end) || (date.After(start) && date.Before(end)) {
		return true
	}
	return false
}
func TimeIsWithinLifeTimeYYYY_MM_DD(date time.Time, start string, end string) (bool, error) {
	startTime, err := GetTimeFromYYY_MM_DD(start)
	if err != nil {
		return false, err
	}
	endTime, err := GetTimeFromYYY_MM_DD(end)
	if err != nil {
		return false, err
	}

	if date.Equal(startTime) || date.Equal(endTime) || (date.After(startTime) && date.Before(endTime)) {
		return true, nil
	}
	return false, nil
}

// Returns string in format YYYYMMDD
func GetDateYYYYMMDD() string {
	return time.Now().Format("20060102")
}
func GetDateYYYY_MM_DD() string {
	return time.Now().Format("2006-01-02")
}

func GetFormattedTimeYYYYMMDD(t time.Time) string {
	return t.Format("20060102")
}
func GetFormattedTime(t time.Time) string {
	return t.Format("Mon Jan 2 2006 15:04:05")
}
func GetFormattedTimeYYYY_MM_DD(t time.Time) string {
	return t.Format("2006-01-02")
}
func GetFormattedTimeShort(t time.Time) string {
	return t.Format("Mon Jan 2 2006 150405")
}

// GetZeroTimeForDay returns the time for 0:00 for the given date
func GetZeroTimeForDay(date time.Time) time.Time {
	year, month, day := date.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, CET)
}

func GetTimeFromYYYYMMDD(date string) (time.Time, error) {
	if len(date) != 8 {
		return time.Time{}, errors.New("date string has unexpected length")
	}

	yearStr := date[0:4]
	monthStr := date[4:6]
	dayStr := date[6:8]

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return time.Time{}, err
	}
	month, err := strconv.Atoi(monthStr)
	if err != nil {
		return time.Time{}, err
	}
	day, err := strconv.Atoi(dayStr)
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, CET), nil

}
func GetTimeFromYYY_MM_DD(date string) (time.Time, error) {
	if len(date) != 10 {
		return time.Time{}, errors.New("date string has unexpected length")
	}

	yearStr := date[0:4]
	monthStr := date[5:7]
	dayStr := date[8:10]

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return time.Time{}, err
	}
	month, err := strconv.Atoi(monthStr)
	if err != nil {
		return time.Time{}, err
	}
	day, err := strconv.Atoi(dayStr)
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, CET), nil
}

func TimeNow() time.Time {
	// Trunacte makes sure that the exact same value that goes in the db,
	// will come out of it again (otherwise mongo will cut precison and persistence tests will fail)
	return time.Now().Truncate(time.Second)
}

func PrintJSON(v interface{}) {
	fmt.Println(ToJSON(v))
}

// ToJSON creates a JSON from v and returns it as string
func ToJSON(v interface{}) string {
	json, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return "Could not print JSON"
	}
	return string(json)
}

func GetFunctionName(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

// OpenInBrowser stores the html under "<name><date>.html" in the current
// directory and opens it in the browser
func OpenInBrowser(name string, html string) error {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("PWD", pwd)

	//path := "file://" + pwd + "/" + "diff.html"
	path := pwd + "/" + "diff-" + name + "_" + GetFormattedTimeShort(time.Now()) + ".html"
	fmt.Println("PATH", path)
	tmpFile, err := os.Create(path)
	if err != nil {
		return err
	}
	tmpFile.WriteString(html)
	tmpFile.Close()

	return err

}

// Ite => if then else. If true returns thenDo else elseDo
func IteString(condition bool, thenDo string, elseDo string) string {
	if condition {
		return thenDo
	} else {
		return elseDo
	}
}

func Round(input float64, decimals int) float64 {
	input = input * math.Pow10(decimals)

	if input < 0 {
		return math.Ceil(input-0.5) / math.Pow10(decimals)
	}
	return math.Floor(input+0.5) / math.Pow10(decimals)
}
