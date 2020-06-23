package wavparse

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Recording struct {
	File     string
	Public   *ListChunk
	Private  *UnidenChunk
	Duration time.Duration
}

type ListChunk struct {
	System           string     // IART
	Department       string     // IGNR
	Channel          string     // INAM
	TGIDFreq         string     // ICMT
	Product          string     // IPRD
	Unknown          string     // IKEY
	Timestamp        *time.Time // ICRD
	Tone             string     // ISRC
	UnitID           string     // ITCH
	FavoriteListName string     // ISBJ
	Reserved         string     // ICOP
}

type FavoriteInfo struct {
	Name            string
	File            string
	LocationControl bool
	Monitor         bool
	QuickKey        string
	NumberTag       string
	ConfigKey0      string
	ConfigKey1      string
	ConfigKey2      string
	ConfigKey3      string
	ConfigKey4      string
	ConfigKey5      string
	ConfigKey6      string
	ConfigKey7      string
	ConfigKey8      string
	ConfigKey9      string
}

func (f *FavoriteInfo) UnmarshalBinary(data []byte) error {
	split := strings.Split(string(data), "\x00")

	if len(split) >= 1 {
		f.Name = split[0]
	}
	if len(split) >= 2 {
		f.File = split[1]
	}
	if len(split) >= 3 {
		toggleBool, toggleBoolErr := parseBool(strings.ToLower(split[2]))
		if toggleBoolErr != nil {
			return toggleBoolErr
		}
		f.LocationControl = toggleBool
	}
	if len(split) >= 4 {
		toggleBool, toggleBoolErr := parseBool(strings.ToLower(split[3]))
		if toggleBoolErr != nil {
			return toggleBoolErr
		}
		f.Monitor = toggleBool
	}
	if len(split) >= 5 {
		f.QuickKey = split[4]
	}
	if len(split) >= 6 {
		f.NumberTag = split[5]
	}
	if len(split) >= 7 {
		f.ConfigKey0 = split[6]
	}
	if len(split) >= 8 {
		f.ConfigKey1 = split[7]
	}
	if len(split) >= 9 {
		f.ConfigKey2 = split[8]
	}
	if len(split) >= 10 {
		f.ConfigKey3 = split[9]
	}
	if len(split) >= 11 {
		f.ConfigKey4 = split[10]
	}
	if len(split) >= 12 {
		f.ConfigKey5 = split[11]
	}
	if len(split) >= 13 {
		f.ConfigKey6 = split[12]
	}
	if len(split) >= 14 {
		f.ConfigKey7 = split[13]
	}
	if len(split) >= 15 {
		f.ConfigKey8 = split[14]
	}
	if len(split) >= 16 {
		f.ConfigKey9 = split[15]
	}

	return nil
}

type SiteInfo struct {
	Name             string
	Avoid            bool
	Latitude         float64
	Longitude        float64
	Range            float64
	Modulation       string
	MotorolaBandPlan string
	EDACS            string
	Shape            string
	Attenuator       bool
}

func (s *SiteInfo) UnmarshalBinary(data []byte) error {
	split := strings.Split(string(data), "\x00")

	if len(split) >= 1 {
		s.Name = split[0]
	}
	if len(split) >= 2 {
		s.Avoid, _ = parseBool(strings.ToLower(split[1]))
	}
	if len(split) >= 3 {
		s.Latitude, _ = strconv.ParseFloat(split[2], 64)
	}
	if len(split) >= 4 {
		s.Longitude, _ = strconv.ParseFloat(split[3], 64)
	}
	if len(split) >= 5 {
		s.Range, _ = strconv.ParseFloat(split[4], 64)
	}
	if len(split) >= 6 {
		s.Modulation = split[5]
	}
	if len(split) >= 7 {
		s.MotorolaBandPlan = split[6]
	}
	if len(split) >= 8 {
		s.EDACS = split[7]
	}
	if len(split) >= 9 {
		s.Shape = split[8]
	}
	if len(split) >= 10 {
		s.Attenuator, _ = parseBool(strings.ToLower(split[9]))
	}
	return nil
}

type SystemInfo struct {
	Name                     string
	Avoid                    bool
	Blank                    string
	Type                     string
	IDSearch                 bool
	EmergencyAlertType       bool
	AlertVolume              string
	MotorolaStatusBit        string
	P25NAC                   string
	QuickKey                 string
	NumberTag                string
	HoldTime                 string
	AnalogAGC                string
	DigitalAGC               string
	EndCode                  string
	PriorityID               string
	EmergencyAlertLightColor string
	EmergencyAlertCondition  string
}

