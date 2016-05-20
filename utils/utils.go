package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"time"
)

// Returns string in format YYYYMMDD
func GetDateYYYYMMDD() string {
	return time.Now().Format("20060102")
}

func GetFormattedTime(t time.Time) string {
	return t.Format("Mon Jan 2 2006 15:04:05")
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
