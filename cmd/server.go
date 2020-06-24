package cmd

import (
	"net"
	"os"
	"path/filepath"

	"github.com/Bearcatter/bearcatter/server"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var serverUdpAddress string
var serverUdpPortNumber int
var serverUsbPath string
var serverRecordingPath string

var serverCfg = &server.Config{}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Bearcatter WebSocket server",
	Long:  `The Bearcatter server provides a bridge between your Uniden scanner and third party WebSocket clients. It also handles automatically copying recordings in real time from your scanner via miniUSB.`,
	Run: func(cmd *cobra.Command, args []string) {
		absRecordingsPath, absRecordingsPathErr := filepath.Abs(serverRecordingPath)
		if absRecordingsPathErr != nil {
			log.Fatalln("Error when attempting to resolve recordings path")
		}

		if _, err := os.Stat(absRecordingsPath); os.IsNotExist(err) {
			if mkdirErr := os.Mkdir(absRecordingsPath, os.ModeDir); mkdirErr != nil {
				log.Fatalf("Error when creating recordings directory %s: %v\n", absRecordingsPath, mkdirErr)
			}
		}

		serverCfg.RecordingsPath = absRecordingsPath

		if serverUdpAddress != "" {
			udpIP, udpIPErr := net.ResolveIPAddr("ip", serverUdpAddress)
			if udpIPErr != nil {
				log.Fatalln("Error when processing UDP address", udpIPErr)
			}

			serverCfg.UDPAddress = &net.UDPAddr{
				IP:   udpIP.IP,
				Port: serverUdpPortNumber,
			}
		}

		if serverUsbPath != "" {
			serverCfg.USBPath = serverUsbPath
		}

		if serverCfg.UDPAddress == nil && serverCfg.USBPath == "" {
			log.Fatal("UDP IP address or USB path must be set!")
		}
		serverCfg.Serve()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringVarP(&serverUdpAddress, "udp.address", "a", "", "IP address or hostname of SDS200")
	serverCmd.Flags().IntVarP(&serverUdpPortNumber, "udp.port", "p", 50536, "UDP port of SDS200")
	serverCmd.Flags().StringVarP(&serverUsbPath, "usb.path", "u", "", "Path to SDS100 USB port")

	serverCmd.Flags().IntVar(&serverCfg.WebSocketPort, "websocket.port", 8080, "WebSocket port to accept connections on")

	serverCmd.Flags().StringVarP(&serverRecordingPath, "recordings.path", "r", "audio", "Path to store recordings in")
	if markErr := serverCmd.MarkFlagDirname("recordings.path"); markErr != nil {
		log.Fatalln("Error when marking recordings directory as only accepting dir names", markErr)
	}
}
