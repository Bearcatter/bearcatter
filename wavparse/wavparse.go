// Package wavparse provides Uniden Bearcat Scanner WAV file parsing.
// Portions from https://github.com/go-audio/wav
package wavparse

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	riff "github.com/go-audio/riff"
)

// ErrNilChunk is returned when a RIFF chunk is nil.
var ErrNilChunk = errors.New("can't decode a nil chunk")

// ErrHeaderParsing is returned when RIFF headers can't be decoded due to a malformed file.
var ErrHeaderParsing = fmt.Errorf("file is invalid or empty when trying to parse riff headers")

// ErrNotUNIDChunk is returned when the chunk given to the UNID decode function is not actually a UNID chunk.
var ErrNotUNIDChunk = errors.New("was given something other than a unid riff chunk")

const timestampFormat = "20060102150405"

var (
	// See http://bwfmetaedit.sourceforge.net/listinfo.html

	// cidLIST is the chunk ID for a LIST chunk
	cidLIST = [4]byte{'L', 'I', 'S', 'T'}
	// cidINFO is the chunk ID for an INFO chunk
	cidINFO = []byte{'I', 'N', 'F', 'O'}
	// cidUNID is the chunk ID for a UNID chunk
	cidUNID = [4]byte{'u', 'n', 'i', 'd'}
)

// DecodeRecording will decode the metadata in the WAV file at the given path.
func DecodeRecording(path string) (*Recording, error) {
	rec := &Recording{
		File: filepath.Base(path),
	}
	f, openErr := os.Open(path)
	if openErr != nil {
		return nil, fmt.Errorf("error when opening wav file: %w", openErr)
	}
	defer f.Close()

	c := riff.New(f)
	if parseHeadersErr := c.ParseHeaders(); parseHeadersErr != nil {
		if parseHeadersErr == io.EOF {
			return nil, ErrHeaderParsing
		}
		return nil, fmt.Errorf("error parsing headers: %w", parseHeadersErr)
	}

	for {
		chunk, chunkErr := c.NextChunk()
		if chunkErr != nil {
			return nil, fmt.Errorf("error when getting next chunk of riff header: %w", chunkErr)
		}
		if chunk.ID == riff.FmtID {
			if decodeErr := chunk.DecodeWavHeader(c); decodeErr != nil {
				return nil, fmt.Errorf("error when decoding wav header: %w", decodeErr)
			}
		} else if chunk.ID == riff.DataFormatID {
			break
		}

		switch chunk.ID {
		case cidLIST:
			decoded, decodedErr := decodeLISTChunk(chunk)
			if decodedErr != nil {
				return nil, fmt.Errorf("error when decoding riff list chunk: %w", decodedErr)
			}
			rec.Public = decoded
		case cidUNID:
			decoded, decodedErr := decodeUNIDChunk(chunk)
			if decodedErr != nil {
				return nil, fmt.Errorf("error when decoding riff unid chunk: %w", decodedErr)
			}
			rec.Private = decoded
		}

		chunk.Done()
	}

	duration, durationErr := c.Duration()
	if durationErr != nil {
		return nil, fmt.Errorf("error getting file duration: %w", durationErr)
	}

	rec.Duration = StopwatchDuration(duration)

	if rec.Public.TGIDFreq == "" && rec.Private.Metadata.TGID != "" {
		rec.Public.TGIDFreq = rec.Private.Metadata.TGID
	}

	if rec.Private.Metadata.TGID == "" && rec.Public.TGIDFreq != "" {
		rec.Private.Metadata.TGID = rec.Public.TGIDFreq
	}

	if rec.Private.Metadata.UnitID == "" && rec.Public.UnitID != "" {
		rec.Private.Metadata.UnitID = rec.Public.UnitID
	}

	return rec, nil
}

