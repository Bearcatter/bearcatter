package server

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Bearcatter/bearcatter/wavparse"
)

type SDSKeyType string
type SDSKeyModeType string
type GltXmlType int

const (
	DefaultGoProcMultiplier = 5
	DefaultGoProcDelay      = 30 // milliseconds

	LF = 1 << 0
	NL = 1 << 1

	VALID_KEY_CMD_LENGTH = 10
	KEY_PUSH_CMD         = "PUSH"
	KEY_HOLD_CMD         = "HOLD"

	KEY_MENU      SDSKeyType = "M"
	KEY_F         SDSKeyType = "F"
	KEY_1         SDSKeyType = "1"
	KEY_2         SDSKeyType = "2"
	KEY_3         SDSKeyType = "3"
	KEY_4         SDSKeyType = "4"
	KEY_5         SDSKeyType = "5"
	KEY_6         SDSKeyType = "6"
	KEY_7         SDSKeyType = "7"
	KEY_8         SDSKeyType = "8"
	KEY_9         SDSKeyType = "9"
	KEY_0         SDSKeyType = "0"
	KEY_DOT       SDSKeyType = "."
	KEY_ENTER     SDSKeyType = "E"
	KEY_ROT_RIGHT SDSKeyType = ">"
	KEY_ROT_LEFT  SDSKeyType = "<"
	KEY_ROT_PUSH  SDSKeyType = "^"
	KEY_VOL_PUSH  SDSKeyType = "V"
	KEY_SQL_PUSH  SDSKeyType = "Q"
	KEY_REPLAY    SDSKeyType = "Y"
	KEY_SYSTEM    SDSKeyType = "A"
	KEY_DEPT      SDSKeyType = "B"
	KEY_CHANNEL   SDSKeyType = "C"
	KEY_ZIP       SDSKeyType = "Z"
	KEY_SERV      SDSKeyType = "T"
	KEY_RANGE     SDSKeyType = "R"

	KEY_MODE_PRESS   SDSKeyModeType = "P" // Press (One Push)
	KEY_MODE_LONG    SDSKeyModeType = "L" // Long Press (Press and Hold a few seconds)
	KEY_MODE_HOLD    SDSKeyModeType = "H" // Hold (Press and Hold until Release receive)
	KEY_MODE_RELEASE SDSKeyModeType = "R" // Release (Cancel Hold state

	GltXmlUnknown GltXmlType = -1
	GltXmlFL      GltXmlType = iota
	GltXmlSYS
	GltXmlDEPT
	GltXmlSITE
	GltXmlCFREQ
	GltXmlTGID
	GltXmlSFREQ
	GltXmlAFREQ
	GltXmlATGID
	GltXmlFTO
	GltXmlCSBANK
	GltXmlUREC
	GltXmlIREC_FILE
	GltXmlUREC_FOLDER
	GltXmlUREC_FILE
	GltXmlTRN_DISCOV
	GltXmlCNV_DISCOV

	AstModeCurrentActivity ASTModeType = "CURRENT_ACTIVITY"
	AstModeLCNMonitor      ASTModeType = "LCN_MONITOR"
	AstActivityLog         ASTModeType = "ACTIVITY_LOG"
	AstLCNFinder           ASTModeType = "LCN_FINDER"

	AprModePause APRModeType = "PAUSE"
	AprModeRESME APRModeType = "RESUME"
)

var (
	// validKeys = loadValidKeys()
	TERMINATE = "quit\r"
)

