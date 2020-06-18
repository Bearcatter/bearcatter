package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jroimartin/gocui"
)

var (
	done       chan struct{}
	conn       *net.UDPConn
	ErrTimeout = errors.New("timeout on Read\n")
)

const (
	LF               = 0x01
	NL               = 0x02
	UDP_READ_TIMEOUT = 500
)

type UDPMsg struct {
	err bool
	msg string
}

func main() {

	portNum := flag.Int("port", 50536, "udp port for SDS200")
	hostName := flag.String("host", "192.168.1.26", "hostname of SDS200")
	sdscmds := flag.Bool("commands", false, "print SDS200 commands help and exit")
	flag.Parse()

	if *sdscmds {
		displayHelp(nil)
		os.Exit(0)
	}

	service := *hostName + ":" + strconv.Itoa(*portNum)
	RemoteAddr, err := net.ResolveUDPAddr("udp", service)
	if err != nil {
		log.Fatal("Can't resolve address: %s\n", err)
	}
	conn, err = net.DialUDP("udp", nil, RemoteAddr)
	if err != nil {
		log.Fatal("Can't dial address: %s\n", err)
	}
	defer conn.Close()
	done = make(chan struct{})

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	go updateUDPOutput(g)
	//go counter(g)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if _, err := g.SetView("main", 0, 0, maxX-1, maxY-10); err != nil &&
		err != gocui.ErrUnknownView {
		return err
	}

	if _, err := g.SetView("errors", 0, maxY-9, maxX-1, maxY-5); err != nil &&
		err != gocui.ErrUnknownView {
		return err
	}
	if _, err := g.SetView("cmdline", 0, maxY-5, maxX-1, maxY-3); err != nil &&
		err != gocui.ErrUnknownView {
		return err
	}

	return nil
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	if err := g.SetKeybinding("cmdline", gocui.KeyEnter, gocui.ModNone, getLine); err != nil {
		return err
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	close(done)
	return gocui.ErrQuit
}

func getLine(g *gocui.Gui, v *gocui.View) error {
	var l string
	var err error

	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}
	v.Clear()
	v.SetCursor(0, 0)

	if v, err = g.View("main"); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	if l == "quit" {
		close(done)
		return gocui.ErrQuit
	} else if l == "help" {
		displayHelp(v)
	} else if l == "clear" {
		v.Clear()
	} else {

		l = crlfStrip([]byte(l), NL|LF)
		if len(l) > 0 {
			fmt.Fprintf(v, fmt.Sprintf("recv: [%s]\n", l))
			writeCmd(l)
		} else {
			fmt.Fprintf(v, fmt.Sprintf("idle\n"))
		}
	}
	if _, err := g.SetCurrentView("cmdline"); err != nil {
		return err
	}
	return nil
}

func formatted(msg string) []byte {
	return []byte(msg + "\r")
}

func byteString(msg []byte, n int) string {
	return crlfStrip([]byte(string(msg[:n])), NL|LF)
}

func crlfStrip(msg []byte, flags uint) string {

	var replacer []string
	switch {
	case (flags & LF) == LF:
		replacer = append(replacer, "\r", "\\r")
		fallthrough
	case (flags & NL) == NL:
		replacer = append(replacer, "\n", "")
	}
	r := strings.NewReplacer(replacer...)
	return r.Replace(string(msg))
}

func counter(g *gocui.Gui) {

	for {
		select {
		case <-done:
			return
		case <-time.After(50 * time.Millisecond):

			g.Update(func(g *gocui.Gui) error {
				e, err := g.View("errors")
				if err != nil {
					return err
				}
				e.Wrap = true
				e.Autoscroll = true
				e.Title = "errors"

				v, err := g.View("main")
				if err != nil {
					return err
				}
				//v.Editable = true
				v.Wrap = true
				v.Autoscroll = true
				v.Title = "--counter--"

				if _, err := g.SetCurrentView("cmdline"); err != nil {
					return err
				}

				return nil
			})
		}
	}
}