// decodeLISTChunk decodes a LIST chunk.
func decodeLISTChunk(ch *riff.Chunk) (*ListChunk, error) {
	recListChunk := &ListChunk{}

	if ch == nil {
		return recListChunk, ErrNilChunk
	}
	if ch.ID == cidLIST {
		// read the entire chunk in memory
		buf := make([]byte, ch.Size)
		var err error
		if _, err = ch.Read(buf); err != nil {
			return recListChunk, fmt.Errorf("failed to read the LIST chunk: %w", err)
		}
		r := bytes.NewReader(buf)
		// INFO subchunk
		scratch := make([]byte, 4)
		if _, err = r.Read(scratch); err != nil {
			return recListChunk, fmt.Errorf("failed to read the INFO subchunk: %w", err)
		}
		if !bytes.Equal(scratch, cidINFO) {
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
				return fmt.Errorf("error when reading riff list chunk id header: %w", err)
			}
			return binary.Read(r, binary.LittleEndian, &size)
		}

		for err == nil {
			err = readSubHeader()
			if err != nil {
				break
			}
			scratch = make([]byte, size)
			if _, readErr := r.Read(scratch); readErr != nil {
				return nil, fmt.Errorf("error while reading value in list chunk: %w", readErr)
			}
			switch id {
			case [4]byte{'I', 'A', 'R', 'T'}: // System
				recListChunk.System = nullTermStr(scratch)
			case [4]byte{'I', 'G', 'N', 'R'}: // Department
				recListChunk.Department = nullTermStr(scratch)
			case [4]byte{'I', 'N', 'A', 'M'}: // Channel
				recListChunk.Channel = nullTermStr(scratch)
			case [4]byte{'I', 'C', 'M', 'T'}: // TGIDFreq
				recListChunk.TGIDFreq = nullTermStr(scratch)
			case [4]byte{'I', 'P', 'R', 'D'}: // Product
				recListChunk.Product = nullTermStr(scratch)
			case [4]byte{'I', 'K', 'E', 'Y'}: // Unknown
				recListChunk.Unknown = nullTermStr(scratch)
			case [4]byte{'I', 'C', 'R', 'D'}: // Timestamp
				ts, tsErr := time.ParseInLocation(timestampFormat, nullTermStr(scratch), time.Local)
				if tsErr != nil {
					return nil, fmt.Errorf("error when parsing timestamp from riff list chunk: %w", tsErr)
				}

				recListChunk.Timestamp = &ts
			case [4]byte{'I', 'S', 'R', 'C'}: // Tone
				recListChunk.Tone = nullTermStr(scratch)
			case [4]byte{'I', 'T', 'C', 'H'}: // UnitID
				if len(nullTermStr(scratch)) > 0 {
					recListChunk.UnitID = nullTermStr(scratch)[4:]
				}
			case [4]byte{'I', 'S', 'B', 'J'}: // FavoriteListName
				recListChunk.FavoriteListName = nullTermStr(scratch)
			case [4]byte{'I', 'C', 'O', 'P'}: // Reserved
				recListChunk.Reserved = nullTermStr(scratch)
			}
		}
	}
	ch.Drain()
	return recListChunk, nil
}

// decodeUNIDChunk decodes a UNID chunk.
func decodeUNIDChunk(ch *riff.Chunk) (*UnidenChunk, error) {
	decodedChunk := &UnidenChunk{}

	if ch == nil {
		return decodedChunk, ErrNilChunk
	}
	if ch.ID != cidUNID {
		return nil, ErrNotUNIDChunk
	}
	rawChunk := RawUnidenChunk{}

	if binErr := ch.ReadBE(&rawChunk); binErr != nil {
		return nil, fmt.Errorf("error when parsing unid chunk as binary: %w", binErr)
	}

	if unmarshalErr := decodedChunk.Favorite.UnmarshalBinary(rawChunk.Favorite[0:len(rawChunk.Favorite)]); unmarshalErr != nil {
		return nil, fmt.Errorf("error when decoding binary to favorite: %w", unmarshalErr)
	}

	if unmarshalErr := decodedChunk.System.UnmarshalBinary(rawChunk.System[0:len(rawChunk.System)]); unmarshalErr != nil {
		return nil, fmt.Errorf("error when decoding binary to system: %w", unmarshalErr)
	}

	if unmarshalErr := decodedChunk.Department.UnmarshalBinary(rawChunk.Department[0:len(rawChunk.Department)]); unmarshalErr != nil {
		return nil, fmt.Errorf("error when decoding binary to department: %w", unmarshalErr)
	}

	if unmarshalErr := decodedChunk.Channel.UnmarshalBinary(rawChunk.Channel[0:len(rawChunk.Channel)]); unmarshalErr != nil {
		return nil, fmt.Errorf("error when decoding binary to channel: %w", unmarshalErr)
	}

	if decodedChunk.System.Type != "Conventional" {
		if unmarshalErr := decodedChunk.Site.UnmarshalBinary(rawChunk.Site[0:len(rawChunk.Site)]); unmarshalErr != nil {
			return nil, fmt.Errorf("error when decoding binary to site: %w", unmarshalErr)
		}

		if unmarshalErr := decodedChunk.Metadata.UnmarshalBinary(rawChunk.Metadata[0:len(rawChunk.Metadata)]); unmarshalErr != nil {
			return nil, fmt.Errorf("error when decoding binary to metadata: %w", unmarshalErr)
		}
	}

	ch.Drain()
	return decodedChunk, nil
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

func parseBool(boolStr string) (bool, error) {
	boolStr = strings.ToLower(boolStr)
	if boolStr == "o" { // Field was truncated, lets just return false.
		return false, nil
	}
	if boolStr == "on" {
		return true, nil
	}
	if boolStr == "off" {
		return false, nil
	}
	return strconv.ParseBool(boolStr)
}
