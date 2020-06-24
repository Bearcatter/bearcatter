package wavparse_test

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Bearcatter/bearcatter/wavparse"
	"github.com/davecgh/go-spew/spew"
	v10 "github.com/go-playground/validator/v10"
	"github.com/gocarina/gocsv"
	"github.com/stretchr/testify/assert"
)

type WavPlayerTime struct {
	time.Time
}

const wavPlayerTimeFormat = "1/02/2006 3:04:05 PM"

// Convert the internal date as CSV string.
func (date *WavPlayerTime) MarshalCSV() (string, error) {
	return date.Time.Format(wavPlayerTimeFormat), nil
}

// Convert the CSV string as internal date.
func (date *WavPlayerTime) UnmarshalCSV(csv string) (err error) {
	date.Time, err = time.ParseInLocation(wavPlayerTimeFormat, csv, time.Local)
	return err
}

type WavPlayerDuration struct {
	time.Duration
}

// Convert the internal duration as CSV string.
func (clock *WavPlayerDuration) MarshalCSV() (string, error) {
	return clock.Duration.String(), nil
}

// Convert the CSV string as internal duration.
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
	testCaseFile, openErr := os.OpenFile("fixtures.csv", os.O_RDONLY, os.ModePerm)
	if openErr != nil {
		t.Fatalf("error when opening csv containing testcases: %v", openErr)
	}
	defer testCaseFile.Close()

	testCases := []*WavPlayerEntry{}

	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = ';'
		return r
	})

	if unmarshalErr := gocsv.UnmarshalFile(testCaseFile, &testCases); unmarshalErr != nil { // Load WavPlayerEntry from file
		t.Fatalf("error when unmarshalling csv containing testcases: %v", unmarshalErr)
	}

	validator := v10.New()

	for _, testCase := range testCases {
		if testCase == nil {
			t.Fatal("Refusing to run a nil fixture")
			continue
		}

		t.Run(testCase.FileName, testDecode(fmt.Sprintf("fixtures/%s", testCase.FileName), *testCase, validator))
	}
}

func testDecode(path string, testCase WavPlayerEntry, validator *v10.Validate) func(t *testing.T) {
	return func(t *testing.T) {
		parsed, parsedErr := wavparse.DecodeRecording(path)
		if parsedErr != nil {
			t.Fatalf("error when parsing file: %v", parsedErr)
		}

		if parsed == nil {
			t.Fatal("Refusing to run a nil decoded fixed")
		}

		t.Run("SimpleEquality", testEquality(*parsed, testCase))
		t.Run("UnitID", testUnitIDEquality(*parsed, testCase))
		t.Run("Validate", testValidation(*parsed, validator))
	}
}

