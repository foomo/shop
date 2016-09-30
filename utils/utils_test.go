package utils

import (
	"log"
	"testing"
)

func TestTimeTransform(t *testing.T) {
	date := "20160307"
	ti, err := GetTimeFromYYYYMMDD(date)
	if err != nil {
		t.Fatal(err)
	}
	date2 := GetFormattedTimeYYYYMMDD(ti)
	log.Println(date, " => ", date2)
	if date != date2 {
		t.Error("date2 is wrong")
	}
}

func TestTimeCompare(t *testing.T) {
	tx, err := GetTimeFromYYYYMMDD("20201224")
	if err != nil {
		t.Fatal(err)
	}
	tStart, err := GetTimeFromYYYYMMDD("20201222")
	if err != nil {
		t.Fatal(err)
	}
	tEnd, err := GetTimeFromYYYYMMDD("20201226")
	if err != nil {
		t.Fatal(err)
	}
	// Should be within lifetime
	if !TimeIsWithinLifeTime(tx, tStart, tEnd) {
		t.Error("expected true")
	}

	tx, err = GetTimeFromYYYYMMDD("20201221")
	if err != nil {
		t.Fatal(err)
	}
	// Should be before lifetime
	if TimeIsWithinLifeTime(tx, tStart, tEnd) {
		t.Error("expected false")
	}
	tx, err = GetTimeFromYYYYMMDD("20201231")
	if err != nil {
		t.Fatal(err)
	}
	// Should be after lifetime
	if TimeIsWithinLifeTime(tx, tStart, tEnd) {
		t.Error("expected false")
	}
}
