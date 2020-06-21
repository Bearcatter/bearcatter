package main

import (
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

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
	conn             *ScannerConn
	s                *http.Server
	c                chan os.Signal
	GoProcDelay      time.Duration
	GoProcMultiplier time.Duration
	mode             Modal
	incomingFile     *AudioFeedFile
}

func (s *ScannerCtrl) IsLocked() bool {
	var locked bool
	s.locker.Lock()
	if s.locker.state == true {
		locked = true
		log.Tracef("UDP Packets Sent: [%d] UDP Packets Recv: [%d]", s.locker.pktSent, s.locker.pktRecv)
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
		log.Infof("Received Message: [%s] To Send To Host: %s", elapsed, msgToHost.msg)
		return msgToHost, true
	default:
		time.Sleep(time.Millisecond * 50)
	}
	return MsgPacket{}, false
}

func (s *ScannerCtrl) SendToRadioMsgChannel(msg []byte) bool {

	if !s.IsLocked() {
		log.Infoln("RadioMsgChannel: No Listener to Receive Msg, Msg Not Sent")
		return false
	}

	select {
	case s.radioMsg <- MsgPacket{msg, time.Now()}:
		log.Infof("Send Msg[ql=%d]: [%s] to Radio Msg Channel", len(s.radioMsg), msg)
		return true
	default:
		log.Infof("Queue Full, No Message Sent: %d", len(s.radioMsg))
		time.Sleep(time.Millisecond * 50)
	}
	return false
}

func (s *ScannerCtrl) SendToHostMsgChannel(msg []byte) bool {

	if !s.IsLocked() {
		log.Warnln("HostMsgChannel: No Listener to Receive Msg, Msg Not Sent")
		// return false
	}

	select {
	case s.hostMsg <- MsgPacket{msg, time.Now()}:
		return true
	default:
		log.Infof("Queue Full, No Message Sent: %d", len(s.radioMsg))
		time.Sleep(time.Millisecond * 50)
	}
	return false
}

func (c *ScannerCtrl) drain() {

	if c.conn.Type == ConnTypeUSB {
		if flushErr := c.conn.usbConn.Flush(); flushErr != nil {
			log.Fatalln("Error while flushing USB", flushErr)
		}
		return
	}

	buffer := make([]byte, 8192)
	for {
		// c.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		_, err := c.conn.Read(buffer)

		if err != nil {
			if e, ok := err.(net.Error); !ok || !e.Timeout() {
			} else {
				log.Infoln("Drained...")
				return
			}
		} else {
			log.Infoln("Packet Draining on WS Close...")
		}
	}
}

func CreateScannerCtrl() *ScannerCtrl {

	ctrl := &ScannerCtrl{}

	ctrl.rq = make(chan bool, 1)
	ctrl.wq = make(chan bool, 1)
	ctrl.drained = make(chan bool, 1)

	ctrl.radioMsg = make(chan MsgPacket, 100)
	ctrl.hostMsg = make(chan MsgPacket, 100)
	ctrl.c = make(chan os.Signal)
	ctrl.GoProcDelay = DefaultGoProcDelay
	ctrl.GoProcMultiplier = DefaultGoProcMultiplier
	return ctrl
}
