package main

import (
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"net"
	"net/http"

	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/golang/glog"
)

type SDSKeyType string
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
	AstModeLCNMonitor                  = "LCN_MONITOR"
	AstActivityLog                     = "ACTIVITY_LOG"
	AstLCNFinder                       = "LCN_FINDER"

	AprModePause APRModeType = "PAUSE"
	AprModeRESME APRModeType = "RESUME"
)

var (
	validKeys = loadValidKeys()
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

type MsgPacket struct {
	msg []byte
	ts  time.Time
}

type Locker struct {
	sync.Mutex
	state   bool
	name    string
	pktSent uint64
	pktRecv uint64
}

type APRModeType string
type ASTModeType string

type Modal struct {
	PSI               bool // PSI Mode on/Off
	WSClientConnected bool
	ASTMode           ASTModeType
	APRMode           APRModeType
}

type ScannerCtrl struct {
	locker           Locker
	rq               chan bool
	wq               chan bool
	drained          chan bool
	radioMsg         chan MsgPacket
	hostMsg          chan MsgPacket
	conn             *net.UDPConn
	s                *http.Server
	c                chan os.Signal
	GoProcDelay      time.Duration
	GoProcMultiplier time.Duration
	mode             Modal
}

func (s *ScannerCtrl) IsLocked() bool {
	var locked bool
	s.locker.Lock()
	if s.locker.state == true {
		locked = true
		glog.V(2).Infof("UDP Packets Sent: [%d] UDP Packets Recv: [%d]\n", s.locker.pktSent, s.locker.pktRecv)
	} else {
		locked = false
	}
	s.locker.Unlock()
	return locked
}

func (s *ScannerCtrl) ReceiveFromRadioMsgChannel() (MsgPacket, bool) {
	select {
	case msgToHost := <-s.radioMsg:
		elapsed := time.Since(msgToHost.ts)
		glog.V(2).Infof("Received Message: [%s] To Send To Host: %s\n", elapsed, msgToHost.msg)
		return msgToHost, true
	default:
		time.Sleep(time.Millisecond * 50)
	}
	return MsgPacket{}, false
}

func (s *ScannerCtrl) SendToRadioMsgChannel(msg []byte) bool {

	if !s.IsLocked() {
		glog.V(3).Infof("RadioMsgChannel: No Listener to Receive Msg, Msg Not Sent\n")
		return false
	}

	select {
	case s.radioMsg <- MsgPacket{msg, time.Now()}:
		glog.V(2).Infof("Send Msg[ql=%d]: [%s] to Radio Msg Channel\n", len(s.radioMsg), msg)
		return true
	default:
		glog.V(2).Infof("Queue Full, No Message Sent: %d\n", len(s.radioMsg))
		time.Sleep(time.Millisecond * 50)
	}
	return false
}

func (s *ScannerCtrl) SendToHostMsgChannel(msg []byte) bool {

	if !s.IsLocked() {
		glog.V(3).Infof("HostMsgChannel: No Listener to Receive Msg, Msg Not Sent\n")
		return false
	}

	select {
	case s.hostMsg <- MsgPacket{msg, time.Now()}:
		return true
	default:
		glog.V(2).Infof("Queue Full, No Message Sent: %d\n", len(s.radioMsg))
		time.Sleep(time.Millisecond * 50)
	}
	return false
}

func CreateScannerCtrl() *ScannerCtrl {

	ctrl := &ScannerCtrl{}

	ctrl.rq = make(chan bool, 1)
	ctrl.wq = make(chan bool, 1)
	ctrl.drained = make(chan bool, 1)

	ctrl.radioMsg = make(chan MsgPacket, 100)
	ctrl.hostMsg = make(chan MsgPacket, 100)
	ctrl.c = make(chan os.Signal, 1)
	ctrl.GoProcDelay = DefaultGoProcDelay
	ctrl.GoProcMultiplier = DefaultGoProcMultiplier
	return ctrl
}

func decodeXMLUdpPacket(xmlPacket []byte, data interface{}) error {

	err := xml.Unmarshal(xmlPacket, data)
	if err != nil {
		return err
	}
	return nil
}

func startWSServer(host string, port int, ctrl *ScannerCtrl) (*http.Server, error) {

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			// handle error
		}

		ctrl.mode.WSClientConnected = true

		defer func() {
			conn.Close()
			if ctrl.locker.name == r.RemoteAddr {
				ctrl.locker.state = false
				glog.V(1).Infof("Unlocking Scanner for new WS Client Connection\n")
			}
		}()

		if ctrl.locker.state == true {
			glog.V(2).Infof("Scanner is already is use by: %s\n", ctrl.locker.name)
			err = wsutil.WriteServerMessage(conn, ws.OpBinary, []byte("Locked by "+ctrl.locker.name))
			if err != nil {
				// handle error
				glog.Error(fmt.Sprintf("Failed To Write: %s\n", err))
				return
			}
			return
		} else {
			glog.V(1).Infof("Locking Scanner for new WS Client Connection: [%s]\n", r.RemoteAddr)
			ctrl.locker.Lock()
			ctrl.locker.state = true
			ctrl.locker.name = r.RemoteAddr
			ctrl.locker.pktSent = 0
			ctrl.locker.pktRecv = 0
			ctrl.locker.Unlock()
		}

		quitReader := make(chan bool, 1)
		quitWriter := make(chan bool, 1)
		done := make(chan bool, 1)

		// WS Reader routine
		go func() {

			for {
				select {
				case _ = <-quitWriter:
					return
				default:
					time.Sleep(time.Millisecond * 50)
				}
				msgFromHost, _, err := wsutil.ReadClientData(conn)

				if err != nil {
					// handle error
					glog.V(1).Infof("Failed to read from WS, Terminating Client Conection: %s\n", err)
					done <- true
					quitReader <- true
					return
				} else {
					if !validMsgFromWSClient(msgFromHost) {
						glog.V(1).Infof("Message From WS Failed Validation: [%s]\n", crlfStrip(msgFromHost, LF))
					} else {
						if string(msgFromHost) == TERMINATE {
							glog.V(1).Infof("Received Client QUIT Command: Terminating Client Conection\n")
							done <- true
							quitReader <- true
							return
						}

						glog.V(2).Infof("Message From Host: [%s]\n", crlfStrip(msgFromHost, LF))
						ctrl.SendToHostMsgChannel(msgFromHost)
					}
				}
			}
		}()

		// WS Writer routine
		go func() {
			for {

				select {
				case _ = <-quitReader:
					return
				default:
					time.Sleep(time.Millisecond * 50)
				}

				msgToHost, ok := ctrl.ReceiveFromRadioMsgChannel()

				if ok {
					elapsed := time.Since(msgToHost.ts)
					glog.V(2).Infof("Received[ql=%d] Message To Host at [%s] [%s]\n",
						len(ctrl.radioMsg), elapsed, crlfStrip(msgToHost.msg, LF|NL))
					err = wsutil.WriteServerMessage(conn, ws.OpBinary, msgToHost.msg)
					if err != nil {
						// handle error
						glog.Error(fmt.Sprintf("Failed To Write: %s\n", err))
						done <- true
						quitWriter <- true
						return
					} else {
						glog.V(2).Infof("Message To Host: [%s]\n", string(msgToHost.msg))
					}
				}
			}
		}()

		<-done
		// TODO - need to reset radio into normal state:  Turn off PSI
		ctrl.SendToHostMsgChannel([]byte("PSI,0\r"))

		// drain any UDP Traffic
		ctrl.drainUDP()

		// Drain WS Messages so new connection won't receive old messages
		for {
			_, ok := ctrl.ReceiveFromRadioMsgChannel()
			if !ok {
				break
			} else {
				glog.V(3).Infof("RadioMsgChannel Draining...\n")
				glog.Flush()
			}
		}
		ctrl.mode.WSClientConnected = false

	})

	s := &http.Server{
		Addr:           host + ":" + strconv.Itoa(port),
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go s.ListenAndServe()
	glog.V(1).Infof("Started WebSocket Server at: %s\n", s.Addr)
	return s, nil
}

