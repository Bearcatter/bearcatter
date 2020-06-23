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
	UnitID         string            `csv:"UID"`
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
			t.Fatal("Refusing to run a nil fixture")
			continue
		}

		// if testCase.FileName != "2020-06-21_12-15-58.wav" {
		// 	continue
		// }

		t.Run(testCase.FileName, testDecode(fmt.Sprintf("fixtures/%s", testCase.FileName), *testCase))
	}
}

func testDecode(path string, testCase WavPlayerEntry) func(t *testing.T) {
	return func(t *testing.T) {
		parsed, parsedErr := DecodeRecording(path)
		if parsedErr != nil {
			t.Fatalf("error when parsing file: %v", parsedErr)
		}

		if parsed == nil {
			t.Fatal("Refusing to run a nil decoded fixed")
		}

		t.Run("SimpleEquality", testEquality(*parsed, testCase))
		t.Run("UnitID", testUnitIDEquality(*parsed, testCase))
	}
}

var badUnitIDFiles = []string{"2020-06-21_12-15-23.wav", "2020-06-21_00-01-50.wav", "2020-06-21_00-00-13.wav", "2020-06-21_12-15-31.wav", "2020-06-21_00-06-49.wav", "2020-06-21_00-02-04.wav", "2020-06-20_23-59-24.wav", "2020-06-20_23-59-15.wav", "2020-06-21_12-15-58.wav", "2020-06-21_00-01-17.wav", "2020-06-21_00-06-45.wav", "2020-06-21_00-01-58.wav", "2020-06-21_00-06-52.wav", "2020-06-21_12-15-28.wav", "2020-06-21_00-07-02.wav", "2020-06-21_00-06-56.wav", "2020-06-21_00-06-54.wav"}

func testUnitIDEquality(parsed Recording, expected WavPlayerEntry) func(t *testing.T) {
	return func(t *testing.T) {
		assert := assert.New(t)

		assert.Equal(expected.UnitID, parsed.Public.UnitID, "UnitID (public) should be equal to expected")

		if contains(badUnitIDFiles, expected.FileName) {
			t.Skipf("Skipping UnitID equality because %s has a bad private UnitID\n", expected.FileName)
		}

		assert.Equal(expected.UnitID, parsed.Private.Metadata.UnitID, "UnitID (private) should be equal to expected")
	}
}

func testEquality(parsed Recording, expected WavPlayerEntry) func(t *testing.T) {
	return func(t *testing.T) {
		assert := assert.New(t)

		assert.Equal(expected.FileName, parsed.File, "File names should be equal to expected")

		if expected.Product != "" {
			assert.Equal(expected.Product, parsed.Public.Product, "Products (public) should be equal to expected")
		}

		if !expected.DateAndTime.Time.IsZero() {
			assert.Equal(&expected.DateAndTime.Time, parsed.Public.Timestamp, "Timestamps (public) should be equal to expected")
		}

		if expected.SystemType != "" {
			assert.Equal(expected.SystemType, parsed.Private.System.Type, "System types (private) should be equal to expected")
		}

		if expected.SiteName != "" {
			assert.Equal(expected.SiteName, parsed.Private.Site.Name, "Site Names (private) should be equal to expected")
		}

		if expected.Latitude != float64(0.0) {
			assert.Equal(expected.Latitude, parsed.Private.Department.Latitude, "Latitude (private) should be equal to expected")
		}

		if expected.Longitude != float64(0.0) {
			assert.Equal(expected.Longitude, parsed.Private.Department.Longitude, "Longitude (private) should be equal to expected")
		}

		if expected.FavoriteName != "" {
			assert.Equal(expected.FavoriteName, parsed.Public.FavoriteListName, "Favorite List Names (public) should be equal to expected")
			assert.Equal(expected.FavoriteName, parsed.Private.Favorite.Name, "Favorite List Names (private) should be equal to expected")
		}

		if expected.SystemName != "" {
			assert.Equal(expected.SystemName, parsed.Public.System, "System Names (public) should be equal to expected")
			assert.Equal(expected.SystemName, parsed.Private.System.Name, "System Names (private) should be equal to expected")
		}

		if expected.DepartmentName != "" {
			assert.Equal(expected.DepartmentName, parsed.Public.Department, "Department Names (public) should be equal to expected")
			assert.Equal(expected.DepartmentName, parsed.Private.Department.Name, "Department Names (private) should be equal to expected")
		}

		if expected.ChannelName != "" {
			assert.Equal(expected.ChannelName, parsed.Public.Channel, "Channel Names (public) should be equal to expected")
			assert.Equal(expected.ChannelName, parsed.Private.Channel.Name, "Channel Names (private) should be equal to expected")
		}

		if expected.SystemType != "Conventional" {
			assert.Equal(expected.Frequency, parsed.Private.Metadata.Frequency, "Frequencies (private) should be equal to expected")
		}

		if expected.TGID != "" {
			assert.Equal(expected.TGID, parsed.Public.TGIDFreq, "TGID (public) should be equal to expected")
			assert.Equal(expected.TGID, parsed.Private.Metadata.TGID, "TGID (private) should be equal to expected")
		}
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
