package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/tarm/serial"
)

const (
	ConnTypeUSB     = "usb"
	ConnTypeNetwork = "network"
)

type ScannerConn struct {
	Type string

	udpAddress *net.UDPAddr
	udpConn    *net.UDPConn

	usbPort   string
	usbConfig *serial.Config
	usbConn   *serial.Port
	usbReader *bufio.Scanner
}

func NewUDPConn(addr *net.UDPAddr) (*ScannerConn, error) {
	return &ScannerConn{
		Type:       ConnTypeNetwork,
		udpAddress: addr,
	}, nil
}

func NewUSBConn(path string) (*ScannerConn, error) {
	return &ScannerConn{
		Type:      ConnTypeUSB,
		usbPort:   path,
		usbConfig: &serial.Config{Name: path, Baud: 115200},
	}, nil
}

func (c *ScannerConn) Open() error {
	var connErr error
	if c.Type == ConnTypeNetwork {
		c.udpConn, connErr = net.DialUDP("udp", nil, c.udpAddress)
	} else if c.Type == ConnTypeUSB {
		c.usbConn, connErr = serial.OpenPort(c.usbConfig)
		c.usbReader = bufio.NewScanner(c.usbConn)
		c.usbReader.Split(ScanLinesWithCR)
	}
	return connErr
}

func (c ScannerConn) Close() error {
	if c.Type == ConnTypeNetwork {
		return c.udpConn.Close()
	} else if c.Type == ConnTypeUSB {
		return c.usbConn.Close()
	}
	return nil
}

func (c ScannerConn) Write(b []byte) (n int, err error) {
	if c.Type == ConnTypeNetwork {
		return c.udpConn.Write(b)
	} else if c.Type == ConnTypeUSB {
		fixed := []byte(strings.Replace(string(b), "\n", "\r", -1))
		return c.usbConn.Write(fixed)
	}
	return 0, nil
}

func (c ScannerConn) Read(b []byte) (n int, err error) {
	if c.Type == ConnTypeNetwork {
		return c.udpConn.Read(b)
	} else if c.Type == ConnTypeUSB {
		if !c.usbReader.Scan() {
			return 0, nil
		}
		scanned := c.usbReader.Bytes()
		copy(b, scanned)
		return len(scanned), nil
	}
	return 0, nil
}

func (c ScannerConn) String() string {
	if c.Type == ConnTypeNetwork {
		return fmt.Sprintf("Connected to %s via UDP", c.udpAddress)
	} else if c.Type == ConnTypeUSB {
		return fmt.Sprintf("Connected to %s via USB", c.usbPort)
	}
	return "unknown"
}