func crlfStrip(msg []byte, flags uint) string {

	var replacer []string
	switch {
	case flags&LF == LF:
		replacer = append(replacer, "\r", "\\r")
		fallthrough
	case flags&NL == NL:
		replacer = append(replacer, "\n", "")
	}
	r := strings.NewReplacer(replacer...)
	return r.Replace(string(msg))
}

func main() {

	portNum := flag.Int("port", 50536, "udp port for SDS200")
	wsPortNum := flag.Int("wsport", 8080, "ws port for server")
	var hostName string
	flag.StringVar(&hostName, "host", "192.168.1.26", "hostname of SDS200")

	flag.Parse()

	service := hostName + ":" + strconv.Itoa(*portNum)

	RemoteAddr, err := net.ResolveUDPAddr("udp", service)

	ctrl := CreateScannerCtrl()
	ctrl.conn, err = net.DialUDP("udp", nil, RemoteAddr)

	// note : you can use net.ResolveUDPAddr for LocalAddr as well
	//        for this tutorial simplicity sake, we will just use nil

	if err != nil {
		glog.Error(fmt.Sprintf("Failed to DailUDP: %s\n", err))
	}

	glog.V(1).Infof("Established connection to %s \n", service)
	glog.V(1).Infof("Remote UDP address : %s \n", ctrl.conn.RemoteAddr().String())
	glog.V(1).Infof("Local UDP client address : %s \n", ctrl.conn.LocalAddr().String())

	defer ctrl.conn.Close()

	// write a message to Scanner
	go func(ctrl *ScannerCtrl) {
		for {
			var err error
			select {
			case <-ctrl.wq:
				glog.V(1).Infof("Shutting down writer...\n")
				return
			case msgToRadio := <-ctrl.hostMsg:
				elapsed := time.Since(msgToRadio.ts)
				glog.V(2).Infof("Received Message:[ql=%d] From Host:[%s]: [%s]\n", len(ctrl.hostMsg), elapsed,
					crlfStrip(msgToRadio.msg, LF|NL))
				_, err = ctrl.conn.Write(msgToRadio.msg)

				if err != nil {
					glog.V(1).Infof("Error Writing to UDP Socket: %s\n", err)
				} else {
					glog.V(2).Infof("Sent Message From Host: %s\n", msgToRadio.msg)
					ctrl.locker.pktSent++
				}

			default:
				time.Sleep(time.Millisecond * ctrl.GoProcDelay * ctrl.GoProcMultiplier)
			}
		}
	}(ctrl)

	// receive message from server
	go func(ctrl *ScannerCtrl) {

		var do_quit bool = false

		for {

			buffer := make([]byte, 65535)
			ctrl.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, _, err := ctrl.conn.ReadFromUDP(buffer)

			buffer = []byte(crlfStrip(buffer, LF|NL))

			if err != nil {
				if e, ok := err.(net.Error); !ok || !e.Timeout() {
					glog.Error(fmt.Sprintf("Error on ReadFromUDP: %s, %d\n", e, n))
				} else {
					// so we timedout - and if we've received a quit then exit after draining the upd packets
					if do_quit {
						select {
						case ctrl.drained <- true:
							glog.V(2).Infof("Draining UDP Packets\n")
						default:
							time.Sleep(time.Millisecond * 50)
						}
						return
					} else {
						// TODO - no ping! ctrl.SendToRadioMsgChannel([]byte("ping"))
					}
				}

			} else {

				ctrl.locker.pktRecv++
				switch string(buffer[:3]) {
				case "APR":
					glog.V(1).Infof("APR: %s\n", buffer[4:])
				case "AST":
					glog.V(1).Infof("AST: %s\n", buffer[4:])
				case "MDL":
					glog.V(1).Infof("MDL: Model: %s\n", buffer[4:])
					ctrl.SendToRadioMsgChannel([]byte("MDL," + string(buffer[4:])))
				case "VER":
					glog.V(1).Infof("VER: Firmare: %s\n", buffer[4:])
					ctrl.SendToRadioMsgChannel([]byte("VER," + string(buffer[4:])))
				case "MSB":
					glog.V(1).Infof("MSB: Params: %s\n", buffer[4:])
					ctrl.SendToRadioMsgChannel([]byte("MSB," + string(buffer[4:])))
				case "MSV":
					glog.V(1).Infof("MSV: Param: %s\n", buffer[4:])
					ctrl.SendToRadioMsgChannel([]byte("MSV," + string(buffer[4:])))
				case "MNU":
					glog.V(1).Infof("MNU: Params: %s\n", buffer[4:])
					ctrl.SendToRadioMsgChannel([]byte("MNU," + string(buffer[4:])))
				case "MSI":
					msiInfo := MsiInfo{}
					glog.V(1).Infof("MSI: %s\n", buffer[4:])
					if decodeXMLUdpPacket(buffer[11:], &msiInfo) != nil {
						glog.Error(fmt.Sprintf("Failed to decode XML: %s\n", err))
					} else {
						glog.V(1).Infof("MSI: Name: %s, Index: %s, MenuType: %s Value: %s Selected %s \n",
							msiInfo.Name, msiInfo.Index, msiInfo.MenuType, msiInfo.Value, msiInfo.Selected)
						for mi := 0; mi < len(msiInfo.MenuItem); mi++ {
							glog.V(1).Infof("\tMENUItem[%d]: Name: %s, Index: %s, Text: %s\n",
								mi, msiInfo.MenuItem[mi].Name, msiInfo.MenuItem[mi].Index, msiInfo.MenuItem[mi].Text)
						}
					}
					ctrl.SendToRadioMsgChannel([]byte("MSI," + string(buffer[4:])))
				case "DTM":
					glog.V(1).Infof("DTM: %s\n", buffer[4:])
					ctrl.SendToRadioMsgChannel([]byte("DTM," + string(buffer[4:])))
				case "LCR":
					glog.V(1).Infof("LCR: %s\n", buffer[4:])
					ctrl.SendToRadioMsgChannel([]byte("LCR," + string(buffer[4:])))
				case "URC":
					glog.V(1).Infof("URC: %s\n", buffer[4:])
					ctrl.SendToRadioMsgChannel([]byte("URC," + string(buffer[4:])))
				case "STS":
				case "GLG":
				case "GLT":
					switch getXmlGLTFormatType(buffer[11:]) {
					case GltXmlFL:
						gltFl := GltFLInfo{}
						if decodeXMLUdpPacket(buffer[11:], &gltFl) != nil {
							glog.Error(fmt.Sprintf("Failed to decode XML: %s\n", err))
						} else {
							for fl := 0; fl < len(gltFl.FL); fl++ {
								glog.V(1).Infof("GLT,FL[%d]: Name: %s, Index: %s, Monitor: %s\n",
									fl+1, gltFl.FL[fl].Name, gltFl.FL[fl].Index, gltFl.FL[fl].Monitor)
							}
							ctrl.SendToRadioMsgChannel([]byte("GLT,FL" + string(buffer[11:])))
						}
					case GltXmlSYS:
						gltSys := GltSysInfo{}
						if decodeXMLUdpPacket(buffer[11:], &gltSys) != nil {
							glog.Error(fmt.Sprintf("Failed to decode XML: %s\n", err))
						} else {
							for sys := 0; sys < len(gltSys.SYS); sys++ {
								glog.V(1).Infof("GLT,SYS[%d]: Name: %s, Index: %s, TrunkID: %s, Type: %s\n",
									sys+1, gltSys.SYS[sys].Name, gltSys.SYS[sys].Index, gltSys.SYS[sys].TrunkId, gltSys.SYS[sys].Type)
							}
							ctrl.SendToRadioMsgChannel([]byte("GLT,SYS" + string(buffer[11:])))
						}

					case GltXmlDEPT:
						gltDept := GltDeptInfo{}
						if decodeXMLUdpPacket(buffer[11:], &gltDept) != nil {
							glog.Error(fmt.Sprintf("Failed to decode XML: %s\n", err))
						} else {
							for dpt := 0; dpt < len(gltDept.DEPT); dpt++ {
								glog.V(1).Infof("GLT,DEPT[%d]: Name: %s, Index: %s, TGroupID: %s\n",
									dpt+1, gltDept.DEPT[dpt].Name, gltDept.DEPT[dpt].Index, gltDept.DEPT[dpt].TGroupId)
							}
							ctrl.SendToRadioMsgChannel([]byte("GLT,DEPT" + string(buffer[11:])))
						}
					case GltXmlSITE:
						gltSite := GltSiteInfo{}
						if decodeXMLUdpPacket(buffer[11:], &gltSite) != nil {
							glog.Error(fmt.Sprintf("Failed to decode XML: %s\n", err))
						} else {
							for site := 0; site < len(gltSite.SITE); site++ {
								glog.V(1).Infof("GLT,SITE[%d]: Name: %s, Index: %s, SiteId: %s\n",
									site+1, gltSite.SITE[site].Name, gltSite.SITE[site].Index, gltSite.SITE[site].SiteId)
							}
							ctrl.SendToRadioMsgChannel([]byte("GLT,SITE" + string(buffer[11:])))
						}
					case GltXmlFTO:
						gltFTO := GltFto{}
						if decodeXMLUdpPacket(buffer[11:], &gltFTO) != nil {
							glog.Error(fmt.Sprintf("Failed to decode XML: %s\n", err))
						} else {
							for fto := 0; fto < len(gltFTO.FTO); fto++ {
								glog.V(1).Infof("GLT,FTO[%d]: Name: %s, Index: %s, Freq: %s, Mod: %s, ToneA: %s, ToneB: %s\n",
									fto+1, gltFTO.FTO[fto].Name, gltFTO.FTO[fto].Index, gltFTO.FTO[fto].Freq, gltFTO.FTO[fto].Mod, gltFTO.FTO[fto].ToneA, gltFTO.FTO[fto].ToneB)
							}
							ctrl.SendToRadioMsgChannel([]byte("GLT,FTO" + string(buffer[11:])))
						}
					case GltXmlCSBANK:
						gltCSBank := GltCSBank{}
						if decodeXMLUdpPacket(buffer[11:], &gltCSBank) != nil {
							glog.Error(fmt.Sprintf("Failed to decode XML: %s\n", err))
						} else {
							for csb := 0; csb < len(gltCSBank.CSBANK); csb++ {
								glog.V(1).Infof("GLT,CSBANK[%d]: Name: %s, Index: %s, Lower: %s, Upper: %s, Mod: %s, Step: %s\n",
									csb+1, gltCSBank.CSBANK[csb].Name, gltCSBank.CSBANK[csb].Index, gltCSBank.CSBANK[csb].Lower, gltCSBank.CSBANK[csb].Upper, gltCSBank.CSBANK[csb].Mod, gltCSBank.CSBANK[csb].Step)
							}
							ctrl.SendToRadioMsgChannel([]byte("GLT,CS_BANK" + string(buffer[11:])))
						}
					case GltXmlTRN_DISCOV:
						gltTrnDisc := GltTrnDiscovery{}
						if decodeXMLUdpPacket(buffer[11:], &gltTrnDisc) != nil {
							glog.Error(fmt.Sprintf("Failed to decode XML: %s\n", err))
						} else {
							for td := 0; td < len(gltTrnDisc.TRNDISCOV); td++ {
								glog.V(1).Infof("GLT,TRN_DISCOV: Name: %s, Delay: %s, Logging: %s, Duration: %s, CompareDB: %s, SystemName: %s SystemType: %s SiteName: %s, TimeOutTimer: %s, AutoStore: %s\n",
									gltTrnDisc.TRNDISCOV[td].Name, gltTrnDisc.TRNDISCOV[td].Delay, gltTrnDisc.TRNDISCOV[td].Logging, gltTrnDisc.TRNDISCOV[td].Duration, gltTrnDisc.TRNDISCOV[td].CompareDB, gltTrnDisc.TRNDISCOV[td].SystemName, gltTrnDisc.TRNDISCOV[td].SystemType, gltTrnDisc.TRNDISCOV[td].SiteName, gltTrnDisc.TRNDISCOV[td].TimeOutTimer, gltTrnDisc.TRNDISCOV[td].AutoStore)
							}
							ctrl.SendToRadioMsgChannel([]byte("GLT,TRN_DISCOV" + string(buffer[11:])))
						}
					case GltXmlCNV_DISCOV:
						gltCnvDisc := GltCnvDiscovery{}
						if decodeXMLUdpPacket(buffer[11:], &gltCnvDisc) != nil {
							glog.Error(fmt.Sprintf("Failed to decode XML: %s\n", err))
						} else {
							for cd := 0; cd < len(gltCnvDisc.CNVDISCOV); cd++ {
								glog.V(1).Infof("GLT,CNV_DISCOV: Name: %s, Lower: %s, Upper: %s, Mod: %s, Step: %s, Delay: %s Logging: %s CompareDB: %s, Duration: %s, TimeOutTimer: %s, AutoStore: %s\n", gltCnvDisc.CNVDISCOV[cd].Name, gltCnvDisc.CNVDISCOV[cd].Lower, gltCnvDisc.CNVDISCOV[cd].Upper, gltCnvDisc.CNVDISCOV[cd].Mod, gltCnvDisc.CNVDISCOV[cd].Step, gltCnvDisc.CNVDISCOV[cd].Delay, gltCnvDisc.CNVDISCOV[cd].Logging, gltCnvDisc.CNVDISCOV[cd].CompareDB, gltCnvDisc.CNVDISCOV[cd].Duration, gltCnvDisc.CNVDISCOV[cd].TimeOutTimer, gltCnvDisc.CNVDISCOV[cd].AutoStore)
								ctrl.SendToRadioMsgChannel([]byte("GLT,CNV_DISCOV" + string(buffer[11:])))
							}
						}
					case GltXmlUREC_FOLDER:
						gltUrecFolder := GltUrecFolder{}
						if decodeXMLUdpPacket(buffer[11:], &gltUrecFolder) != nil {
							glog.Error(fmt.Sprintf("Failed to decode XML: %s\n", err))
						} else {
							for fi := 0; fi < len(gltUrecFolder.URECFOLDER); fi++ {
								glog.V(1).Infof("GLT,UREC_FOLDER: Name: %s, Index: %s, Text: %s\n",
									gltUrecFolder.URECFOLDER[fi].Name, gltUrecFolder.URECFOLDER[fi].Index, gltUrecFolder.URECFOLDER[fi].Text)
							}
							ctrl.SendToRadioMsgChannel([]byte("GLT,UREC_FOLDER" + string(buffer[11:])))
						}

					default:
						glog.V(1).Infof("Unhandled GltXml Type: %s\n", buffer)
					}
				case "VOL":
					glog.V(1).Infof("VOL: Volume: %s\n", buffer[4:])
					ctrl.SendToRadioMsgChannel([]byte("VOL," + string(buffer[4:])))
				case "SQL":
					glog.V(1).Infof("SQL: Squelch: %s\n", buffer[4:])
					ctrl.SendToRadioMsgChannel([]byte("SQL," + string(buffer[4:])))
				case "PWR":
					glog.V(1).Infof("PWR: Power: %s\n", buffer[4:])
					ctrl.SendToRadioMsgChannel([]byte("PWR," + string(buffer[4:])))
				case "GSI":
					si := ScannerInfo{}
					if decodeXMLUdpPacket(buffer[11:], &si) != nil {
						glog.Error(fmt.Sprintf("Failed to decode XML: %s\n", err))
					} else {
						glog.V(1).Infof("GSI: System: %s, Department: %s, Site: %s, Freq: [%s] Mon: [%s]\n",
							si.System.Name, si.Department.Name, si.Site.Name, si.SiteFrequency.Freq, si.MonitorList.Name)
						glog.V(1).Infof("\tMode: %s\n", si.Mode)
						ctrl.SendToRadioMsgChannel([]byte("GSI," + string(buffer[11:])))
					}
				case "PSI":
					switch {
					case string(buffer[4:6]) == "OK":
						ctrl.mode.PSI = false
						glog.V(1).Infof("PSI: Stopped\n")
					case string(buffer[4:9]) == "<XML>":
						ctrl.mode.PSI = true
						si := ScannerInfo{}
						if decodeXMLUdpPacket(buffer[11:], &si) != nil {
							glog.Error(fmt.Sprintf("Failed to decode XML: %s\n", err))
						} else {
							glog.V(1).Infof("%s: System: %s, Department: %s, Site: %s, Freq: [%s] Mon: [%s]\n",
								buffer[:3], si.System.Name, si.Department.Name, si.Site.Name, si.SiteFrequency.Freq, si.MonitorList.Name)
							glog.V(1).Infof("\tMode: %s\n", si.Mode)
							ctrl.SendToRadioMsgChannel([]byte("GSI," + string(buffer[11:])))
						}
					default:
						glog.V(1).Infof("PSI: Invalid Mode:: %s\n", string(buffer[4:]))
						continue
					}
				case "KEY":
					glog.V(1).Infof("KEY: %s\n", buffer[4:])
					ctrl.SendToRadioMsgChannel([]byte("KEY," + string(buffer[4:])))

				default:
					glog.V(1).Infof("Unhandle Key: %s\n", buffer[:3])
				}
			}

			select {
			case <-ctrl.rq:
				glog.V(1).Infof("Shutting down reader...\n")
				do_quit = true
			default:
				time.Sleep(time.Millisecond * ctrl.GoProcDelay)
			}
		}
	}(ctrl)

	ctrl.s, err = startWSServer("", *wsPortNum, ctrl)

	signal.Notify(ctrl.c, os.Interrupt)
	<-ctrl.c

	// gracefully terminate go routines
	glog.V(1).Infof("Terminating on signal...\n")
	ctrl.rq <- true
	ctrl.wq <- true

	glog.V(1).Infof("Waiting on Drained UDP...\n")
	<-ctrl.drained
	glog.V(1).Infof("Drained UDP, Closing Connection...\n")
	ctrl.conn.Close()
	const timeout = 5 * time.Second

	glog.V(1).Infof("Shutting down WebSocket server\n")

	ctx, _ := context.WithTimeout(context.Background(), timeout)
	if err := ctrl.s.Shutdown(ctx); err != nil {
		glog.Error(fmt.Sprintf("Failed to Shutdown: %s\n", err))
	}
	glog.V(1).Infof("UDPServer Terminated.\n")
}