type ScannerInfo struct {
	XMLName     xml.Name `xml:"ScannerInfo"`
	Text        string   `xml:",chardata"`
	Mode        string   `xml:"Mode,attr"`
	VScreen     string   `xml:"V_Screen,attr"`
	MonitorList struct {
		Text      string `xml:",chardata"`
		Name      string `xml:"Name,attr"`
		Index     string `xml:"Index,attr"`
		ListType  string `xml:"ListType,attr"`
		QKey      string `xml:"Q_Key,attr"`
		NTag      string `xml:"N_Tag,attr"`
		DBCounter string `xml:"DB_Counter,attr"`
	} `xml:"MonitorList"`
	System struct {
		Text       string `xml:",chardata"`
		Name       string `xml:"Name,attr"`
		Index      string `xml:"Index,attr"`
		Avoid      string `xml:"Avoid,attr"`
		SystemType string `xml:"SystemType,attr"`
		QKey       string `xml:"Q_Key,attr"`
		NTag       string `xml:"N_Tag,attr"`
		Hold       string `xml:"Hold,attr"`
	} `xml:"System"`
	Department struct {
		Text  string `xml:",chardata"`
		Name  string `xml:"Name,attr"`
		Index string `xml:"Index,attr"`
		Avoid string `xml:"Avoid,attr"`
		QKey  string `xml:"Q_Key,attr"`
		Hold  string `xml:"Hold,attr"`
	} `xml:"Department"`
	TGID struct {
		Text    string `xml:",chardata"`
		Name    string `xml:"Name,attr"`
		Index   string `xml:"Index,attr"`
		Avoid   string `xml:"Avoid,attr"`
		TGID    string `xml:"TGID,attr"`
		SetSlot string `xml:"SetSlot,attr"`
		RecSlot string `xml:"RecSlot,attr"`
		NTag    string `xml:"N_Tag,attr"`
		Hold    string `xml:"Hold,attr"`
		SvcType string `xml:"SvcType,attr"`
		PCh     string `xml:"P_Ch,attr"`
		LVL     string `xml:"LVL,attr"`
	} `xml:"TGID"`
	UnitID struct {
		Text string `xml:",chardata"`
		Name string `xml:"Name,attr"`
		UID  string `xml:"U_Id,attr"`
	} `xml:"UnitID"`
	Site struct {
		Text  string `xml:",chardata"`
		Name  string `xml:"Name,attr"`
		Index string `xml:"Index,attr"`
		Avoid string `xml:"Avoid,attr"`
		QKey  string `xml:"Q_Key,attr"`
		Hold  string `xml:"Hold,attr"`
		Mod   string `xml:"Mod,attr"`
	} `xml:"Site"`
	SiteFrequency struct {
		Text string `xml:",chardata"`
		Freq string `xml:"Freq,attr"`
		IFX  string `xml:"IFX,attr"`
		SAS  string `xml:"SAS,attr"`
		SAD  string `xml:"SAD,attr"`
	} `xml:"SiteFrequency"`
	DualWatch struct {
		Text string `xml:",chardata"`
		PRI  string `xml:"PRI,attr"`
		CC   string `xml:"CC,attr"`
		WX   string `xml:"WX,attr"`
	} `xml:"DualWatch"`
	TrunkingDiscovery struct {
		Text       string `xml:",chardata"`
		SystemName string `xml:"SystemName,attr"`
		SiteName   string `xml:"SiteName,attr"`
		TGID       string `xml:"TGID,attr"`
		TgidName   string `xml:"TgidName,attr"`
		SAD        string `xml:"SAD,attr"`
		RecSlot    string `xml:"RecSlot,attr"`
		PastTime   string `xml:"PastTime,attr"`
		HitCount   string `xml:"HitCount,attr"`
		UID        string `xml:"U_Id,attr"`
	} `xml:"TrunkingDiscovery"`
	Property struct {
		Text      string `xml:",chardata"`
		F         string `xml:"F,attr"`
		VOL       string `xml:"VOL,attr"`
		SQL       string `xml:"SQL,attr"`
		Sig       string `xml:"Sig,attr"`
		Att       string `xml:"Att,attr"`
		Rec       string `xml:"Rec,attr"`
		KeyLock   string `xml:"KeyLock,attr"`
		P25Status string `xml:"P25Status,attr"`
		Mute      string `xml:"Mute,attr"`
		Backlight string `xml:"Backlight,attr"`
		ALed      string `xml:"A_Led,attr"`
		Dir       string `xml:"Dir,attr"`
		Rssi      string `xml:"Rssi,attr"`
	} `xml:"Property"`
	ViewDescription struct {
		Text      string `xml:",chardata"`
		PlainText []struct {
			Chardata string `xml:",chardata"`
			AttrText string `xml:"Text,attr"`
		} `xml:"PlainText"`
		PopupScreen struct {
			Chardata string `xml:",chardata"`
			AttrText string `xml:"Text,attr"`
			Button   struct {
				Chardata string `xml:",chardata"`
				AttrText string `xml:"Text,attr"`
				KeyCode  string `xml:"KeyCode,attr"`
			} `xml:"Button"`
		} `xml:"PopupScreen"`
	} `xml:"ViewDescription"`
}

