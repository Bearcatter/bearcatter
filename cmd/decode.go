package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/Bearcatter/bearcatter/wavparse"
	"github.com/gocarina/gocsv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var recordingsPath string
var outputFileName string
var outputFilePath string
var outputFormat string
var continueOnError bool
var jsonIndent string
var jsonMultipleFiles bool
var jsonMultipleFilesCount int
var csvDelimiter string
var csvUseCRLF bool

// decodeCmd represents the decode command
var decodeCmd = &cobra.Command{
	Use:   "decode",
	Short: "Decode will inspect a directory of WAV files and dump metadata of each to a single file",
	Long: `The decode command will decode every WAV file in the given directory and dump metadata to a CSV or JSON file(s).
Metadata includes publicly documented and reverse engineered fields.`,
	Run: func(cmd *cobra.Command, args []string) {
		outputFormat = strings.ToLower(outputFormat)

		if outputFileName == "recordings.csv" && outputFormat == "json" {
			outputFileName = "recordings.json"
		}

		if outputFormat != "csv" && outputFormat != "json" {
			log.Fatalf(`%s is not a valid output format. Valid options are "csv" or "json"\n`, outputFormat)
		}

		if outputFormat != filepath.Ext(outputFileName)[1:] {
			log.Warnf("Output file name %s does not has output format extension %s\n", outputFileName, outputFormat)
		}

		csvDelimiterRune, _ := utf8.DecodeRuneInString(csvDelimiter)
		if csvDelimiterRune == utf8.RuneError {
			log.Fatalln("output.csv.delimiter can only be a single character")
		}

		if outputFormat == "csv" {
			gocsv.SetCSVWriter(func(out io.Writer) *gocsv.SafeCSVWriter {
				writer := csv.NewWriter(out)
				writer.Comma = csvDelimiterRune
				writer.UseCRLF = csvUseCRLF
				return gocsv.NewSafeCSVWriter(writer)
			})
		}

		var recordingsPathErr error
		recordingsPath, recordingsPathErr = filepath.Abs(recordingsPath)
		if recordingsPathErr != nil {
			log.Fatalln("Error when attempting to resolve recordings path", recordingsPathErr)
		}

		var outputPathErr error
		outputFilePath, outputPathErr = filepath.Abs(outputFileName)
		if outputPathErr != nil {
			log.Fatalln("Error when attempting to resolve output file path", outputPathErr)
		}

		var wavs []string

		if walkErr := filepath.Walk(recordingsPath, findWAVs(&wavs)); walkErr != nil {
			log.Fatalln("Error when walking recordings directory", walkErr)
		}

		log.Infof("Found %d files in %s\n", len(wavs), recordingsPath)

		parsedRecordings := []*wavparse.Recording{}

		errorLogLevel := log.FatalLevel

		if continueOnError {
			errorLogLevel = log.WarnLevel
		}

		for _, filePath := range wavs {
			decoded, decodeErr := wavparse.DecodeRecording(filePath)
			if decodeErr != nil {
				log.StandardLogger().Logln(errorLogLevel, "Error when decoding WAV file", decodeErr)
			}

			if decoded == nil {
				log.StandardLogger().Logf(errorLogLevel, "File %s was not decodable", filePath)
			}

			if jsonMultipleFiles {
				jsonFileName := fmt.Sprintf("%s.json", decoded.File)
				outputFile, outputFileErr := os.OpenFile(jsonFileName, os.O_RDWR|os.O_CREATE, os.ModePerm)
				if outputFileErr != nil {
					log.Fatalf("Error when creating output file %s: %v\n", jsonFileName, outputFileErr)
				}
				defer outputFile.Close()
				if saveErr := save(&decoded, outputFile); saveErr != nil {
					log.StandardLogger().Logf(errorLogLevel, "Error when saving %s file: %v", outputFormat, saveErr)
				}
				jsonMultipleFilesCount += 1
			}

			fmt.Printf(".")

			if !jsonMultipleFiles {
				parsedRecordings = append(parsedRecordings, decoded)
			}
		}
		fmt.Printf("\n")

		if !jsonMultipleFiles {
			outputFile, outputFileErr := os.OpenFile(outputFilePath, os.O_RDWR|os.O_CREATE, os.ModePerm)
			if outputFileErr != nil {
				log.Fatalf("Error when creating output file %s: %v\n", outputFilePath, outputFileErr)
			}
			defer outputFile.Close()

			if saveErr := save(&parsedRecordings, outputFile); saveErr != nil {
				log.StandardLogger().Logf(errorLogLevel, "Error when saving %s file: %v", outputFormat, saveErr)
			}

			log.Infof("Wrote %d lines to %s\n", len(parsedRecordings), outputFilePath)
		} else {
			log.Infof("Wrote %d JSON files\n", jsonMultipleFilesCount)
		}
	},
}

func init() {
	rootCmd.AddCommand(decodeCmd)

	decodeCmd.Flags().BoolVarP(&continueOnError, "continue", "c", true, "Whether to continue exporting if individual file error happens")

	decodeCmd.Flags().StringVar(&csvDelimiter, "output.csv.delimiter", ",", "Field delimiter")

	decodeCmd.Flags().BoolVar(&csvUseCRLF, "output.csv.crlf", false, "True to use \\r\\n as the line terminator")

	decodeCmd.Flags().StringVar(&jsonIndent, "output.json.indent", "\t", "String to indent JSON with. Set to empty string for no indentation.")

	decodeCmd.Flags().BoolVar(&jsonMultipleFiles, "output.json.multiple", false, "If true, one JSON file will be output to the current directory for each WAV file")

	decodeCmd.Flags().StringVarP(&recordingsPath, "recordings.path", "r", "audio", "Path to find recordings in")
	if markErr := decodeCmd.MarkFlagDirname("recordings.path"); markErr != nil {
		log.Fatalln("Error when marking recordings directory as only accepting dir names", markErr)
	}

	decodeCmd.Flags().StringVarP(&outputFormat, "output.format", "f", "csv", `What format to output results in. Valid options are "csv" or "json"`)
	decodeCmd.Flags().StringVarP(&outputFileName, "output.file", "o", "recordings.csv", "Path to store output in")
	if markErr := decodeCmd.MarkFlagFilename("output.file", "csv", "json"); markErr != nil {
		log.Fatalln("Error when marking output file as only accepting certain extensions", markErr)
	}
}

func findWAVs(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) != ".wav" || info.IsDir() {
			return nil
		}
		*files = append(*files, path)
		return nil
	}
}

func save(in interface{}, outputFile *os.File) error {
	var marshalled []byte
	var marshalErr error
	if outputFormat == "csv" {
		marshalled, marshalErr = gocsv.MarshalBytes(in)
	} else if outputFormat == "json" {
		if jsonIndent == "" {
			marshalled, marshalErr = json.Marshal(in)
		} else {
			marshalled, marshalErr = json.MarshalIndent(in, "", jsonIndent)
		}
	}

	if marshalErr != nil {
		return marshalErr
	}

	_, saveErr := outputFile.Write(marshalled)
	return saveErr
}