func (s *SystemInfo) UnmarshalBinary(data []byte) error {
	split := strings.Split(string(data), "\x00")

	if len(split) >= 1 {
		s.Name = split[0]
	}
	if len(split) >= 2 {
		s.Avoid, _ = parseBool(strings.ToLower(split[1]))
	}
	if len(split) >= 3 {
		s.Blank = split[2]
	}
	if len(split) >= 4 {
		s.Type = split[3]
	}
	if len(split) >= 5 {
		s.IDSearch, _ = parseBool(strings.ToLower(split[4]))
	}
	if len(split) >= 6 {
		s.EmergencyAlertType, _ = parseBool(strings.ToLower(split[5]))
	}
	if len(split) >= 7 {
		s.AlertVolume = split[6]
	}
	if len(split) >= 8 {
		s.MotorolaStatusBit = split[7]
	}
	if len(split) >= 9 {
		s.P25NAC = split[8]
	}
	if len(split) >= 10 {
		s.QuickKey = split[9]
	}
	if len(split) >= 11 {
		s.NumberTag = split[10]
	}
	if len(split) >= 12 {
		s.HoldTime = split[11]
	}
	if len(split) >= 13 {
		s.AnalogAGC = split[12]
	}
	if len(split) >= 14 {
		s.DigitalAGC = split[13]
	}
	if len(split) >= 15 {
		s.EndCode = split[14]
	}
	if len(split) >= 16 {
		s.PriorityID = split[15]
	}
	if len(split) >= 17 {
		s.EmergencyAlertLightColor = split[16]
	}
	if len(split) >= 18 {
		s.EmergencyAlertCondition = split[17]
	}

	return nil
}

type DepartmentInfo struct {
	Name      string
	Avoid     bool
	Latitude  float64
	Longitude float64
	Range     float64
	Shape     string
	NumberTag string
}

func (d *DepartmentInfo) UnmarshalBinary(data []byte) error {
	split := strings.Split(string(data), "\x00")

	if len(split) >= 1 {
		d.Name = split[0]
	}
	if len(split) >= 2 {
		d.Avoid, _ = parseBool(strings.ToLower(split[1]))
	}
	if len(split) >= 3 {
		d.Latitude, _ = strconv.ParseFloat(split[2], 64)
	}
	if len(split) >= 4 {
		d.Longitude, _ = strconv.ParseFloat(split[3], 64)
	}
	if len(split) >= 5 {
		d.Range, _ = strconv.ParseFloat(split[4], 64)
	}
	if len(split) >= 6 {
		d.Shape = split[5]
	}
	if len(split) >= 7 {
		d.NumberTag = split[6]
	}

	return nil
}

type ServiceType int

func (s ServiceType) String() string {
	switch int(s) {
	case 1:
		return "Multi Dispatch"
	case 2:
		return "Law Dispatch"
	case 3:
		return "Fire Dispatch"
	case 4:
		return "EMS Dispatch"
	case 5:
		return "Reserved"
	case 6:
		return "Multi Tac"
	case 7:
		return "Law Tac"
	case 8:
		return "Fire Tac"
	case 9:
		return "EMS Tac"
	case 10:
		return "Reserved"
	case 11:
		return "Interop"
	case 12:
		return "Hospital"
	case 13:
		return "Ham"
	case 14:
		return "Public Works"
	case 15:
		return "Aircraft"
	case 16:
		return "Federal"
	case 17:
		return "Business"
	case 18:
		return "Reserved"
	case 19:
		return "Reserved"
	case 20:
		return "Railroad"
	case 21:
		return "Other"
	case 22:
		return "Multi Talk"
	case 23:
		return "Law Talk"
	case 24:
		return "Fire Talk"
	case 25:
		return "EMS Talk"
	case 26:
		return "Transportation"
	case 27:
		return "Reserved"
	case 28:
		return "Reserved"
	case 29:
		return "Emergency Ops"
	case 30:
		return "Military"
	case 31:
		return "Media"
	case 32:
		return "Schools"
	case 33:
		return "Security"
	case 34:
		return "Utilities"
	case 35:
		return "Reserved"
	case 36:
		return "Reserved"
	case 37:
		return "Corrections"
	case 208:
		return "Custom 1"
	case 209:
		return "Custom 2"
	case 210:
		return "Custom 3"
	case 211:
		return "Custom 4"
	case 212:
		return "Custom 5"
	case 213:
		return "Custom 6"
	case 214:
		return "Custom 7"
	case 215:
		return "Custom 8"
	case 216:
		return "Racing Officials"
	case 217:
		return "Racing Teams"
	case 255:
		return "Unspecified"
	default: // or 0
		return "Unknown"
	}
}

type ChannelInfo struct {
	Name            string
	Avoid           bool
	TGIDFrequency   string
	Mode            string
	ToneCode        string
	ServiceType     ServiceType
	Attenuator      int // Conventional systems only
	DelayValue      int
	VolumeOffset    int
	AlertToneType   string
	AlertToneVolume string
	AlertLightColor string
	AlertLightType  string
	NumberTag       string
	Priority        bool
}