type GltFLInfo struct {
	XMLName xml.Name `xml:"GLT"`
	Text    string   `xml:",chardata"`
	FL      []struct {
		Text    string `xml:",chardata"`
		Index   string `xml:"Index,attr"`
		Name    string `xml:"Name,attr"`
		Monitor string `xml:"Monitor,attr"`
		QKey    string `xml:"Q_Key,attr"`
		NTag    string `xml:"N_Tag,attr"`
	} `xml:"FL"`
	Footer struct {
		Text string `xml:",chardata"`
		No   string `xml:"No,attr"`
		EOT  string `xml:"EOT,attr"`
	} `xml:"Footer"`
}

type GltSysInfo struct {
	XMLName xml.Name `xml:"GLT"`
	Text    string   `xml:",chardata"`
	SYS     []struct {
		Text    string `xml:",chardata"`
		Index   string `xml:"Index,attr"`
		TrunkId string `xml:"TrunkId,attr"`
		Name    string `xml:"Name,attr"`
		Avoid   string `xml:"Avoid,attr"`
		Type    string `xml:"Type,attr"`
		QKey    string `xml:"Q_Key,attr"`
		NTag    string `xml:"N_Tag,attr"`
	} `xml:"SYS"`
	Footer struct {
		Text string `xml:",chardata"`
		No   string `xml:"No,attr"`
		EOT  string `xml:"EOT,attr"`
	} `xml:"Footer"`
}

type GltDeptInfo struct {
	XMLName xml.Name `xml:"GLT"`
	Text    string   `xml:",chardata"`
	DEPT    []struct {
		Text     string `xml:",chardata"`
		Index    string `xml:"Index,attr"`
		TGroupId string `xml:"TGroupId,attr"`
		Name     string `xml:"Name,attr"`
		Avoid    string `xml:"Avoid,attr"`
		QKey     string `xml:"Q_Key,attr"`
	} `xml:"DEPT"`
	Footer struct {
		Text string `xml:",chardata"`
		No   string `xml:"No,attr"`
		EOT  string `xml:"EOT,attr"`
	} `xml:"Footer"`
}

type GltSiteInfo struct {
	XMLName xml.Name `xml:"GLT"`
	Text    string   `xml:",chardata"`
	SITE    []struct {
		Text   string `xml:",chardata"`
		Index  string `xml:"Index,attr"`
		SiteId string `xml:"SiteId,attr"`
		Name   string `xml:"Name,attr"`
		Avoid  string `xml:"Avoid,attr"`
		QKey   string `xml:"Q_Key,attr"`
	} `xml:"SITE"`
	Footer struct {
		Text string `xml:",chardata"`
		No   string `xml:"No,attr"`
		EOT  string `xml:"EOT,attr"`
	} `xml:"Footer"`
}

// GLT CFREQ

// GLT TGID

// GLT AFREQ

// GLT ATGID

type GltFto struct {
	XMLName xml.Name `xml:"GLT"`
	Text    string   `xml:",chardata"`
	FTO     []struct {
		Text  string `xml:",chardata"`
		Index string `xml:"Index,attr"`
		Freq  string `xml:"Freq,attr"`
		Mod   string `xml:"Mod,attr"`
		Name  string `xml:"Name,attr"`
		ToneA string `xml:"ToneA,attr"`
		ToneB string `xml:"ToneB,attr"`
	} `xml:"FTO"`
	Footer struct {
		Text string `xml:",chardata"`
		No   string `xml:"No,attr"`
		EOT  string `xml:"EOT,attr"`
	} `xml:"Footer"`
}