func updateUDPOutput(g *gocui.Gui) {

	var msg chan UDPMsg = make(chan UDPMsg)
	var cnt int = 0
	var ect int = 0

	go func() {
		for {
			l, err := readUDPConn()
			if err != nil && err != ErrTimeout {
				msg <- UDPMsg{err: true, msg: l}
				ect++
			} else if len(l) > 0 {
				cnt++
				msg <- UDPMsg{err: false, msg: l}
			}

			time.Sleep(time.Millisecond * 25)
		}
	}()

	type Queue struct {
		q []UDPMsg
		sync.Mutex
	}

	var queue Queue
	for {

		select {
		case <-done:
			return
		case m := <-msg:
			queue.Lock()
			queue.q = append(queue.q, m)
			queue.Unlock()
		default:
			g.Update(func(g *gocui.Gui) error {
				e, err := g.View("errors")
				if err != nil {
					return err
				}
				v, err := g.View("main")
				if err != nil {
					return err
				}
				v.Autoscroll = true
				v.Wrap = true
				if len(queue.q) > 0 {
					queue.Lock()
					m := queue.q[0]
					copy(queue.q[0:], queue.q[1:])
					queue.q = queue.q[:len(queue.q)-1]
					queue.Unlock()

					if m.err == true {
						e.Title = "Error:"
						fmt.Fprintf(e, fmt.Sprintf("error msg: %s\n", m.msg))
					} else if len(m.msg) > 0 {
						v.Title = "UDP Packet Arrived: "
						fmt.Fprintf(v, fmt.Sprintf("%s\n", byteString([]byte(m.msg), len(m.msg))))
					}
				} else {
					v.Title = "Waiting for UDP Packet: "
				}
				v, err = g.View("cmdline")
				v.Editable = true
				v.Wrap = true
				if _, err := g.SetCurrentView("cmdline"); err != nil {
					return err
				}

				return nil
			})
			time.Sleep(time.Millisecond * 25)
		}
	}
}

func writeCmd(input string) {

	_, err := conn.Write(formatted(input))
	if err != nil {
		log.Fatal("Write failed: %s\n", err)
	}
}

func readUDPConn() (string, error) {

	var output string
	buffer := make([]byte, 65535)
	conn.SetReadDeadline(time.Now().Add(UDP_READ_TIMEOUT * time.Millisecond))
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		if e, ok := err.(net.Error); !ok || !e.Timeout() {
			//log.Printf("Error on ReadFromUDP: %s, %d\n", e, n)
			return "", errors.New("error on Read\n")
		} else {
			//log.Printf("timedout on UDP Read\n")
			return "", ErrTimeout
		}
	} else {
		// process packet
		output = byteString(buffer, n)
	}

	return output, nil
}

func displayHelp(v interface{}) {

	var f io.Writer
	if v == nil {
		f = os.Stdin
	} else {
		f = v.(*gocui.View)
	}
	fmt.Fprintf(f, "MDL\t\tGet Model Info\n")
	fmt.Fprintf(f, "VER\t\tGet Firmware Version\n")
	fmt.Fprintf(f, "VOL\t\tGet/Set Volume\n")
	fmt.Fprintf(f, "SQL\t\tGet/Set Squelch\n")
	fmt.Fprintf(f, "KEY\t\tPush KEY - KEY,{key-code},{mode}, where mode is PUSH or HOLD\n")
	fmt.Fprintf(f, "QSH\t\tGo to quick search hold mode\n")
	fmt.Fprintf(f, "STS\t\tGet Current Status\n")
	fmt.Fprintf(f, "JNT\t\tJump Number tag\n")
	fmt.Fprintf(f, "NXT\t\tNext\n")
	fmt.Fprintf(f, "PRV\t\tPrevious\n")
	fmt.Fprintf(f, "FQK\t\tGet/Set Favorites List Quick Keys Status\n")
	fmt.Fprintf(f, "SQK\t\tGet/Set System Quick Keys Status\n")
	fmt.Fprintf(f, "DQK\t\tGet/Set Department Quick Keys Status\n")
	fmt.Fprintf(f, "PSI,[ms]\t\tPush Scanner Information - ms is the update rate\n")
	fmt.Fprintf(f, "GSI\t\tGet Scanner Information\n")
	fmt.Fprintf(f, "GLT,[list]\t\tGet xxx list, FL, SYS, SITE, DEPT, etc... \n")
	fmt.Fprintf(f, "HLD\t\tHold\n")
	fmt.Fprintf(f, "AVD\t\tSet Avoid Option\n")
	fmt.Fprintf(f, "SVC\t\tGet/Set Service Type Settings\n")
	fmt.Fprintf(f, "JPM\t\tJump Mode\n")
	fmt.Fprintf(f, "DTM\t\tGet/Set Date and Time.\n")
	fmt.Fprintf(f, "LCR\t\tGet/Set Location and range.\n")
	fmt.Fprintf(f, "AST\t\tAnalize Start\n")
	fmt.Fprintf(f, "APR\t\tAnalize Pauze/Resume\n")
	fmt.Fprintf(f, "URC[0|1]\t\tUser Record Control. 0 to stop, 1 to start\n")
	fmt.Fprintf(f, "MNU\t\tMenu Mode command. i.e. MNU,TOP\n")
	fmt.Fprintf(f, "MSI\t\tMenu Status Info\n")
	fmt.Fprintf(f, "MSV\t\tMenu Set Value. i.e. MSV,,10\n")
	fmt.Fprintf(f, "MSB\t\tMenu Structure Back. i.e. MSB,, to pop to the previous menu\n")
}
