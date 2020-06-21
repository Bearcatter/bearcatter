// Package wavparse provides Uniden Bearcat Scanner WAV file parsing.
// Portions from https://github.com/go-audio/wav
package wavparse

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	riff "github.com/go-audio/riff"
)

const timestampFormat = "20060102150405"

var frequencyRegex = regexp.MustCompile(`(?i)692558682d00000000(\d*)00000200000064646464`)

var (
	// See http://bwfmetaedit.sourceforge.net/listinfo.html
	markerIART    = [4]byte{'I', 'A', 'R', 'T'}
	markerISFT    = [4]byte{'I', 'S', 'F', 'T'}
	markerICRD    = [4]byte{'I', 'C', 'R', 'D'}
	markerICOP    = [4]byte{'I', 'C', 'O', 'P'}
	markerIARL    = [4]byte{'I', 'A', 'R', 'L'}
	markerINAM    = [4]byte{'I', 'N', 'A', 'M'}
	markerIENG    = [4]byte{'I', 'E', 'N', 'G'}
	markerIGNR    = [4]byte{'I', 'G', 'N', 'R'}
	markerIPRD    = [4]byte{'I', 'P', 'R', 'D'}
	markerISRC    = [4]byte{'I', 'S', 'R', 'C'}
	markerISBJ    = [4]byte{'I', 'S', 'B', 'J'}
	markerICMT    = [4]byte{'I', 'C', 'M', 'T'}
	markerITRK    = [4]byte{'I', 'T', 'R', 'K'}
	markerITRKBug = [4]byte{'i', 't', 'r', 'k'}
	markerITCH    = [4]byte{'I', 'T', 'C', 'H'}
	markerIKEY    = [4]byte{'I', 'K', 'E', 'Y'}
	markerIMED    = [4]byte{'I', 'M', 'E', 'D'}
	// cidLIST is the chunk ID for a LIST chunk
	cidLIST = [4]byte{'L', 'I', 'S', 'T'}
	// cidINFO is the chunk ID for an INFO chunk
	cidINFO = []byte{'I', 'N', 'F', 'O'}
	// CIDUnid is the chunk ID for a unid chunk
	CIDUNID = [4]byte{'u', 'n', 'i', 'd'}
)

type Recording struct {
	File     string
	Public   *ListChunk
	Private  *UnidenChunk
	Duration time.Duration
}

type ListChunk struct {
	System           string
	Department       string
	Channel          string
	TGIDFreq         string
	Product          string
	Unknown          string
	Timestamp        *time.Time
	Tone             string
	UnitID           int64
	FavoriteListName string
	Reserved         string
}

type FavoriteInfo struct {
	Name string
	File string
}

type SiteInfo struct {
	Name string
	Type string
}

type LocationInfo struct {
	Latitude  float64
	Longitude float64
	Range     float64
	Type      string
}

type SystemInfo struct {
	Name string
	Type string
}
type UnidenChunk struct {
	FavoriteList FavoriteInfo
	System       SystemInfo
	Department   string
	Channel      string
	TGID         string
	Frequency    float64
	Site         SiteInfo
	UnitID       int64
	Location     LocationInfo
	Debug        []string
}

func DecodeRecording(file string) (*Recording, error) {
	rec := &Recording{
		File: filepath.Base(file),
	}
	f, openErr := os.Open(file)
	if openErr != nil {
		return nil, openErr
	}
	defer f.Close()

	c := riff.New(f)
	if parseHeadersErr := c.ParseHeaders(); parseHeadersErr != nil {
		if parseHeadersErr == io.EOF {
			return nil, fmt.Errorf("file %s is invalid or empty unable to parse headers", file)
		}
		return nil, fmt.Errorf("error parsing headers: %v", parseHeadersErr)
	}

	for {
		chunk, chunkErr := c.NextChunk()
		if chunkErr != nil {
			return nil, chunkErr
		}
		if chunk.ID == riff.FmtID {
			chunk.DecodeWavHeader(c)
		} else if chunk.ID == riff.DataFormatID {
			break
		}

		switch chunk.ID {
		case cidLIST:
			decoded, decodedErr := decodeLISTChunk(chunk)
			if decodedErr != nil {
				return nil, fmt.Errorf("error when decoding riff list chunk: %v", decodedErr)
			}
			rec.Public = decoded
		case CIDUNID:
			decoded, decodedErr := decodeUNIDChunk(chunk)
			if decodedErr != nil {
				return nil, fmt.Errorf("error when decoding riff unid chunk: %v", decodedErr)
			}
			rec.Private = decoded
		}

		chunk.Done()
	}

	duration, durationErr := c.Duration()
	if durationErr != nil {
		return nil, fmt.Errorf("error getting file duration: %v", durationErr)
	}

	rec.Duration = duration

	return rec, nil
}

