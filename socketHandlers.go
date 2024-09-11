package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"song-recognition/db"
	"song-recognition/shazam"
	"song-recognition/spotify"
	"song-recognition/utils"
	"song-recognition/wav"

	socketio "github.com/googollee/go-socket.io"
	"github.com/mdobak/go-xerrors"
)

func status(statusType, message string) string {
	data := map[string]interface{}{"type": statusType, "message": message}
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger := utils.GetLogger()
		ctx := context.Background()
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "failed to marshal data", slog.Any("error", err))
		return ""
	}
	return string(jsonData)
}

func handleTotalSongs(socket socketio.Conn) {
	logger := utils.GetLogger()
	ctx := context.Background()

	db, err := db.NewDBClient()
	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "error connecting to DB", slog.Any("error", err))
		return
	}
	defer db.Close()

	totalSongs, err := db.TotalSongs()
	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "Log error getting total songs", slog.Any("error", err))
		return
	}

	socket.Emit("totalSongs", totalSongs)
}

func handleSave(socket socketio.Conn, songPath string, title string, artist string, pzID string) {
	logger := utils.GetLogger()
	ctx := context.Background()

	// check if track already exist
	db, err := db.NewDBClient()
	if err != nil {
		logger.ErrorContext(ctx, "error connecting to DB", slog.Any("error", err))
	}
	defer db.Close()

	song, songExists, err := db.GetSongByKey(utils.GenerateSongKey(title, artist))
	if err == nil {
		if songExists {
			statusMsg := fmt.Sprintf(
				"'%s' by '%s' already exists in the database (https://www.phi.zone/songs/%s)",
				song.Title, song.Artist, song.SongID)

			socket.Emit("saveStatus", status("error", statusMsg))
			return
		}
	} else {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "failed to get song by key", slog.Any("error", err))
	}

	err = spotify.ProcessAndSaveSong(songPath, title, artist, pzID)
	if err != nil {
		socket.Emit("saveStatus", status("error", err.Error()))
		logger.Info(err.Error())
		return
	}

	statusMsg := ""
	statusMsg = fmt.Sprintf("'%s' by '%s' was saved", title, artist)
	socket.Emit("saveStatus", status("success", statusMsg))
}

func handleFind(socket socketio.Conn, songFilePath string) {
	logger := utils.GetLogger()
	wavFilePath, err := wav.ConvertToWAV(songFilePath, 1)
	if err != nil {
		socket.Emit("findStatus", status("error", err.Error()))
		logger.Info(err.Error())
		return
	}

	wavInfo, err := wav.ReadWavInfo(wavFilePath)
	if err != nil {
		socket.Emit("findStatus", status("error", err.Error()))
		logger.Info(err.Error())
		return
	}

	samples, err := wav.WavBytesToSamples(wavInfo.Data)
	if err != nil {
		socket.Emit("findStatus", status("error", err.Error()))
		logger.Info(err.Error())
		return
	}

	matches, searchDuration, err := shazam.FindMatches(samples, wavInfo.Duration, wavInfo.SampleRate)
	if err != nil {
		socket.Emit("findStatus", status("error", err.Error()))
		logger.Info(err.Error())
		return
	}

	var simplifiedMatches []map[string]interface{}
	for _, match := range matches {
		simplifiedMatches = append(simplifiedMatches, map[string]interface{}{
			"id":        match.PhiZoneID,
			"timestamp": match.Timestamp,
			"score":     match.Score,
		})
	}

	socket.Emit("findResult", map[string]interface{}{
		"matches":  simplifiedMatches,
		"timeTook": searchDuration.Seconds(),
	})

	logger.Info("Find matches emitted successfully")
}