func dumpBuf(buf []byte, len int) {

	for i := 0; i < len; i++ {
		if IsPrint(string(buf[i])) {
			glog.V(1).Infof("%c", buf[i])
		} else {
			glog.V(1).Infof("\n")
		}
	}
	glog.V(1).Infof("\n")
	glog.Flush()
}

func isAlphaNum(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

func IsPrint(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

func getXmlGLTFormatType(gltBuffer []byte) GltXmlType {

	switch {
	case strings.Contains(string(gltBuffer), "FL Index"):
		return GltXmlFL
	case strings.Contains(string(gltBuffer), "SYS Index"):
		return GltXmlSYS
	case strings.Contains(string(gltBuffer), "DEPT Index"):
		return GltXmlDEPT
	case strings.Contains(string(gltBuffer), "SITE"):
		return GltXmlSITE
	case strings.Contains(string(gltBuffer), "TRN_DISCOV"):
		return GltXmlTRN_DISCOV
	case strings.Contains(string(gltBuffer), "CNV_DISCOV"):
		return GltXmlCNV_DISCOV
	case strings.Contains(string(gltBuffer), "FTO"):
		return GltXmlFTO
	case strings.Contains(string(gltBuffer), "UREC_FOLDER"):
		return GltXmlUREC_FOLDER
	case strings.Contains(string(gltBuffer), "CS_BANK"):
		return GltXmlCSBANK
		/*
			GltXmlCFREQ
			GltXmlTGID
			GltXmlSFREQ
			GltXmlAFREQ
			GltXmlATGID
			GltXmlUREC
			GltXmlIREC_FILE
			GltXmlUREC_FILE
		*/
	default:
		return GltXmlUnknown
	}
}

func loadValidKeys() map[string]SDSKeyType {
	sdsKeys := make(map[string]SDSKeyType)

	sdsKeys["M"] = KEY_MENU
	sdsKeys["F"] = KEY_F
	sdsKeys["1"] = KEY_1
	sdsKeys["2"] = KEY_2
	sdsKeys["3"] = KEY_3
	sdsKeys["4"] = KEY_4
	sdsKeys["5"] = KEY_5
	sdsKeys["6"] = KEY_6
	sdsKeys["7"] = KEY_7
	sdsKeys["8"] = KEY_8
	sdsKeys["9"] = KEY_9
	sdsKeys["0"] = KEY_0
	sdsKeys["."] = KEY_DOT
	sdsKeys["E"] = KEY_ENTER
	sdsKeys[">"] = KEY_ROT_RIGHT
	sdsKeys["<"] = KEY_ROT_LEFT
	sdsKeys["^"] = KEY_ROT_PUSH
	sdsKeys["V"] = KEY_VOL_PUSH
	sdsKeys["Q"] = KEY_SQL_PUSH
	sdsKeys["Y"] = KEY_REPLAY
	sdsKeys["A"] = KEY_SYSTEM
	sdsKeys["B"] = KEY_DEPT
	sdsKeys["C"] = KEY_CHANNEL
	sdsKeys["Z"] = KEY_ZIP
	sdsKeys["T"] = KEY_SERV
	sdsKeys["R"] = KEY_RANGE
	return sdsKeys
}

func isValidKey(buffer []byte) bool {
	if len(buffer) != VALID_KEY_CMD_LENGTH {
		return false
	}

	_, ok := validKeys[string(buffer[4:5])]
	if !ok && !(string(buffer[6:10]) == KEY_PUSH_CMD || string(buffer[6:10]) == KEY_HOLD_CMD) {
		return false
	}
	return true
}

// TODO -- add msg validation
func validMsgFromWSClient(msgFromHost []byte) bool {
	return true
}

func (c *ScannerCtrl) drainUDP() {
	buffer := make([]byte, 65535)
	for {
		c.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		_, _, err := c.conn.ReadFromUDP(buffer)

		if err != nil {
			if e, ok := err.(net.Error); !ok || !e.Timeout() {
			} else {
				glog.V(3).Infof("Drained...\n")
				return
			}
		} else {
			glog.V(3).Infof("Packet Draining on WS Close...\n")
			glog.Flush()
		}
	}
}
