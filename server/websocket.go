package server

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	log "github.com/sirupsen/logrus"
)

func startWSServer(host string, port int, ctrl *ScannerCtrl) (*http.Server, error) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			// handle error
			log.Errorln("Error during WS upgrade", err)
		}

		ctrl.mode.WSClientConnected = true

		defer func() {
			conn.Close()
			if ctrl.locker.name == r.RemoteAddr {
				ctrl.locker.state = false
				log.Infoln("Unlocking Scanner for new WS Client Connection")
			}
		}()

		if ctrl.locker.state == true {
			log.Infoln("Scanner is already is use by", ctrl.locker.name)
			if writeErr := wsutil.WriteServerMessage(conn, ws.OpBinary, []byte("Locked by "+ctrl.locker.name)); writeErr != nil {
				// handle error
				log.Errorln("Failed to Write", writeErr)
				return
			}
			return
		} else {
			log.Infof("Locking Scanner for new WS Client Connection: [%s]", r.RemoteAddr)
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
				msgFromHost, _, readErr := wsutil.ReadClientData(conn)

				if readErr != nil {
					// handle error
					log.Errorln("Failed to read from WS, Terminating Client Connection", readErr)
					done <- true
					quitReader <- true
					return
				}

				if !validMsgFromWSClient(msgFromHost) {
					log.Warnln("Message From WS Failed Validation", crlfStrip(msgFromHost, LF))
					return
				}

				strMsg := string(crlfStrip(msgFromHost, LF))
				if strMsg == TERMINATE {
					log.Infoln("Received Client QUIT Command: Terminating Client Connection")
					done <- true
					quitReader <- true
					return
				}

				if strMsg == "" || strMsg == "\n" {
					continue
				}

				if strings.HasPrefix(strMsg, "HP,") {
					log.Infof("HomePatrol message From Host: [%s]", crlfStrip(msgFromHost, LF))
					hpCmd := homepatrolCommand(strings.Split(strMsg[3:], "|"))
					log.Infof("Sending HomePatrol message %#q\n", hpCmd)
					success := ctrl.SendToHostMsgChannel([]byte(hpCmd))
					log.Infoln("Sent message?", success)
					continue
				}
				// log.Infof("Message From Host: [%s]", crlfStrip(msgFromHost, LF))
				ctrl.SendToHostMsgChannel(msgFromHost)
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
					log.Infof("Received[ql=%d] Message To Host at [%s] [%s]",
						len(ctrl.radioMsg), elapsed, crlfStrip(msgToHost.msg, LF|NL))
					if writeErr := wsutil.WriteServerMessage(conn, ws.OpBinary, msgToHost.msg); writeErr != nil {
						// handle error
						log.Errorln("Failed to Write", writeErr)
						done <- true
						quitWriter <- true
						return
					} else {
						log.Infof("Message To Host: [%s]", string(msgToHost.msg))
					}
				}
			}
		}()

		<-done
		// TODO - need to reset radio into normal state:  Turn off PSI
		ctrl.SendToHostMsgChannel([]byte("PSI,0\r"))

		// drain any UDP Traffic
		ctrl.drain()

		// Drain WS Messages so new connection won't receive old messages
		for {
			_, ok := ctrl.ReceiveFromRadioMsgChannel()
			if !ok {
				break
			} else {
				log.Infoln("RadioMsgChannel Draining...")
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
	log.Infoln("Started WebSocket Server at", s.Addr)
	return s, nil
}

// TODO -- add msg validation
func validMsgFromWSClient(msgFromHost []byte) bool {
	return true
}