func testEquality(parsed wavparse.Recording, expected WavPlayerEntry) func(t *testing.T) {
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

var badUnitIDFiles = []string{"2020-06-21_12-15-23.wav", "2020-06-21_00-01-50.wav", "2020-06-21_00-00-13.wav", "2020-06-21_12-15-31.wav", "2020-06-21_00-06-49.wav", "2020-06-21_00-02-04.wav", "2020-06-20_23-59-24.wav", "2020-06-20_23-59-15.wav", "2020-06-21_12-15-58.wav", "2020-06-21_00-01-17.wav", "2020-06-21_00-06-45.wav", "2020-06-21_00-01-58.wav", "2020-06-21_00-06-52.wav", "2020-06-21_12-15-28.wav", "2020-06-21_00-07-02.wav", "2020-06-21_00-06-56.wav", "2020-06-21_00-06-54.wav", "2020-06-21_18-00-59.wav", "2020-06-21_18-09-41.wav", "2020-06-21_18-19-44.wav", "2020-06-21_17-40-38.wav", "2020-06-21_18-31-44.wav", "2020-06-21_16-11-02.wav", "2020-06-21_17-38-20.wav", "2020-06-21_17-22-46.wav", "2020-06-21_18-29-27.wav", "2020-06-21_17-19-13.wav", "2020-06-21_17-42-17.wav", "2020-06-21_18-18-13.wav", "2020-06-21_17-14-53.wav", "2020-06-21_18-09-56.wav", "2020-06-21_16-14-25.wav", "2020-06-21_18-19-46.wav", "2020-06-21_17-20-56.wav", "2020-06-21_18-14-06.wav", "2020-06-21_16-11-29.wav", "2020-06-21_17-45-33.wav", "2020-06-22_18-53-23.wav", "2020-06-21_18-37-22.wav", "2020-06-21_16-15-49.wav", "2020-06-21_18-23-43.wav", "2020-06-21_18-09-47.wav", "2020-06-21_17-33-12.wav", "2020-06-21_18-19-57.wav", "2020-06-21_17-20-53.wav", "2020-06-21_17-50-07.wav", "2020-06-21_18-13-22.wav", "2020-06-21_16-22-09.wav", "2020-06-21_18-30-02.wav", "2020-06-21_17-28-37.wav", "2020-06-21_17-42-07.wav", "2020-06-21_18-06-03.wav", "2020-06-21_17-42-11.wav", "2020-06-21_17-45-24.wav", "2020-06-21_18-18-15.wav", "2020-06-21_18-23-40.wav", "2020-06-21_17-55-21.wav", "2020-06-21_18-14-29.wav", "2020-06-21_18-17-44.wav", "2020-06-21_18-25-09.wav", "2020-06-21_16-14-22.wav", "2020-06-21_18-31-41.wav", "2020-06-21_18-25-20.wav", "2020-06-21_18-21-50.wav", "2020-06-21_17-59-08.wav", "2020-06-21_16-10-53.wav", "2020-06-21_17-50-04.wav", "2020-06-21_18-13-09.wav", "2020-06-22_18-53-09.wav", "2020-06-21_18-05-45.wav", "2020-06-21_17-21-38.wav", "2020-06-21_18-15-40.wav", "2020-06-21_18-01-09.wav", "2020-06-21_17-22-41.wav", "2020-06-21_16-23-05.wav", "2020-06-21_17-56-02.wav", "2020-06-21_16-34-09.wav", "2020-06-21_18-02-03.wav", "2020-06-21_17-13-06.wav", "2020-06-21_17-48-16.wav", "2020-06-21_17-27-16.wav", "2020-06-21_18-38-02.wav", "2020-06-21_16-55-24.wav", "2020-06-21_18-21-22.wav", "2020-06-21_18-13-53.wav", "2020-06-21_17-45-57.wav", "2020-06-22_23-07-26.wav", "2020-06-21_18-12-39.wav", "2020-06-21_17-58-10.wav", "2020-06-21_18-20-48.wav", "2020-06-21_18-05-35.wav", "2020-06-21_17-26-41.wav", "2020-06-21_17-13-11.wav", "2020-06-21_16-22-52.wav", "2020-06-21_18-17-09.wav", "2020-06-21_18-31-19.wav", "2020-06-21_18-10-01.wav", "2020-06-21_18-21-09.wav", "2020-06-21_18-21-35.wav", "2020-06-21_18-07-25.wav", "2020-06-21_17-48-00.wav", "2020-06-21_18-11-41.wav", "2020-06-21_16-23-07.wav", "2020-06-21_18-08-58.wav", "2020-06-21_17-40-48.wav", "2020-06-21_17-47-41.wav", "2020-06-21_17-27-39.wav", "2020-06-21_17-57-51.wav", "2020-06-21_17-46-00.wav", "2020-06-21_18-02-10.wav", "2020-06-21_17-56-05.wav", "2020-06-21_17-38-45.wav", "2020-06-21_17-38-47.wav", "2020-06-21_18-02-06.wav", "2020-06-21_17-39-13.wav", "2020-06-21_17-18-36.wav", "2020-06-21_16-55-08.wav", "2020-06-21_17-30-22.wav", "2020-06-21_16-53-55.wav", "2020-06-21_16-20-44.wav", "2020-06-21_17-56-12.wav", "2020-06-21_18-12-16.wav", "2020-06-21_18-34-37.wav", "2020-06-21_18-16-43.wav", "2020-06-21_16-55-05.wav", "2020-06-21_16-45-01.wav", "2020-06-21_18-21-17.wav", "2020-06-21_18-28-33.wav", "2020-06-21_17-43-03.wav", "2020-06-21_16-49-01.wav", "2020-06-21_18-20-43.wav", "2020-06-21_16-59-10.wav", "2020-06-21_18-12-30.wav", "2020-06-21_18-29-59.wav", "2020-06-21_16-44-57.wav", "2020-06-21_18-02-21.wav", "2020-06-21_17-14-05.wav", "2020-06-21_18-37-59.wav", "2020-06-21_18-28-31.wav", "2020-06-21_16-17-23.wav", "2020-06-21_18-19-39.wav", "2020-06-21_18-38-09.wav", "2020-06-21_17-47-58.wav", "2020-06-21_17-15-44.wav", "2020-06-21_17-49-49.wav", "2020-06-21_18-31-05.wav", "2020-06-21_16-26-02.wav", "2020-06-21_18-16-41.wav", "2020-06-21_16-59-07.wav", "2020-06-21_17-51-14.wav", "2020-06-21_18-12-25.wav", "2020-06-21_18-08-55.wav", "2020-06-21_16-44-46.wav", "2020-06-21_16-34-06.wav", "2020-06-21_18-00-35.wav", "2020-06-21_16-56-52.wav", "2020-06-21_18-28-20.wav", "2020-06-21_17-49-59.wav", "2020-06-21_17-15-41.wav", "2020-06-21_16-46-42.wav", "2020-06-21_16-48-53.wav", "2020-06-21_18-26-18.wav", "2020-06-21_17-52-40.wav", "2020-06-21_17-46-21.wav", "2020-06-21_18-05-38.wav", "2020-06-21_17-17-44.wav", "2020-06-21_18-33-11.wav", "2020-06-21_18-20-45.wav", "2020-06-21_16-24-02.wav", "2020-06-21_18-12-36.wav", "2020-06-21_18-33-07.wav", "2020-06-21_18-09-06.wav", "2020-06-21_18-07-17.wav", "2020-06-21_17-33-53.wav", "2020-06-21_18-31-16.wav", "2020-06-21_16-46-40.wav", "2020-06-21_16-48-50.wav", "2020-06-21_16-28-01.wav", "2020-06-21_17-16-07.wav", "2020-06-21_17-21-46.wav", "2020-06-21_16-23-34.wav", "2020-06-21_17-38-14.wav", "2020-06-21_17-58-51.wav", "2020-06-21_16-16-17.wav", "2020-06-21_17-45-28.wav", "2020-06-21_18-18-19.wav", "2020-06-21_17-28-39.wav", "2020-06-22_18-53-38.wav", "2020-06-21_16-22-07.wav", "2020-06-21_17-24-05.wav", "2020-06-21_18-41-18.wav", "2020-06-21_17-47-38.wav", "2020-06-21_17-50-20.wav", "2020-06-21_18-43-20.wav", "2020-06-21_18-15-58.wav", "2020-06-21_18-12-51.wav", "2020-06-21_17-38-17.wav", "2020-06-21_18-01-13.wav", "2020-06-21_18-01-07.wav", "2020-06-21_18-08-23.wav", "2020-06-21_17-21-36.wav", "2020-06-21_17-13-53.wav", "2020-06-21_18-41-33.wav", "2020-06-21_18-35-03.wav", "2020-06-21_18-32-36.wav", "2020-06-21_17-54-46.wav", "2020-06-21_16-51-28.wav", "2020-06-21_16-22-05.wav", "2020-06-21_17-52-32.wav", "2020-06-21_16-11-08.wav", "2020-06-21_18-06-27.wav", "2020-06-21_17-36-03.wav", "2020-06-21_18-01-16.wav", "2020-06-21_17-31-22.wav", "2020-06-21_16-44-31.wav", "2020-06-21_18-30-22.wav", "2020-06-21_16-54-34.wav", "2020-06-21_17-23-22.wav", "2020-06-21_17-37-43.wav", "2020-06-21_16-45-59.wav", "2020-06-21_16-14-29.wav", "2020-06-21_17-54-42.wav", "2020-06-21_18-36-56.wav", "2020-06-21_16-10-59.wav", "2020-06-21_18-41-37.wav", "2020-06-21_16-22-14.wav", "2020-06-21_18-12-41.wav", "2020-06-21_17-45-11.wav", "2020-06-21_17-41-49.wav", "2020-06-21_16-54-37.wav", "2020-06-21_18-14-34.wav", "2020-06-21_16-22-02.wav", "2020-06-21_18-09-59.wav", "2020-06-21_18-41-21.wav", "2020-06-21_17-29-40.wav", "2020-06-21_17-54-40.wav", "2020-06-21_18-31-48.wav", "2020-06-21_18-03-05.wav", "2020-06-21_17-29-41.wav", "2020-06-21_17-57-04.wav", "2020-06-21_16-42-46.wav", "2020-06-22_18-51-11.wav", "2020-06-21_18-00-54.wav", "2020-06-21_17-48-50.wav", "2020-06-21_17-52-34.wav", "2020-06-21_17-17-30.wav", "2020-06-21_16-44-27.wav", "2020-06-21_18-06-09.wav", "2020-06-21_18-08-18.wav", "2020-06-21_18-06-38.wav"}

func testUnitIDEquality(parsed wavparse.Recording, expected WavPlayerEntry) func(t *testing.T) {
	return func(t *testing.T) {
		assert := assert.New(t)

		assert.Equal(expected.UnitID, parsed.Public.UnitID, "UnitID (public) should be equal to expected")

		if contains(badUnitIDFiles, expected.FileName) {
			t.Skipf("Skipping UnitID equality because %s has a bad private UnitID\n", expected.FileName)
		}

		assert.Equal(expected.UnitID, parsed.Private.Metadata.UnitID, "UnitID (private) should be equal to expected")
	}
}

func testValidation(parsed wavparse.Recording, validator *v10.Validate) func(t *testing.T) {
	return func(t *testing.T) {
		if err := validator.Struct(parsed); err != nil {
			t.Logf("error when validating struct: %v\n", err)
			t.Logf("spew: %s\n", spew.Sdump(parsed))
			t.FailNow()
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