type GltCSBank struct {
	XMLName xml.Name `xml:"GLT"`
	Text    string   `xml:",chardata"`
	CSBANK  []struct {
		Text  string `xml:",chardata"`
		Index string `xml:"Index,attr"`
		Name  string `xml:"Name,attr"`
		Lower string `xml:"Lower,attr"`
		Upper string `xml:"Upper,attr"`
		Mod   string `xml:"Mod,attr"`
		Step  string `xml:"Step,attr"`
	} `xml:"CS_BANK"`
	Footer struct {
		Text string `xml:",chardata"`
		No   string `xml:"No,attr"`
		EOT  string `xml:"EOT,attr"`
	} `xml:"Footer"`
}

// GLT,UREC

// GLT,IREC_FILE

// GLT,UREC_FILE,[folder_index]

type GltUrecFolder struct {
	XMLName    xml.Name `xml:"GLT"`
	Text       string   `xml:",chardata"`
	URECFOLDER []struct {
		Text  string `xml:",chardata"`
		Index string `xml:"Index,attr"`
		Name  string `xml:"Name,attr"`
	} `xml:"UREC_FOLDER"`
	Footer struct {
		Text string `xml:",chardata"`
		No   string `xml:"No,attr"`
		EOT  string `xml:"EOT,attr"`
	} `xml:"Footer"`
}

type GltTrnDiscovery struct {
	XMLName   xml.Name `xml:"GLT"`
	Text      string   `xml:",chardata"`
	TRNDISCOV []struct {
		Text         string `xml:",chardata"`
		Name         string `xml:"Name,attr"`
		Delay        string `xml:"Delay,attr"`
		Logging      string `xml:"Logging,attr"`
		Duration     string `xml:"Duration,attr"`
		CompareDB    string `xml:"CompareDB,attr"`
		SystemName   string `xml:"SystemName,attr"`
		SystemType   string `xml:"SystemType,attr"`
		SiteName     string `xml:"SiteName,attr"`
		TimeOutTimer string `xml:"TimeOutTimer,attr"`
		AutoStore    string `xml:"AutoStore,attr"`
	} `xml:"TRN_DISCOV"`
	Footer struct {
		Text string `xml:",chardata"`
		No   string `xml:"No,attr"`
		EOT  string `xml:"EOT,attr"`
	} `xml:"Footer"`
}

// GLT,CNV_DISCOV
type GltCnvDiscovery struct {
	XMLName   xml.Name `xml:"GLT"`
	Text      string   `xml:",chardata"`
	CNVDISCOV []struct {
		Text         string `xml:",chardata"`
		Name         string `xml:"Name,attr"`
		Lower        string `xml:"Lower,attr"`
		Upper        string `xml:"Upper,attr"`
		Mod          string `xml:"Mod,attr"`
		Step         string `xml:"Step,attr"`
		Delay        string `xml:"Delay,attr"`
		Logging      string `xml:"Logging,attr"`
		CompareDB    string `xml:"CompareDB,attr"`
		Duration     string `xml:"Duration,attr"`
		TimeOutTimer string `xml:"TimeOutTimer,attr"`
		AutoStore    string `xml:"AutoStore,attr"`
	} `xml:"CNV_DISCOV"`
	Footer struct {
		Text string `xml:",chardata"`
		No   string `xml:"No,attr"`
		EOT  string `xml:"EOT,attr"`
	} `xml:"Footer"`
}

type MsiInfo struct {
	XMLName  xml.Name `xml:"MSI"`
	Text     string   `xml:",chardata"`
	Name     string   `xml:"Name,attr"`
	Index    string   `xml:"Index,attr"`
	MenuType string   `xml:"MenuType,attr"`
	Value    string   `xml:"Value,attr"`
	Selected string   `xml:"Selected,attr"`
	MenuItem []struct {
		Text  string `xml:",chardata"`
		Name  string `xml:"Name,attr"`
		Index string `xml:"Index,attr"`
	} `xml:"MenuItem"`
	Footer struct {
		Text string `xml:",chardata"`
		No   string `xml:"No,attr"`
		EOT  string `xml:"EOT,attr"`
	} `xml:"Footer"`
}

