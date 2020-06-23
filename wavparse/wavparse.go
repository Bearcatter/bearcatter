// Package wavparse provides Uniden Bearcat Scanner WAV file parsing.
// Portions from https://github.com/go-audio/wav
package wavparse

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	riff "github.com/go-audio/riff"
)

const timestampFormat = "20060102150405"

var (
	// See http://bwfmetaedit.sourceforge.net/listinfo.html
	markerIART = [4]byte{'I', 'A', 'R', 'T'} // System
	markerICRD = [4]byte{'I', 'C', 'R', 'D'} // Department
	markerICOP = [4]byte{'I', 'C', 'O', 'P'} // Channel
	markerINAM = [4]byte{'I', 'N', 'A', 'M'} // TGIDFreq
	markerIGNR = [4]byte{'I', 'G', 'N', 'R'} // Product
	markerIPRD = [4]byte{'I', 'P', 'R', 'D'} // Unknown
	markerISRC = [4]byte{'I', 'S', 'R', 'C'} // Timestamp
	markerISBJ = [4]byte{'I', 'S', 'B', 'J'} // Tone
	markerICMT = [4]byte{'I', 'C', 'M', 'T'} // UnitID
	markerITCH = [4]byte{'I', 'T', 'C', 'H'} // FavoriteListName
	markerIKEY = [4]byte{'I', 'K', 'E', 'Y'} // Reserved

	// cidLIST is the chunk ID for a LIST chunk
	cidLIST = [4]byte{'L', 'I', 'S', 'T'}
	// cidINFO is the chunk ID for an INFO chunk
	cidINFO = []byte{'I', 'N', 'F', 'O'}
	// CIDUnid is the chunk ID for a unid chunk
	CIDUNID = [4]byte{'u', 'n', 'i', 'd'}
)

// DecodeRecording will decode the metadata in the WAV file at the given path.
func DecodeRecording(path string) (*Recording, error) {
	rec := &Recording{
		File: filepath.Base(path),
	}
	f, openErr := os.Open(path)
	if openErr != nil {
		return nil, fmt.Errorf("error when opening wav file: %v", openErr)
	}
	defer f.Close()

	c := riff.New(f)
	if parseHeadersErr := c.ParseHeaders(); parseHeadersErr != nil {
		if parseHeadersErr == io.EOF {
			return nil, fmt.Errorf("file %s is invalid or empty unable to parse headers", path)
		}
		return nil, fmt.Errorf("error parsing headers: %v", parseHeadersErr)
	}

	for {
		chunk, chunkErr := c.NextChunk()
		if chunkErr != nil {
			return nil, fmt.Errorf("error when getting next chunk of riff header: %v", chunkErr)
		}
		if chunk.ID == riff.FmtID {
			if decodeErr := chunk.DecodeWavHeader(c); decodeErr != nil {
				return nil, fmt.Errorf("error when decoding wav header: %v", decodeErr)
			}
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
			return recListChunk, fmt.Errorf("failed to read the LIST chunk: %v", err)
		}
		r := bytes.NewReader(buf)
		// INFO subchunk
		scratch := make([]byte, 4)
		if _, err = r.Read(scratch); err != nil {
			return recListChunk, fmt.Errorf("failed to read the INFO subchunk: %v", err)
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
				return fmt.Errorf("error when reading riff list chunk id header: %v", err)
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
				return nil, fmt.Errorf("error while reading value in list chunk: %v", readErr)
			}
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
					return nil, fmt.Errorf("error when parsing timestamp from riff list chunk: %v", tsErr)
				}

				recListChunk.Timestamp = &ts
			case markerISRC:
				recListChunk.Tone = nullTermStr(scratch)
			case markerITCH:
				if len(nullTermStr(scratch)) > 0 {
					recListChunk.UnitID = nullTermStr(scratch)[4:]
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
	decodedChunk := &UnidenChunk{}

	if ch == nil {
		return decodedChunk, fmt.Errorf("can't decode a nil chunk")
	}
	if ch.ID != CIDUNID {
		return nil, fmt.Errorf("was given something other than a unid riff chunk")
	}
	rawChunk := RawUnidenChunk{}

	if binErr := ch.ReadBE(&rawChunk); binErr != nil {
		return nil, fmt.Errorf("error when parsing unid chunk as binary: %v\n", binErr)
	}

	if unmarshalErr := decodedChunk.Favorite.UnmarshalBinary(rawChunk.Favorite[0:len(rawChunk.Favorite)]); unmarshalErr != nil {
		return nil, fmt.Errorf("error when decoding binary to favorite: %v", unmarshalErr)
	}

	if unmarshalErr := decodedChunk.System.UnmarshalBinary(rawChunk.System[0:len(rawChunk.System)]); unmarshalErr != nil {
		return nil, fmt.Errorf("error when decoding binary to system: %v", unmarshalErr)
	}

	if unmarshalErr := decodedChunk.Department.UnmarshalBinary(rawChunk.Department[0:len(rawChunk.Department)]); unmarshalErr != nil {
		return nil, fmt.Errorf("error when decoding binary to department: %v", unmarshalErr)
	}

	if unmarshalErr := decodedChunk.Channel.UnmarshalBinary(rawChunk.Channel[0:len(rawChunk.Channel)]); unmarshalErr != nil {
		return nil, fmt.Errorf("error when decoding binary to channel: %v", unmarshalErr)
	}

	if decodedChunk.System.Type != "Conventional" {
		if unmarshalErr := decodedChunk.Site.UnmarshalBinary(rawChunk.Site[0:len(rawChunk.Site)]); unmarshalErr != nil {
			return nil, fmt.Errorf("error when decoding binary to site: %v", unmarshalErr)
		}
	}

	if unmarshalErr := decodedChunk.Metadata.UnmarshalBinary(rawChunk.Metadata[0:len(rawChunk.Metadata)]); unmarshalErr != nil {
		return nil, fmt.Errorf("error when decoding binary to metadata: %v", unmarshalErr)
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
