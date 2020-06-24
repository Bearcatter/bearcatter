package server

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	UDPAddress     *net.UDPAddr
	USBPath        string
	WebSocketPort  int
	RecordingsPath string
}

func (c *Config) Serve() {
	ctrl := CreateScannerCtrl()
	var connCreateErr error

	if c.UDPAddress != nil {
		ctrl.conn, connCreateErr = NewUDPConn(c.UDPAddress)
	} else if c.USBPath != "" {
		ctrl.conn, connCreateErr = NewUSBConn(c.USBPath)
	} else {
		log.Fatal("IP address or USB path must be set!")
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
				log.Debugf("Host->Scanner:[ql=%d]: [%s]: [%#q]", len(ctrl.hostMsg), elapsed, msgToRadio.msg)
				if _, writeErr := ctrl.conn.Write(msgToRadio.msg); writeErr != nil {
					log.Errorln("Error Writing to scanner", writeErr)
					continue
				}
				ctrl.locker.pktSent++

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
			log.Debugf("Scanner->Host:[ql=%d]: [%#q]\n", len(ctrl.hostMsg), buffer[0:n])
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
				comma := ","
				if xmlMessageType == "GSI" || xmlMessageType == "PSI" {
					comma = ",,"
				}
				buffer = []byte(fmt.Sprintf(`%s,<XML>%s%s`, xmlMessageType, comma, string(xmlMessage)))
				isXML = false
				xmlMessageType = ""
				xmlMessage = make([]byte, 0)
			}

			msgType := string(buffer[:3])

			ctrl.locker.pktRecv++

			if buffer[3] == byte('\t') {
				log.Warnln("Received a HomePatrol message!")
			}

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
				log.Infof("URC: Recording? %t, ErrorCode: %d, ErrorMessage: %s\n", recStatus.Recording, recStatus.ErrorCode, *recStatus.ErrorMessage)
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
			// HomePatrol Commands
			case "RMT":
				log.Infoln(msgType, string(buffer))
				ctrl.SendToRadioMsgChannel(buffer)
			case "AUF":
				split := strings.Split(string(buffer[0:n]), "\t")

				hpCmd := split[1]

				if len(split) > 2 {
					if split[2] == "ERR" {
						log.Warnln("Scanner threw ERR during file transfer!")
						continue
					} else if split[2] == "NG" {
						log.Warnln("Scanner said last command was invalid during file transfer")
						continue
					}
				}

				switch hpCmd {
				case "ERR":
					log.Warnln("Scanner threw DATA ERR during file transfer!")
					continue
				case "NG":
					log.Warnln("Scanner said last command was invalid during DATA")
					continue
				case "STS":
					ctrl.SendToRadioMsgChannel(buffer)
				case "INFO":
					ctrl.SendToHostMsgChannel([]byte(homepatrolCommand([]string{"AUF", "INFO", "ACK"})))

					newFile, newFileErr := NewAudioFeedFile(split[2:])
					if newFileErr != nil {
						if newFileErr != ErrNoFile {
							log.Errorln("Error processing new file notification!", newFileErr)
						}
						continue
					}
					log.Debugf("Incoming file %v\n", newFile)
					ctrl.incomingFile = newFile

					ctrl.SendToHostMsgChannel([]byte(homepatrolCommand([]string{"AUF", "DATA"})))
				case "DATA":
					dataSubCmd := split[2]
					switch dataSubCmd {
					case "EOT":
						// End of transmission
						log.Infoln("Finished receiving file!")

						ctrl.SendToHostMsgChannel([]byte(homepatrolCommand([]string{"AUF", "DATA", "ACK"})))

						ctrl.incomingFile.Finished = true

						filePath := fmt.Sprintf("%s/%s", c.RecordingsPath, ctrl.incomingFile.Name)

						if saveAudioErr := ioutil.WriteFile(filePath, ctrl.incomingFile.Data, 0777); saveAudioErr != nil {
							log.Errorln("Error when saving audio file!", saveAudioErr)
						}

						if metadataErr := ctrl.incomingFile.ParseMetadata(filePath); metadataErr != nil {
							log.Errorln("Error when parsing metadata", metadataErr)
							continue
						}

						metadataJSON, metadataJSONErr := json.MarshalIndent(&ctrl.incomingFile.Metadata, "", "    ")
						if metadataJSONErr != nil {
							log.Errorln("Error when marshalling metadata", metadataJSONErr)
							continue
						}

						if saveMetadataErr := ioutil.WriteFile(fmt.Sprintf("%s.json", filePath), metadataJSON, 0777); saveMetadataErr != nil {
							log.Errorln("Error when saving metadata file!", saveMetadataErr)
						}

						break
					case "CAN":
						log.Warnln("File transfer canceled")
					default: // Receiving data
						blockNum := split[2]
						fileData := split[3]
						log.Infof("Receiving file block %s with file length %d\n", blockNum, len(fileData))
						hexData, hexDataErr := hex.DecodeString(fileData)
						if hexDataErr != nil {
							log.Errorln("Error when converting incoming file chunk to hex", hexDataErr)
						}
						ctrl.incomingFile.Data = append(ctrl.incomingFile.Data, hexData...)

						time.Sleep(50 * time.Millisecond)

						ctrl.SendToHostMsgChannel([]byte(homepatrolCommand([]string{"AUF", "DATA", "ACK"})))
						if !ctrl.incomingFile.Finished {
							time.Sleep(50 * time.Millisecond)
							ctrl.SendToHostMsgChannel([]byte(homepatrolCommand([]string{"AUF", "DATA"})))
						}
					}
				}

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
				log.Traceln("Sleeping")
				time.Sleep(time.Millisecond * ctrl.GoProcDelay)
			}
		}
	}(ctrl)

	if ctrl.conn.Type == ConnTypeUSB {
		ticker := time.NewTicker(1 * time.Second)
		go func(ctrl *ScannerCtrl) {
			time.Sleep(1 * time.Second)
			success := ctrl.SendToHostMsgChannel([]byte(homepatrolCommand([]string{"AUF", "STS", "ON"})))
			log.Infoln("success", success)

			time.Sleep(1 * time.Second)

			for {
				select {
				case <-ticker.C:
					ctrl.SendToHostMsgChannel([]byte(homepatrolCommand([]string{"AUF", "INFO"})))
				case <-ctrl.rq:
					log.Infoln("Shutting down file polling")
					ticker.Stop()
					return
				}
			}
		}(ctrl)
	}

	var wsErr error
	ctrl.s, wsErr = startWSServer("", c.WebSocketPort, ctrl)
	if wsErr != nil {
		log.Fatalln("Failed to start WebSocket server", wsErr)
	}

	signal.Notify(ctrl.c, os.Interrupt)
	<-ctrl.c

	// gracefully terminate go routines
	log.Infoln("Terminating on signal...")

	if ctrl.conn.Type == ConnTypeUSB {
		if ctrl.incomingFile != nil && ctrl.incomingFile.Finished {
			log.Infoln("Terminating file transfer session")
			ctrl.SendToHostMsgChannel([]byte(homepatrolCommand([]string{"AUF", "INFO", "CAN"})))
			ctrl.SendToHostMsgChannel([]byte(homepatrolCommand([]string{"AUF", "DATA", "CAN"})))
		}

		ctrl.SendToHostMsgChannel([]byte(homepatrolCommand([]string{"AUF", "STS", "OFF"})))
		time.Sleep(50 * time.Millisecond)
	}

	ctrl.rq <- true
	ctrl.wq <- true

	// log.Infoln("Waiting to drain...")
	// <-ctrl.drained
	// ctrl.drain()
	// log.Infoln("Drained UDP, Closing Connection...")
	// ctrl.conn.Close()
	const timeout = 5 * time.Second

	log.Infoln("Shutting down WebSocket server")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	if err := ctrl.s.Shutdown(ctx); err != nil {
		log.Errorln("Failed to Shutdown", err)
	}
	cancel()
	log.Infoln("UDPServer Terminated.")
}
