package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"song-recognition/utils"

	"github.com/mdobak/go-xerrors"
)

func main() {
	err := utils.CreateFolder("tmp")
	if err != nil {
		logger := utils.GetLogger()
		err := xerrors.New(err)
		ctx := context.Background()
		logger.ErrorContext(ctx, "failed to create tmp dir", slog.Any("error", err))
	}

	err = utils.CreateFolder(SONGS_DIR)
	if err != nil {
		err := xerrors.New(err)
		logger := utils.GetLogger()
		ctx := context.Background()
		logMsg := fmt.Sprintf("failed to create directory %v", SONGS_DIR)
		logger.ErrorContext(ctx, logMsg, slog.Any("error", err))
	}

	if len(os.Args) < 2 {
		fmt.Println("Expected 'find', 'erase', or 'serve' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "find":
		if len(os.Args) < 3 {
			fmt.Println("Usage: main.go find <path_to_wav_file>")
			os.Exit(1)
		}
		filePath := os.Args[2]
		find(filePath)
	case "serve":
		serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
		protocol := serveCmd.String("proto", "http", "Protocol to use (http or https)")
		port := serveCmd.String("p", "5000", "Port to use")
		serveCmd.Parse(os.Args[2:])
		serve(*protocol, *port)
	case "erase":
		erase(SONGS_DIR)
	default:
		fmt.Println("Expected 'find', 'erase', or 'serve' subcommands")
		os.Exit(1)
	}
}
