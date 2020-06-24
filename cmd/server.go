package cmd

import (
	"net"
	"os"
	"path/filepath"

	"github.com/Bearcatter/bearcatter/server"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var serverCfg = &server.Config{}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Bearcatter WebSocket server",
	Long:  `The Bearcatter server provides a bridge between your Uniden scanner and third party WebSocket clients. It also handles automatically copying recordings in real time from your scanner via miniUSB.`,
	Run: func(cmd *cobra.Command, args []string) {
		if serverCfg.UDPAddress == nil && serverCfg.USBPath == "" {
			log.Fatal("UDP IP address or USB path must be set!")
		}
		serverCfg.Serve()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	udpAddress := serverCmd.Flags().StringP("udp.address", "a", "", "IP address or hostname of SDS200")
	udpPortNumber := serverCmd.Flags().IntP("udp.port", "p", 50536, "UDP port of SDS200")
	usbPath := serverCmd.Flags().StringP("usb.path", "u", "", "Path to SDS100 USB port")

	serverCmd.Flags().IntVar(&serverCfg.WebSocketPort, "websocket.port", 8080, "WebSocket port to accept connections on")

	rPath := serverCmd.Flags().StringP("recordings.path", "r", "audio", "Path to store recordings in")
	if markErr := serverCmd.MarkFlagDirname("recordings.path"); markErr != nil {
		log.Fatalln("Error when marking recordings directory as only accepting dir names", markErr)
	}

	absRecordingsPath, absRecordingsPathErr := filepath.Abs(*rPath)
	if absRecordingsPathErr != nil {
		log.Fatalln("Error when attempting to resolve recordings path")
	}

	if _, err := os.Stat(absRecordingsPath); os.IsNotExist(err) {
		if mkdirErr := os.Mkdir(absRecordingsPath, os.ModeDir); mkdirErr != nil {
			log.Fatalf("Error when creating recordings directory %s: %v\n", absRecordingsPath, mkdirErr)
		}
	}

	if udpAddress != nil && *udpAddress != "" {
		udpIP, udpIPErr := net.ResolveIPAddr("ip", *udpAddress)
		if udpIPErr != nil {
			log.Fatalln("Error when processing UDP address", udpIPErr)
		}

		serverCfg.UDPAddress = &net.UDPAddr{
			IP:   udpIP.IP,
			Port: *udpPortNumber,
		}
	}

	if usbPath != nil && *usbPath != "" {
		serverCfg.USBPath = *usbPath
	}
}
