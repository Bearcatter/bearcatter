package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/image/riff"
)

func main() {
	info, infoErr := DecodeRecording("test.wav")
	if infoErr != nil {
		panic(infoErr)
	}
	spew.Dump(info)
}

// 20200619223802
const TimestampFormat = "20060102150405"

type Recording struct {
	System           string     `mapstructure:"IART"`
	Department       string     `mapstructure:"IGNR"`
	Channel          string     `mapstructure:"INAM"`
	TGIDFreq         string     `mapstructure:"ICMT"`
	Product          string     `mapstructure:"IPRD"`
	Unknown          string     `mapstructure:"IKEY"`
	Timestamp        *time.Time `mapstructure:"ICRD"`
	Tone             string     `mapstructure:"ISRC"`
	Unit             int        `mapstructure:"ITCH"`
	FavoriteListName string     `mapstructure:"ISBJ"`
	Reserved         string     `mapstructure:"ICOP"`
}

func DecodeRecording(file string) (*Recording, error) {
	riff, riffErr := getRiffMap(file)
	if riffErr != nil {
		return nil, riffErr
	}

	rec := &Recording{
		System:           riff["IART"],
		Department:       riff["IGNR"],
		Channel:          riff["INAM"],
		TGIDFreq:         riff["ICMT"],
		Product:          riff["IPRD"],
		Unknown:          riff["IKEY"],
		Tone:             riff["ISRC"],
		FavoriteListName: riff["ISBJ"],
		Reserved:         riff["ICOP"],
	}

	ts, tsErr := time.Parse(TimestampFormat, riff["ICRD"])
	if tsErr != nil {
		return nil, tsErr
	}

	rec.Timestamp = &ts

	if strings.HasPrefix(riff["ITCH"], "UID:") {
		var intErr error
		rec.Unit, intErr = strconv.Atoi(riff["ITCH"][4:])
		if intErr != nil {
			return nil, intErr
		}
	}

	return rec, nil
}

func getRiffMap(file string) (map[string]string, error) {
	fh, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	_, r, err := riff.NewReader(fh)
	if err != nil {
		return nil, err
	}
	chunkID, chunkLen, chunkData, err := r.Next()
	if err != nil {
		return map[string]string{}, err
	}
	if chunkID == riff.LIST {
		_, list, err := riff.NewListReader(chunkLen, chunkData)
		if err != nil {
			return map[string]string{}, err
		}
		infoList := map[string]string{}
		for {
			chunkID, _, chunkData, err := list.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return map[string]string{}, err
			}
			b, err := ioutil.ReadAll(chunkData)
			if err != nil {
				return map[string]string{}, err
			}
			v := strings.TrimRight(string(b), "\x00")
			infoList[fmt.Sprintf("%s", chunkID)] = v
		}
		return infoList, nil
	} else {
		return map[string]string{}, errors.New("chunk is not riff.LIST")
	}
}
