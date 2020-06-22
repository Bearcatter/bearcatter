package wavparse

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/stretchr/testify/assert"
)

type WavPlayerTime struct {
	time.Time
}

const wavPlayerTimeFormat = "1/02/2006 03:04:05 PM"

// Convert the internal date as CSV string
func (date *WavPlayerTime) MarshalCSV() (string, error) {
	return date.Time.Format(wavPlayerTimeFormat), nil
}

// Convert the CSV string as internal date
func (date *WavPlayerTime) UnmarshalCSV(csv string) (err error) {
	date.Time, err = time.Parse(wavPlayerTimeFormat, csv)
	return err
}

type WavPlayerDuration struct {
	time.Duration
}

// Convert the internal duration as CSV string
func (clock *WavPlayerDuration) MarshalCSV() (string, error) {
	return clock.Duration.String(), nil
}

// Convert the CSV string as internal duration
func (clock *WavPlayerDuration) UnmarshalCSV(csv string) (err error) {
	split := strings.Split(csv, ":")
	clock.Duration, err = time.ParseDuration(fmt.Sprintf("%sh%sm%ss", split[0], split[1], split[2]))
	return err
}

type WavPlayerEntry struct {
	FilePath       string            `csv:"File path"`
	FileName       string            `csv:"File name"`
	Product        string            `csv:"Scanner type"`
	DateAndTime    WavPlayerTime     `csv:"Date and time"`
	Duration       WavPlayerDuration `csv:"Duration"`
	ScanMode       string            `csv:"Scan mode"`
	SystemType     string            `csv:"Type"`
	Frequency      float64           `csv:"Frequency"`
	Code           string            `csv:"Code"`
	FavoriteName   string            `csv:"Favorite name"`
	SystemName     string            `csv:"System name"`
	DepartmentName string            `csv:"Department name"`
	ChannelName    string            `csv:"Channel name"`
	SiteName       string            `csv:"Site"`
	TGID           string            `csv:"TGID"`
	UnitID         int64             `csv:"UID"`
	UnitIDName     string            `csv:"UID Name"`
	Latitude       float64           `csv:"Latitude"`
	Longitude      float64           `csv:"Longitude"`
}

func TestDecodeRecording(t *testing.T) {
	testCaseFile, openErr := os.OpenFile("fixtures.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if openErr != nil {
		panic(openErr)
	}
	defer testCaseFile.Close()

	testCases := []*WavPlayerEntry{}

	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = ';'
		return r
	})

	if unmarshalErr := gocsv.UnmarshalFile(testCaseFile, &testCases); unmarshalErr != nil { // Load WavPlayerEntry from file
		panic(unmarshalErr)
	}

	for _, testCase := range testCases {
		if testCase == nil {
			t.Log("Refusing to run a nil parsed fixture")
			continue
		}
		t.Run(testCase.FileName, testDecodeRecordingCase(fmt.Sprintf("fixtures/%s", testCase.FileName), *testCase))
	}
}

func testDecodeRecordingCase(path string, expected WavPlayerEntry) func(t *testing.T) {
	return func(t *testing.T) {
		assert := assert.New(t)

		parsed, parsedErr := DecodeRecording(path)
		if parsedErr != nil {
			t.Fatalf("error when parsing file: %v", parsedErr)
		}

		assert.Equal(parsed.File, expected.FileName, "File names should be equal")

		assert.Equal(parsed.Duration, expected.Duration.Duration, "Duration should be equal")

		assert.Equal(parsed.Public.Product, expected.Product, "Products (public) should be equal")

		assert.Equal(parsed.Public.Timestamp, &expected.DateAndTime.Time, "Timestamps (public) should be equal")

		assert.Equal(parsed.Private.System.Type, expected.SystemType, "System types (private) should be equal")

		assert.Equal(parsed.Private.Frequency, expected.Frequency, "Frequencies (private) should be equal")

		assert.Equal(parsed.Public.FavoriteListName, expected.FavoriteName, "Favorite List Names (public) should be equal")
		assert.Equal(parsed.Private.FavoriteList.Name, expected.FavoriteName, "Favorite List Names (private) should be equal")

		assert.Equal(parsed.Public.System, expected.SystemName, "System Names (public) should be equal")
		assert.Equal(parsed.Private.System.Name, expected.SystemName, "System Names (private) should be equal")

		assert.Equal(parsed.Public.Department, expected.DepartmentName, "Department Names (public) should be equal")
		assert.Equal(parsed.Private.Department, expected.DepartmentName, "Department Names (private) should be equal")

		assert.Equal(parsed.Public.Channel, expected.ChannelName, "Channel Names (public) should be equal")
		assert.Equal(parsed.Private.Channel, expected.ChannelName, "Channel Names (private) should be equal")

		assert.Equal(parsed.Private.Site.Name, expected.SiteName, "Site Names (private) should be equal")

		assert.Equal(parsed.Public.TGIDFreq, expected.TGID, "TGID (public) should be equal")
		assert.Equal(parsed.Private.TGID, expected.TGID, "TGID (private) should be equal")

		assert.Equal(parsed.Public.UnitID, expected.UnitID, "UnitID (public) should be equal")
		assert.Equal(parsed.Private.UnitID, expected.UnitID, "UnitID (private) should be equal")

		assert.Equal(parsed.Private.Location.Latitude, expected.Latitude, "Latitude (private) should be equal")
		assert.Equal(parsed.Private.Location.Longitude, expected.Longitude, "Longitude (private) should be equal")
	}
}
