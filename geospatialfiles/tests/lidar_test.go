package tests

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/jblindsay/go-spatial/geospatialfiles/lidar"
)

var testLidarRead = true

func TestLidarRead(t *testing.T) {
	if testLidarRead {
		inFile := "/Users/johnlindsay/Documents/Data/Half Dome/points.las"
		input, err := lidar.CreateFromFile(inFile)
		if err != nil {
			println(err.Error())
		}
		defer input.Close()

		var buffer bytes.Buffer

		buffer.WriteString(fmt.Sprintf("File Name: %v\n", input.GetFileName()))
		day, month := convertYearday(int(input.Header.FileCreationDay), int(input.Header.FileCreationYear))
		buffer.WriteString(fmt.Sprintf("Creation Date: %v %v, %v\n", day, month, input.Header.FileCreationYear))
		buffer.WriteString(fmt.Sprintf("Generating software: %v\n", input.Header.GeneratingSoftware))
		buffer.WriteString(fmt.Sprintf("LAS version: %v.%v\n", input.Header.VersionMajor, input.Header.VersionMajor))
		buffer.WriteString(fmt.Sprintf("Number of Points: %v\n", input.Header.NumberPoints))
		buffer.WriteString(fmt.Sprintf("Point record length: %v\n", input.Header.PointRecordLength))
		buffer.WriteString(fmt.Sprintf("Point record format: %v\n", input.Header.PointFormatID))

		buffer.WriteString(fmt.Sprintf("Point record format: %v\n", input.Header.String()))

		println(buffer.String())

	} else {
		t.SkipNow()
	}
}

func convertYearday(yday int, year int) (int, string) {
	var months = [...]string{
		"January",
		"February",
		"March",
		"April",
		"May",
		"June",
		"July",
		"August",
		"September",
		"October",
		"November",
		"December",
	}

	var lastDayOfMonth = [...]int{
		31,
		59,
		90,
		120,
		151,
		181,
		212,
		243,
		273,
		304,
		334,
		365,
	}

	var month int
	day := yday
	if isLeapYear(year) {
		// Leap year
		switch {
		case day > 31+29-1:
			// After leap day; pretend it wasn't there.
			day--
		case day == 31+29-1:
			// Leap day.
			month = 1
			day = 29
			return day, months[month]
		}
	}

	month = 0
	for m := range lastDayOfMonth {
		//for m := 0; m < 12; m++ {
		if day > lastDayOfMonth[m] {
			month++
		} else {
			break
		}
	}
	if month > 0 {
		day -= lastDayOfMonth[month-1]
	}

	return day, months[month]
}

func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
