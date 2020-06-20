package main

import (
	"bytes"
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"net"
	"os/signal"

	"os"
	"time"

	log "github.com/sirupsen/logrus"
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

func main() {

	portNum := flag.Int("port", 50536, "udp port for SDS200")
	wsPortNum := flag.Int("wsport", 8080, "ws port for server")
	var hostName string
	flag.StringVar(&hostName, "host", "", "hostname of SDS200")
	var usbPath string
	flag.StringVar(&usbPath, "usb", "", "path to SDS100 USB port")

	flag.Parse()

	ctrl := CreateScannerCtrl()
	var connCreateErr error

	if hostName != "" {
		ctrl.conn, connCreateErr = NewUDPConn(hostName, *portNum)
	} else if usbPath != "" {
		ctrl.conn, connCreateErr = NewUSBConn(usbPath)
	}

	if connCreateErr != nil {
		log.Fatalln("Failed to create connection", connCreateErr)
	}

	if connOpenErr := ctrl.conn.Open(); connOpenErr != nil {
		log.Fatalln("Failed to open connection", connOpenErr)
	}

	log.Infoln(ctrl.conn.String())
	if ctrl.conn.Type == ConnTypeNetwork {
		log.Infoln("Remote UDP address", ctrl.conn.udpConn.RemoteAddr().String())
		log.Infoln("Local UDP client address", ctrl.conn.udpConn.LocalAddr().String())
	}

	defer ctrl.conn.Close()

	// write a message to Scanner
	go func(ctrl *ScannerCtrl) {
		for {
			select {
			case <-ctrl.wq:
				log.Infoln("Shutting down writer...")
				return
			case msgToRadio := <-ctrl.hostMsg:
				elapsed := time.Since(msgToRadio.ts)
				log.Infof("Received Message:[ql=%d] From Host:[%s]: [%s]", len(ctrl.hostMsg), elapsed,
					crlfStrip(msgToRadio.msg, LF|NL))
				if _, writeErr := ctrl.conn.Write(msgToRadio.msg); writeErr != nil {
					log.Errorln("Error Writing to scanner", writeErr)
				} else {
					log.Infoln("Sent Message From Host", string(msgToRadio.msg))
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

		xmlMessage := make([]byte, 0)
		isXML := false
		var xmlMessageType string

		for {
			// ctrl.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			buffer := make([]byte, 16384)
			n, readErr := ctrl.conn.Read(buffer)
			if readErr != nil {
				log.Errorln("Error on read!", readErr)
				if e, ok := readErr.(net.Error); !ok || !e.Timeout() {
					log.Errorln("Error on read", e, n)
				} else {
					log.Errorln("Unknown error, quitting!")
					// so we timedout - and if we've received a quit then exit after draining the upd packets
					if do_quit {
						select {
						case ctrl.drained <- true:
							log.Infoln("Draining Packets")
						default:
							time.Sleep(time.Millisecond * 50)
						}
						return
					} else {
						// TODO - no ping! ctrl.SendToRadioMsgChannel([]byte("ping"))
					}
				}
			}

			if ctrl.conn.Type == ConnTypeNetwork {
				buffer = []byte(crlfStrip(buffer, LF|NL))
			}

			if bytes.Equal(buffer[4:9], []byte(`<XML>`)) {
				xmlMessageType = string(buffer[0:3])
				copy(xmlMessage, buffer[0:n])
				if IsValidXMLMessage(xmlMessageType, xmlMessage) == false {
					isXML = true
					continue
				}
				buffer = xmlMessage
				isXML = false
				xmlMessageType = ""
				xmlMessage = make([]byte, 0)
			}

			if isXML {
				xmlMessage = append(xmlMessage, buffer[0:n]...)
				if IsValidXMLMessage(xmlMessageType, xmlMessage) == false {
					continue
				}
				// Double comma to match the /r that is normally here
				buffer = []byte(fmt.Sprintf(`%s,<XML>,,%s`, xmlMessageType, string(xmlMessage)))
				isXML = false
				xmlMessageType = ""
				xmlMessage = make([]byte, 0)
			}

			msgType := string(buffer[:3])

			ctrl.locker.pktRecv++
			switch msgType {
			case "APR":
				log.Infoln("APR", string(buffer[4:]))
			case "AST":
				log.Infoln("AST", string(buffer[4:]))
			case "MDL":
				log.Infoln("MDL: Model", string(buffer[4:]))
				ctrl.SendToRadioMsgChannel([]byte("MDL," + string(buffer[4:])))
			case "VER":
				log.Infoln("VER: Firmare", string(buffer[4:]))
				ctrl.SendToRadioMsgChannel([]byte("VER," + string(buffer[4:])))
			case "MSB":
				log.Infoln("MSB: Params", string(buffer[4:]))
				ctrl.SendToRadioMsgChannel([]byte("MSB," + string(buffer[4:])))
			case "MSV":
				log.Infoln("MSV: Param", string(buffer[4:]))
				ctrl.SendToRadioMsgChannel([]byte("MSV," + string(buffer[4:])))
			case "MNU":
				log.Infoln("MNU: Params", string(buffer[4:]))
				ctrl.SendToRadioMsgChannel([]byte("MNU," + string(buffer[4:])))
			case "MSI":
				msiInfo := MsiInfo{}
				log.Infoln("MSI", string(buffer[4:]))
				if decodeErr := xml.Unmarshal(buffer[11:], &msiInfo); decodeErr != nil {
					log.Errorln("Failed to decode XML", decodeErr)
				} else {
					log.Infof("MSI: Name: %s, Index: %s, MenuType: %s Value: %s Selected %s ",
						msiInfo.Name, msiInfo.Index, msiInfo.MenuType, msiInfo.Value, msiInfo.Selected)
					for mi := 0; mi < len(msiInfo.MenuItem); mi++ {
						log.Infof("\tMENUItem[%d]: Name: %s, Index: %s, Text: %s",
							mi, msiInfo.MenuItem[mi].Name, msiInfo.MenuItem[mi].Index, msiInfo.MenuItem[mi].Text)
					}
				}
				ctrl.SendToRadioMsgChannel([]byte("MSI," + string(buffer[4:])))
			case "DTM":
				log.Infoln("DTM:", string(buffer[4:]))
				timeInfo := NewDateTimeInfo(string(buffer[4:n]))
				log.Infof("DTM: DST?: %t, Time: %s, RTC OK? %t\n", timeInfo.DaylightSavings, timeInfo.Time, timeInfo.RTCOK)
				ctrl.SendToRadioMsgChannel([]byte("DTM," + string(buffer[4:])))
			case "LCR":
				log.Infoln("LCR:", string(buffer[4:]))
				locInfo := NewLocationInfo(string(buffer[4:n]))
				log.Infof("LCR: Latitude: %f, Longitude: %f, Range: %f\n", locInfo.Latitude, locInfo.Longitude, locInfo.Range)
				ctrl.SendToRadioMsgChannel([]byte("LCR," + string(buffer[4:])))
			case "URC":
				log.Infoln("URC:", string(buffer[4:]))
				recStatus := NewUserRecordStatus(string(buffer[4:n]))
				log.Infof("URC: Recording? %t, ErrorCode: %d, ErrorMessage: %s\n", recStatus.Recording, recStatus.ErrorCode, recStatus.ErrorMessage)
				ctrl.SendToRadioMsgChannel([]byte("URC," + string(buffer[4:])))
			case "STS":
				log.Infoln("STS", string(buffer[4:]))
				stsInfo := NewScannerStatus(string(buffer[4:]))
				log.Infof("STS: Line 1: %s, Line 2: %s, Line 3: %s, Line 4: %s, SQL: %t, Signal Level: %d\n",
					stsInfo.Line1, stsInfo.Line2, stsInfo.Line3, stsInfo.Line4, stsInfo.Squelch, stsInfo.SignalLevel)
				ctrl.SendToRadioMsgChannel([]byte("STS," + string(buffer[4:])))
			case "GLG":
			case "GLT":
				switch getXmlGLTFormatType(buffer[11:]) {
				case GltXmlFL:
					gltFl := GltFLInfo{}
					if decodeErr := xml.Unmarshal(buffer[11:], &gltFl); decodeErr != nil {
						log.Errorln("Failed to decode XML", decodeErr)
					} else {
						for fl := 0; fl < len(gltFl.FL); fl++ {
							log.Infof("GLT,FL[%d]: Name: %s, Index: %s, Monitor: %s",
								fl+1, gltFl.FL[fl].Name, gltFl.FL[fl].Index, gltFl.FL[fl].Monitor)
						}
						ctrl.SendToRadioMsgChannel([]byte("GLT,FL" + string(buffer[11:])))
					}
				case GltXmlSYS:
					gltSys := GltSysInfo{}
					if decodeErr := xml.Unmarshal(buffer[11:], &gltSys); decodeErr != nil {
						log.Errorln("Failed to decode XML", decodeErr)
					} else {
						for sys := 0; sys < len(gltSys.SYS); sys++ {
							log.Infof("GLT,SYS[%d]: Name: %s, Index: %s, TrunkID: %s, Type: %s",
								sys+1, gltSys.SYS[sys].Name, gltSys.SYS[sys].Index, gltSys.SYS[sys].TrunkId, gltSys.SYS[sys].Type)
						}
						ctrl.SendToRadioMsgChannel([]byte("GLT,SYS" + string(buffer[11:])))
					}

				case GltXmlDEPT:
					gltDept := GltDeptInfo{}
					if decodeErr := xml.Unmarshal(buffer[11:], &gltDept); decodeErr != nil {
						log.Errorln("Failed to decode XML", decodeErr)
					} else {
						for dpt := 0; dpt < len(gltDept.DEPT); dpt++ {
							log.Infof("GLT,DEPT[%d]: Name: %s, Index: %s, TGroupID: %s",
								dpt+1, gltDept.DEPT[dpt].Name, gltDept.DEPT[dpt].Index, gltDept.DEPT[dpt].TGroupId)
						}
						ctrl.SendToRadioMsgChannel([]byte("GLT,DEPT" + string(buffer[11:])))
					}
				case GltXmlSITE:
					gltSite := GltSiteInfo{}
					if decodeErr := xml.Unmarshal(buffer[11:], &gltSite); decodeErr != nil {
						log.Errorln("Failed to decode XML", decodeErr)
					} else {
						for site := 0; site < len(gltSite.SITE); site++ {
							log.Infof("GLT,SITE[%d]: Name: %s, Index: %s, SiteId: %s",
								site+1, gltSite.SITE[site].Name, gltSite.SITE[site].Index, gltSite.SITE[site].SiteId)
						}
						ctrl.SendToRadioMsgChannel([]byte("GLT,SITE" + string(buffer[11:])))
					}
				case GltXmlFTO:
					gltFTO := GltFto{}
					if decodeErr := xml.Unmarshal(buffer[11:], &gltFTO); decodeErr != nil {
						log.Errorln("Failed to decode XML", decodeErr)
					} else {
						for fto := 0; fto < len(gltFTO.FTO); fto++ {
							log.Infof("GLT,FTO[%d]: Name: %s, Index: %s, Freq: %s, Mod: %s, ToneA: %s, ToneB: %s",
								fto+1, gltFTO.FTO[fto].Name, gltFTO.FTO[fto].Index, gltFTO.FTO[fto].Freq, gltFTO.FTO[fto].Mod, gltFTO.FTO[fto].ToneA, gltFTO.FTO[fto].ToneB)
						}
						ctrl.SendToRadioMsgChannel([]byte("GLT,FTO" + string(buffer[11:])))
					}
				case GltXmlCSBANK:
					gltCSBank := GltCSBank{}
					if decodeErr := xml.Unmarshal(buffer[11:], &gltCSBank); decodeErr != nil {
						log.Errorln("Failed to decode XML", decodeErr)
					} else {
						for csb := 0; csb < len(gltCSBank.CSBANK); csb++ {
							log.Infof("GLT,CSBANK[%d]: Name: %s, Index: %s, Lower: %s, Upper: %s, Mod: %s, Step: %s",
								csb+1, gltCSBank.CSBANK[csb].Name, gltCSBank.CSBANK[csb].Index, gltCSBank.CSBANK[csb].Lower, gltCSBank.CSBANK[csb].Upper, gltCSBank.CSBANK[csb].Mod, gltCSBank.CSBANK[csb].Step)
						}
						ctrl.SendToRadioMsgChannel([]byte("GLT,CS_BANK" + string(buffer[11:])))
					}
				case GltXmlTRN_DISCOV:
					gltTrnDisc := GltTrnDiscovery{}
					if decodeErr := xml.Unmarshal(buffer[11:], &gltTrnDisc); decodeErr != nil {
						log.Errorln("Failed to decode XML", decodeErr)
					} else {
						for td := 0; td < len(gltTrnDisc.TRNDISCOV); td++ {
							log.Infof("GLT,TRN_DISCOV: Name: %s, Delay: %s, Logging: %s, Duration: %s, CompareDB: %s, SystemName: %s SystemType: %s SiteName: %s, TimeOutTimer: %s, AutoStore: %s",
								gltTrnDisc.TRNDISCOV[td].Name, gltTrnDisc.TRNDISCOV[td].Delay, gltTrnDisc.TRNDISCOV[td].Logging, gltTrnDisc.TRNDISCOV[td].Duration, gltTrnDisc.TRNDISCOV[td].CompareDB, gltTrnDisc.TRNDISCOV[td].SystemName, gltTrnDisc.TRNDISCOV[td].SystemType, gltTrnDisc.TRNDISCOV[td].SiteName, gltTrnDisc.TRNDISCOV[td].TimeOutTimer, gltTrnDisc.TRNDISCOV[td].AutoStore)
						}
						ctrl.SendToRadioMsgChannel([]byte("GLT,TRN_DISCOV" + string(buffer[11:])))
					}
				case GltXmlCNV_DISCOV:
					gltCnvDisc := GltCnvDiscovery{}
					if decodeErr := xml.Unmarshal(buffer[11:], &gltCnvDisc); decodeErr != nil {
						log.Errorln("Failed to decode XML", decodeErr)
					} else {
						for cd := 0; cd < len(gltCnvDisc.CNVDISCOV); cd++ {
							log.Infof("GLT,CNV_DISCOV: Name: %s, Lower: %s, Upper: %s, Mod: %s, Step: %s, Delay: %s Logging: %s CompareDB: %s, Duration: %s, TimeOutTimer: %s, AutoStore: %s", gltCnvDisc.CNVDISCOV[cd].Name, gltCnvDisc.CNVDISCOV[cd].Lower, gltCnvDisc.CNVDISCOV[cd].Upper, gltCnvDisc.CNVDISCOV[cd].Mod, gltCnvDisc.CNVDISCOV[cd].Step, gltCnvDisc.CNVDISCOV[cd].Delay, gltCnvDisc.CNVDISCOV[cd].Logging, gltCnvDisc.CNVDISCOV[cd].CompareDB, gltCnvDisc.CNVDISCOV[cd].Duration, gltCnvDisc.CNVDISCOV[cd].TimeOutTimer, gltCnvDisc.CNVDISCOV[cd].AutoStore)
							ctrl.SendToRadioMsgChannel([]byte("GLT,CNV_DISCOV" + string(buffer[11:])))
						}
					}
				case GltXmlUREC_FOLDER:
					gltUrecFolder := GltUrecFolder{}
					if decodeErr := xml.Unmarshal(buffer[11:], &gltUrecFolder); decodeErr != nil {
						log.Errorln("Failed to decode XML", decodeErr)
					} else {
						for fi := 0; fi < len(gltUrecFolder.URECFOLDER); fi++ {
							log.Infof("GLT,UREC_FOLDER: Name: %s, Index: %s, Text: %s",
								gltUrecFolder.URECFOLDER[fi].Name, gltUrecFolder.URECFOLDER[fi].Index, gltUrecFolder.URECFOLDER[fi].Text)
						}
						ctrl.SendToRadioMsgChannel([]byte("GLT,UREC_FOLDER" + string(buffer[11:])))
					}

				default:
					log.Infoln("Unhandled GltXml Type", buffer)
				}
			case "VOL":
				log.Infoln("VOL: Volume", string(buffer[4:]))
				ctrl.SendToRadioMsgChannel([]byte("VOL," + string(buffer[4:])))
			case "SQL":
				log.Infoln("SQL: Squelch", string(buffer[4:]))
				ctrl.SendToRadioMsgChannel([]byte("SQL," + string(buffer[4:])))
			case "PWR":
				log.Infoln("PWR: Power", string(buffer[4:]))
				ctrl.SendToRadioMsgChannel([]byte("PWR," + string(buffer[4:])))
			case "GSI":
				si := ScannerInfo{}
				if decodeErr := xml.Unmarshal(buffer[11:], &si); decodeErr != nil {
					log.Errorln("Failed to decode XML", decodeErr)
				} else {
					log.Infof("GSI: System: %s, Department: %s, Site: %s, Freq: [%s] Mon: [%s] Mode: [%s]",
						si.System.Name, si.Department.Name, si.Site.Name, si.SiteFrequency.Freq, si.MonitorList.Name, si.Mode)
					ctrl.SendToRadioMsgChannel([]byte("GSI," + string(buffer[11:])))
				}
			case "PSI":
				switch {
				case string(buffer[4:6]) == "OK":
					ctrl.mode.PSI = false
					log.Infoln("PSI: Stopped")
				case string(buffer[4:9]) == "<XML>":
					ctrl.mode.PSI = true
					si := ScannerInfo{}
					if decodeErr := xml.Unmarshal(buffer[11:], &si); decodeErr != nil {
						log.Errorln("Failed to decode XML", decodeErr)
					} else {
						log.Infof("GSI: System: %s, Department: %s, Site: %s, Freq: [%s] Mon: [%s] Mode: [%s]",
							si.System.Name, si.Department.Name, si.Site.Name, si.SiteFrequency.Freq, si.MonitorList.Name, si.Mode)
						ctrl.SendToRadioMsgChannel([]byte("PSI," + string(buffer[11:])))
					}
				default:
					log.Infoln("PSI: Invalid Mode::", string(buffer[4:]))
					continue
				}
			case "KEY":
				log.Infoln("KEY", string(buffer[4:]))
				ctrl.SendToRadioMsgChannel([]byte("KEY," + string(buffer[4:])))

			case "ERR":
				log.Errorln("Scanner threw an error!")

			default:
				log.Infoln("Unhandled Key", msgType)
			}

			select {
			case <-ctrl.rq:
				log.Infoln("Shutting down reader...")
				do_quit = true
			default:
				log.Infoln("Sleeping")
				time.Sleep(time.Millisecond * ctrl.GoProcDelay)
			}
		}
	}(ctrl)

	var wsErr error
	ctrl.s, wsErr = startWSServer("", *wsPortNum, ctrl)
	if wsErr != nil {
		log.Fatalln("Failed to start WebSocket server", wsErr)
	}

	signal.Notify(ctrl.c, os.Interrupt)
	<-ctrl.c

	// gracefully terminate go routines
	log.Infoln("Terminating on signal...")
	ctrl.rq <- true
	ctrl.wq <- true

	// log.Infoln("Waiting to drain...")
	// <-ctrl.drained
	// ctrl.drain()
	// log.Infoln("Drained UDP, Closing Connection...")
	// ctrl.conn.Close()
	const timeout = 5 * time.Second

	log.Infoln("Shutting down WebSocket server")

	ctx, _ := context.WithTimeout(context.Background(), timeout)
	if err := ctrl.s.Shutdown(ctx); err != nil {
		log.Errorln("Failed to Shutdown", err)
	}
	log.Infoln("UDPServer Terminated.")
}