func (c *ChannelInfo) UnmarshalBinary(data []byte) error {
	split := strings.Split(string(data), "\x00")

	if len(split) >= 1 {
		c.Name = split[0]
	}
	if len(split) >= 2 {
		c.Avoid, _ = parseBool(strings.ToLower(split[1]))
	}
	if len(split) >= 3 {
		c.TGIDFrequency = split[2]
	}
	if len(split) >= 4 {
		c.Mode = split[3]
	}
	if len(split) >= 5 {
		c.ToneCode = split[4]
	}
	if len(split) >= 6 {
		parsed, _ := strconv.ParseInt(split[5], 10, 32)
		c.ServiceType = ServiceType(parsed)
	}

	conventionalOffset := 0

	if len(split) > 15 { // Conventional systems have one extra channel field, Attenuator
		conventionalOffset = 1
		if len(split) >= 7 {
			parsed, _ := strconv.ParseInt(split[6], 10, 32)
			c.Attenuator = int(parsed)
		}
	}

	if len(split) >= 7 {
		parsed, _ := strconv.ParseInt(split[conventionalOffset+6], 10, 32)
		c.DelayValue = int(parsed)
	}
	if len(split) >= 8 {
		parsed, _ := strconv.ParseInt(split[conventionalOffset+7], 10, 32)
		c.VolumeOffset = int(parsed)
	}
	if len(split) >= 9 {
		c.AlertToneType = split[conventionalOffset+8]
	}
	if len(split) >= 10 {
		c.AlertToneVolume = split[conventionalOffset+9]
	}
	if len(split) >= 11 {
		c.AlertLightColor = split[conventionalOffset+10]
	}
	if len(split) >= 12 {
		c.AlertLightType = split[conventionalOffset+11]
	}
	if len(split) >= 13 {
		c.NumberTag = split[conventionalOffset+12]
	}
	if len(split) >= 14 {
		c.Priority, _ = parseBool(strings.ToLower(split[conventionalOffset+13]))
	}

	return nil
}

type Metadata struct {
	TGID      string
	Frequency float64
	WACN      string
	NAC       string
	UnitID    string

	RawTGID      string
	RawFrequency string
	RawWACN      string
	RawNAC       string
	RawUnitID    string

	FrequencyFmt string
	WACNFmt      string
	UnknownFmt   string
	NACFmt       string
}

func (t *Metadata) UnmarshalBinary(data []byte) error {
	fmtChunkSplit := strings.Split(string(data[0:65]), "\x00")

	if len(fmtChunkSplit) >= 1 {
		t.RawTGID = fmtChunkSplit[0]
		if len(t.RawTGID) >= 5 {
			t.TGID = t.RawTGID[5:]
		}
	}

	uidStr := string(data[99:110])

	if uidStr[0:4] == "UID:" {
		t.RawUnitID = strings.Split(uidStr, "\x00")[0]
		t.UnitID = t.RawUnitID[4:]
	}

	if len(fmtChunkSplit) >= 3 {
		t.FrequencyFmt = fmtChunkSplit[2]

		t.RawFrequency = fmt.Sprintf(t.FrequencyFmt, data[68:70], data[70:72])

		t.RawFrequency = strings.TrimLeft(t.RawFrequency, "0")

		t.Frequency, _ = strconv.ParseFloat(strings.Split(t.RawFrequency, " ")[0], 64)
	}

	if len(fmtChunkSplit) >= 4 {
		t.WACNFmt = fmtChunkSplit[3]

		t.RawWACN = fmt.Sprintf(t.WACNFmt, data[212:216])

		t.WACN = t.RawWACN[5:]
	}

	if len(fmtChunkSplit) >= 6 {
		t.UnknownFmt = fmtChunkSplit[5]
	}

	if len(fmtChunkSplit) >= 7 {
		t.NACFmt = fmtChunkSplit[6]

		t.RawNAC = fmt.Sprintf(t.NACFmt, data[174:176])
		t.NAC = t.RawNAC[1 : len(t.RawNAC)-2]
	}

	return nil
}

type RawUnidenChunk struct {
	// Start byte 600
	Favorite   [65]byte   // 0-65 	   / 600-665
	System     [65]byte   // 65-130 	 / 665-730
	Department [65]byte   // 130-195 	 / 730-795
	Channel    [65]byte   // 195-260 	 / 795-860
	Site       [65]byte   // 260-325 	 / 860-925
	Empty      [283]byte  // 325-608 	 / 925-1208
	Metadata   [216]byte  // 608-824 	 / 1208-1424
	Remainder  [1224]byte // 824-2048  / 1424-2648
	// Total size is 2048
}

type UnidenChunk struct {
	Favorite   FavoriteInfo
	System     SystemInfo
	Department DepartmentInfo
	Channel    ChannelInfo
	Site       SiteInfo
	Metadata   Metadata
}