// decodeLISTChunk decodes a LIST chunk
func decodeLISTChunk(ch *riff.Chunk) (*ListChunk, error) {
	recListChunk := &ListChunk{}

	if ch == nil {
		return recListChunk, fmt.Errorf("can't decode a nil chunk")
	}
	if ch.ID == cidLIST {
		// read the entire chunk in memory
		buf := make([]byte, ch.Size)
		var err error
		if _, err = ch.Read(buf); err != nil {
			return recListChunk, fmt.Errorf("failed to read the LIST chunk - %v", err)
		}
		r := bytes.NewReader(buf)
		// INFO subchunk
		scratch := make([]byte, 4)
		if _, err = r.Read(scratch); err != nil {
			return recListChunk, fmt.Errorf("failed to read the INFO subchunk - %v", err)
		}
		if !bytes.Equal(scratch, cidINFO[:]) {
			// "expected an INFO subchunk but got %s", string(scratch)
			// TODO: support adtl subchunks
			ch.Drain()
			return recListChunk, nil
		}

		// the rest is a list of string entries
		var (
			id   [4]byte
			size uint32
		)
		readSubHeader := func() error {
			if err := binary.Read(r, binary.BigEndian, &id); err != nil {
				return err
			}
			return binary.Read(r, binary.LittleEndian, &size)
		}

		for err == nil {
			err = readSubHeader()
			if err != nil {
				break
			}
			scratch = make([]byte, size)
			r.Read(scratch)
			switch id {
			case markerIART:
				recListChunk.System = nullTermStr(scratch)
			case markerIGNR:
				recListChunk.Department = nullTermStr(scratch)
			case markerINAM:
				recListChunk.Channel = nullTermStr(scratch)
			case markerICMT:
				recListChunk.TGIDFreq = nullTermStr(scratch)
			case markerIPRD:
				recListChunk.Product = nullTermStr(scratch)
			case markerIKEY:
				recListChunk.Unknown = nullTermStr(scratch)
			case markerICRD:
				ts, tsErr := time.Parse(timestampFormat, nullTermStr(scratch))
				if tsErr != nil {
					return nil, tsErr
				}

				recListChunk.Timestamp = &ts
			case markerISRC:
				recListChunk.Tone = nullTermStr(scratch)
			case markerITCH:
				if len(nullTermStr(scratch)) > 0 {
					uid, uidErr := strconv.ParseInt(nullTermStr(scratch)[4:], 10, 64)
					if uidErr != nil {
						return nil, uidErr
					}
					recListChunk.UnitID = uid
				}
			case markerISBJ:
				recListChunk.FavoriteListName = nullTermStr(scratch)
			case markerICOP:
				recListChunk.Reserved = nullTermStr(scratch)
			}
		}
	}
	ch.Drain()
	return recListChunk, nil
}

// decodeUNIDChunk decodes a UNID chunk
func decodeUNIDChunk(ch *riff.Chunk) (*UnidenChunk, error) {
	recUnidChunk := &UnidenChunk{}

	if ch == nil {
		return recUnidChunk, fmt.Errorf("can't decode a nil chunk")
	}
	if ch.ID == CIDUNID {
		// read the entire chunk in memory
		buf := make([]byte, ch.Size)
		_, readErr := ch.Read(buf)
		if readErr != nil {
			return recUnidChunk, fmt.Errorf("failed to read the UNID chunk - %v", readErr)
		}

		split := bytes.Split(buf, []byte("\x00"))

		uidStr := ""

		for _, byt := range split {
			bytStr := nullTermStr(byt)
			if nullTermStr(byt) == "" {
				continue
			}
			if strings.HasPrefix(bytStr, "UID:") {
				uidStr = bytStr[4:]
			}
		}

		lat, latErr := strconv.ParseFloat(nullTermStr(split[26]), 10)
		if latErr != nil {
			log.WithError(latErr).Debugln("Unable to marshal latitude from UNID chunk!")
		}
		lon, lonErr := strconv.ParseFloat(nullTermStr(split[27]), 10)
		if lonErr != nil {
			log.WithError(lonErr).Debugln("Unable to marshal longitude from UNID chunk!")
		}
		ran, ranErr := strconv.ParseFloat(nullTermStr(split[28]), 10)
		if ranErr != nil {
			log.WithError(ranErr).Debugln("Unable to marshal range from UNID chunk!")
		}

		uid := int64(0)
		var uidErr error
		if uidStr != "" {
			uid, uidErr = strconv.ParseInt(uidStr, 10, 64)
		}

		if uidErr != nil {
			log.WithError(uidErr).Warnf("Unable to marshal Unit ID from UNID chunk! %#q\n", uidStr)
		}

		bufHex := make([]byte, hex.EncodedLen(len(buf)))
		hex.Encode(bufHex, buf)

		freq := float64(0.0)
		freqHex := frequencyRegex.FindSubmatch(bufHex)
		var freqErr error
		if len(freqHex) > 0 {
			// There is almost certainly a better way to do this...
			freq, freqErr = strconv.ParseFloat(fmt.Sprintf("%s.%s", freqHex[1][1:4], freqHex[1][4:]), 64)
		}
		if len(freqHex) == 0 || freqErr != nil {
			log.Warnf("Unable to marshal Frequency from UNID chunk %#q! %v", freqHex, freqErr)
		}

		recUnidChunk = &UnidenChunk{
			FavoriteList: FavoriteInfo{
				Name: nullTermStr(split[0]),
				File: nullTermStr(split[1]),
			},
			System: SystemInfo{
				Name: nullTermStr(split[14]),
				Type: nullTermStr(split[17]),
			},
			Department: nullTermStr(split[24]),
			Channel:    nullTermStr(split[32]),
			TGID:       nullTermStr(split[34]),
			Site: SiteInfo{
				Name: nullTermStr(split[42]),
				Type: nullTermStr(split[49]),
			},
			Location: LocationInfo{
				Latitude:  lat,
				Longitude: lon,
				Range:     ran,
				Type:      nullTermStr(split[29]),
			},
			UnitID:    uid,
			Frequency: freq,
			Debug:     strings.Split(string(buf), "\x00"),
		}
	}
	ch.Drain()
	return recUnidChunk, nil
}

func nullTermStr(b []byte) string {
	return string(b[:clen(b)])
}

func clen(n []byte) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}
