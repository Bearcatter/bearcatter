package main

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

	udpHostname string
	udpPort     int
	udpAddress  *net.UDPAddr
	udpConn     *net.UDPConn

	usbPort   string
	usbConfig *serial.Config
	usbConn   *serial.Port
	usbReader *bufio.Reader
}

func NewUDPConn(hostname string, port int) (*ScannerConn, error) {
	addr, addrErr := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", hostname, port))
	if addrErr != nil {
		return nil, addrErr
	}

	return &ScannerConn{
		Type:        ConnTypeNetwork,
		udpHostname: hostname,
		udpPort:     port,
		udpAddress:  addr,
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
		c.usbReader = bufio.NewReader(c.usbConn)
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
		return c.usbReader.Read(b)
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