type ScannerStatus struct {
	// Best guesses based on
	// https://github.com/suidroot/pyUniden/blob/e7705be191474ffada8af12bf1f09c0d4a65057d/pyuniden/main.py#L84-L122
	// http://www.servicedocs.com/ARTIKELEN/7200250170003.pdf
	// http://www.netfiles.ru/share/linked/f1/BCD396T_Protocol.pdf
	Line1          string
	Line2          string
	Line3          string
	Line4          string
	Line5          string
	Line6          string
	Line7          string
	Line8          string
	Line9          string
	Line10         string
	Line11         string
	Line12         string
	Line13         string
	Line14         string
	Line15         string
	Line16         string
	Line17         string
	Line18         string
	Line19         string
	Line20         string
	Frequency      float64
	Squelch        bool
	Mute           bool
	WeatherAlerts  bool
	CCLed          bool
	AlertLED       bool
	BacklightLevel int
	SignalLevel    int
}

func (s *ScannerStatus) Command() string {
	return "STS"
}

func NewScannerStatus(raw string) *ScannerStatus {
	lines := strings.Split(raw, ",")

	resp := ScannerStatus{}

	if len(lines) >= 2 {
		resp.Line1 = lines[1]
	}
	if len(lines) >= 4 {
		resp.Line2 = lines[3]
	}
	if len(lines) >= 6 {
		resp.Line3 = lines[5]
	}
	if len(lines) >= 8 {
		resp.Line4 = lines[7]
	}
	if len(lines) >= 10 {
		resp.Line5 = lines[9]
	}
	if len(lines) >= 12 {
		resp.Line6 = lines[11]
	}
	if len(lines) >= 14 {
		resp.Line7 = lines[13]
	}
	if len(lines) >= 16 {
		resp.Line8 = lines[15]
	}
	if len(lines) >= 18 {
		resp.Line9 = lines[17]
	}
	if len(lines) >= 20 {
		resp.Line10 = lines[19]
	}
	if len(lines) >= 22 {
		resp.Line11 = lines[21]
	}
	if len(lines) >= 24 {
		resp.Line12 = lines[23]
	}
	if len(lines) >= 26 {
		resp.Line13 = lines[25]
	}
	if len(lines) >= 28 {
		resp.Line14 = lines[27]
	}
	if len(lines) >= 30 {
		resp.Line15 = lines[29]
	}
	if len(lines) >= 32 {
		resp.Line16 = lines[31]
	}
	if len(lines) >= 34 {
		resp.Line17 = lines[33]
	}
	if len(lines) >= 36 {
		resp.Line18 = lines[35]
	}
	if len(lines) >= 38 {
		resp.Line19 = lines[37]
	}
	if len(lines) >= 40 {
		resp.Line20 = lines[39]
	}

	resp.Squelch = (lines[36] == "0")
	resp.SignalLevel, _ = strconv.Atoi(lines[41])
	resp.BacklightLevel, _ = strconv.Atoi(lines[43])

	return &resp
}

const DateTimeFormat = "2006,1,2,15,4,5"

type DateTimeInfo struct {
	DaylightSavings bool
	Time            *time.Time
	RTCOK           bool
}

func (d *DateTimeInfo) String() string {
	dst := 0
	if d.DaylightSavings {
		dst = 1
	}
	rtc := 0
	if d.RTCOK {
		rtc = 1
	}
	return fmt.Sprintf("%d,%s,%d", dst, d.Time.Format(DateTimeFormat), rtc)
}

func NewDateTimeInfo(raw string) *DateTimeInfo {
	parsedTime, _ := time.ParseInLocation(DateTimeFormat, raw[2:len(raw)-2], time.Local)
	dst, _ := strconv.ParseBool(raw[0:1])
	rtc, _ := strconv.ParseBool(raw[len(raw)-1:])
	return &DateTimeInfo{
		DaylightSavings: dst,
		Time:            &parsedTime,
		RTCOK:           rtc,
	}
}

type LocationInfo struct {
	Latitude  float64
	Longitude float64
	Range     float64
}

func (l *LocationInfo) String() string {
	return fmt.Sprintf("%f,%f,%f", l.Latitude, l.Longitude, l.Range)
}

func NewLocationInfo(raw string) *LocationInfo {
	split := strings.Split(raw, ",")
	lat, _ := strconv.ParseFloat(split[0], 10)
	lon, _ := strconv.ParseFloat(split[1], 10)
	ran, _ := strconv.ParseFloat(split[2], 10)
	return &LocationInfo{
		Latitude:  lat,
		Longitude: lon,
		Range:     ran,
	}
}

type UserRecordStatus struct {
	Recording    bool
	ErrorCode    *int
	ErrorMessage *string
}

func (u *UserRecordStatus) String() string {
	if u.Recording {
		return "1"
	}
	return "0"
}

func NewUserRecordStatus(raw string) *UserRecordStatus {
	split := strings.Split(raw, ",")
	status, _ := strconv.ParseBool(split[0])
	errCode := 0
	errMsg := ""
	ret := &UserRecordStatus{
		Recording: status,
	}
	if len(split) > 1 {
		errCode, _ = strconv.Atoi(split[1])
		ret.ErrorCode = &errCode
		switch errCode {
		case 1:
			errMsg = "FILE ACCESS ERROR"
		case 2:
			errMsg = "LOW BATTERY"
		case 3:
			errMsg = "SESSION OVER LIMIT"
		case 4:
			errMsg = "RTC LOST"
		default:
			errMsg = "Unknown"
		}
		ret.ErrorMessage = &errMsg
	}
	return ret
}

type MenuMode struct {
	ID    string
	Index string
}

func (m *MenuMode) String() string {
	bits := []string{m.ID}
	if m.Index != "" && (m.ID == "SCAN_SYSTEM" || m.ID == "SCAN_DEPARTMENT" || m.ID == "SCAN_SITE" || m.ID == "SCAN_CHANNEL" || m.ID == "SRCH_RANGE" || m.ID == "FTO_CHANNEL") {
		bits = append(bits, m.Index)
	}
	return strings.Join(bits, ",")
}

type MenuBack struct {
	ReturnLevel int
}

func (m *MenuBack) String() string {
	if m.ReturnLevel > 0 {
		return fmt.Sprintf("0,%d", m.ReturnLevel)
	}
	return ""
}

type MenuSetValue struct {
	ItemIndex int
	Value     string
}

func (m *MenuSetValue) String() string {
	if m.Value != "" {
		return fmt.Sprintf("0,%s", strings.ReplaceAll(m.Value, ",", `\t`))
	}
	return fmt.Sprintf("0,%d", m.ItemIndex)
}

type KeyPress struct {
	Key  string
	Mode string
}

func (k *KeyPress) String() string {
	return fmt.Sprintf("%s,%s", k.Key, k.Mode)
}

type AudioFeedFile struct {
	Name           string
	Size           int64
	ExpectedBlocks int64
	Timestamp      *time.Time
	Data           []byte
	Finished       bool
	Metadata       *wavparse.Recording
}

func (a *AudioFeedFile) ParseMetadata(file string) error {
	var metadataErr error
	a.Metadata, metadataErr = wavparse.DecodeRecording(file)
	return metadataErr
}

var ErrNoFile = fmt.Errorf("no file name was set, probably waiting for info")

func NewAudioFeedFile(pieces []string) (*AudioFeedFile, error) {
	if pieces[0] == "" {
		return nil, ErrNoFile
	}
	file := &AudioFeedFile{
		Name: pieces[0],
	}

	size, sizeErr := strconv.ParseInt(pieces[1], 10, 64)
	if sizeErr != nil {
		return nil, sizeErr
	}

	file.Size = size

	file.ExpectedBlocks = (size / 4096) * 2

	// 06/20/2020 20:31:24
	ts, tsErr := time.ParseInLocation("01/02/2006 15:04:05", pieces[2], time.Local)
	if tsErr != nil {
		return nil, tsErr
	}

	file.Timestamp = &ts

	return file, nil
}
