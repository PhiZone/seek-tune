package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"song-recognition/db"
	"song-recognition/shazam"
	"song-recognition/spotify"
	"song-recognition/utils"
	"song-recognition/wav"

	"github.com/mdobak/go-xerrors"
)

// 歌曲保存
func handleHttpSave(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()
	ctx := context.Background()
	// 检查请求方法是否为POST
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed) // 状态码: 405
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest) // 状态码: 400
		return
	}

	songPath := r.FormValue("songPath") // 本地文件路径
	songURL := r.FormValue("songUrl")   // 远程文件URL
	title := r.FormValue("title")       // 歌曲标题
	artist := r.FormValue("artist")     // 曲师
	uuid := r.FormValue("uuid")         // UUID
	// 检查所有字段是否都有值，songPath和songURL至少有一个
	if (songPath == "" && songURL == "") || (songPath != "" && songURL != "") {
		http.Error(w, "Either songPath or songUrl must be provided, but not both", http.StatusBadRequest) // 状态码: 400
		return
	}
	// 其他字段不能为空
	if title == "" || artist == "" || uuid == "" {
		http.Error(w, "Missing required field: title, artist, or uuid", http.StatusBadRequest) // 状态码: 400
		return
	}
	// 如果是URL，使用末尾的文件名作为音乐文件名，缓存到./urlSongTemp目录下
	if songURL != "" {
		//使用GET请求下载文件
		filePath, err := utils.DownloadFile(songURL, "./urlSongTemp", uuid)
		if err != nil {
			// 如果下载失败，返回错误信息
			http.Error(w, err.Error(), http.StatusInternalServerError) // 状态码: 500
			return
		}
		songPath = filePath
	}

	database, err := db.NewDBClient()
	if err != nil {
		logger.ErrorContext(ctx, "error connecting to DB", slog.Any("error", err))
		http.Error(w, "Error connecting to DB", http.StatusInternalServerError) // 状态码: 500
		return
	}
	defer database.Close()

	_, songExists, err := database.GetSongByUUID(uuid) //GetSongByKey(utils.GenerateSongKey(title, artist))
	if err == nil {
		if songExists {
			// 这个ID已经存在于数据库中
			statusMsg := fmt.Sprintf("UUID: %s already exists in the database", uuid)
			http.Error(w, statusMsg, http.StatusConflict) // 状态码: 409
			return
		}
	} else {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "failed to get song by key", slog.Any("error", err)) // 通过key获取歌曲失败
		http.Error(w, "Failed to get song by key", http.StatusInternalServerError)    // 状态码: 500
		return
	}

	err = spotify.ProcessAndSaveSong(songPath, title, artist, uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // 状态码: 500
		logger.Info(err.Error())
		return
	}

	statusMsg := fmt.Sprintf("'%s' by '%s' was saved", title, artist) // 歌曲已保存
	w.WriteHeader(http.StatusOK)                                      // 状态码: 200
	wr, err := w.Write([]byte(statusMsg))
	if err != nil {
		logger.ErrorContext(ctx, "Failed to write response", slog.Any("error", err))
		http.Error(w, "Failed to write response", http.StatusInternalServerError) // 状态码: 500
		return
	}
	logger.Info("HTTP save response written successfully", slog.Int("bytesWritten", wr))
	// 如果songURL不为空，删除songFilePath指向的文件
	if songURL != "" {
		if err := os.Remove(songPath); err != nil {
			logger.ErrorContext(ctx, "Failed to delete file", slog.Any("error", err))
		}
	}
}

// 歌曲总数
func handleHttpTotalSongs(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()
	ctx := context.Background()
	// 检查请求方法是否为GET
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	db, err := db.NewDBClient()
	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "error connecting to DB", slog.Any("error", err))
		http.Error(w, "Error connecting to DB", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	totalSongs, err := db.TotalSongs()
	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "Log error getting total songs", slog.Any("error", err))
		http.Error(w, "Error getting total songs", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	wr, err := w.Write([]byte(fmt.Sprintf("Total songs: %d", totalSongs)))
	if err != nil {
		logger.ErrorContext(ctx, "Failed to write response", slog.Any("error", err))
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
	logger.Info("HTTP total songs response written successfully", slog.Int("bytesWritten", wr))
}

// 意义不明
func handleHttpSongsUnsaved(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()
	ctx := context.Background()

	var requestedIDs []string
	if err := json.NewDecoder(r.Body).Decode(&requestedIDs); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		logger.ErrorContext(ctx, "Invalid request payload", slog.Any("error", err))
		return
	}

	db, err := db.NewDBClient()
	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "error connecting to DB", slog.Any("error", err))
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	nonExistentIDs, err := db.FindNonExistentSongs(requestedIDs)
	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "error checking for non-existent songs", slog.Any("error", err))
		http.Error(w, "Error checking for non-existent songs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(nonExistentIDs); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		logger.ErrorContext(ctx, "Error encoding response", slog.Any("error", err))
	}
}

// 歌曲查找
func handleHttpFind(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()
	ctx := context.Background()

	// 检查请求方法是否为GET
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// 从请求参数中获取songFilePath和songUrl
	songPath := r.URL.Query().Get("songPath")
	songURL := r.URL.Query().Get("songUrl")

	if (songPath == "" && songURL == "") || (songPath != "" && songURL != "") {
		http.Error(w, "Either songFilePath or songUrl must be provided, but not both", http.StatusBadRequest) // 状态码: 400
		return
	}
	// 如果是URL，使用末尾的文件名作为音乐文件名，缓存到./urlSongTemp目录下
	if songURL != "" {
		filePath, err := utils.DownloadFile(songURL, "./urlSongTemp", utils.GenerateUniqueID())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Info(err.Error())
			return
		}
		songPath = filePath
	}

	wavFilePath, err := wav.ConvertToWAV(songPath, 1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Info(err.Error())
		return
	}

	wavInfo, err := wav.ReadWavInfo(wavFilePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Info(err.Error())
		return
	}

	samples, err := wav.WavBytesToSamples(wavInfo.Data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Info(err.Error())
		return
	}

	matches, _, err := shazam.FindMatches(samples, wavInfo.Duration, wavInfo.SampleRate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Info(err.Error())
		return
	}

	var simplifiedMatches []map[string]interface{}
	for _, match := range matches {
		simplifiedMatches = append(simplifiedMatches, map[string]interface{}{
			"id":    match.UUID,
			"score": match.Score,
		})
	}
	// 如果是空的，返回状态码404
	if len(simplifiedMatches) == 0 {
		http.Error(w, "No match found", http.StatusNotFound)
		return
	}
	response := simplifiedMatches

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		logger.ErrorContext(ctx, "Error encoding response", slog.Any("error", err))
	}
	// 如果songURL不为空，删除songFilePath指向的文件
	if songURL != "" {
		if err := os.Remove(songPath); err != nil {
			logger.ErrorContext(ctx, "Failed to delete file", slog.Any("error", err))
		}
	}
}

// 通过UUID查找歌曲是否存在
func handleHttpSongExists(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()
	ctx := context.Background()

	// 检查请求方法是否为GET
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		http.Error(w, "Missing required field: uuid", http.StatusBadRequest)
		return
	}

	db, err := db.NewDBClient()
	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "error connecting to DB", slog.Any("error", err))
		http.Error(w, "Error connecting to DB", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	exists, err := db.SongExistsByID(uuid)
	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "error checking song existence", slog.Any("error", err))
		http.Error(w, "Error checking song existence", http.StatusInternalServerError)
		return
	}
	// 如果exits为true，返回状态码200，否则返回状态码404
	if exists {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}

	// 空响应体
	wr, _ := w.Write([]byte(""))
	logger.Info("HTTP song exists response written successfully", slog.Int("bytesWritten", wr))
}

// 更新UUID对应的歌曲信息
func handleHttpUpdate(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()
	ctx := context.Background()
	// 检查请求方法是否为PATCH
	if r.Method != http.MethodPatch {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed) // 状态码: 405
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest) // 状态码: 400
		return
	}

	songPath := r.FormValue("songPath") // 本地文件路径
	songURL := r.FormValue("songUrl")   // 远程文件URL
	title := r.FormValue("title")       // 歌曲标题
	artist := r.FormValue("artist")     // 曲师
	uuid := r.FormValue("uuid")         // UUID
	// 检查所有字段是否都有值，songPath和songURL至少有一个
	if (songPath == "" && songURL == "") || (songPath != "" && songURL != "") {
		http.Error(w, "Either songPath or songUrl must be provided, but not both", http.StatusBadRequest) // 状态码: 400
		return
	}
	// 其他字段不能为空
	if title == "" || artist == "" || uuid == "" {
		http.Error(w, "Missing required field: title, artist, or uuid", http.StatusBadRequest) // 状态码: 400
		return
	}

	database, err := db.NewDBClient()
	if err != nil {
		logger.ErrorContext(ctx, "error connecting to DB", slog.Any("error", err))
		http.Error(w, "Error connecting to DB", http.StatusInternalServerError) // 状态码: 500
		return
	}
	defer database.Close()

	_, songExists, err := database.GetSongByUUID(uuid)
	if err == nil {
		if !songExists {
			statusMsg := fmt.Sprintf("No data found for uuid: %s", uuid)
			http.Error(w, statusMsg, http.StatusNotFound) // 状态码: 404
			return
		}
	} else {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "failed to get song by key", slog.Any("error", err)) // 通过key获取歌曲失败
		http.Error(w, "Failed to get song by key", http.StatusInternalServerError)    // 状态码: 500
		return
	}

	// 如果是URL，使用末尾的文件名作为音乐文件名，缓存到./urlSongTemp目录下
	if songURL != "" {
		//使用GET请求下载文件
		filePath, err := utils.DownloadFile(songURL, "./urlSongTemp", uuid)
		if err != nil {
			// 如果下载失败，返回错误信息
			http.Error(w, err.Error(), http.StatusInternalServerError) // 状态码: 500
			return
		}
		songPath = filePath
	}

	err = spotify.ProcessAndUpdateSong(songPath, title, artist, uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // 状态码: 500
		logger.Info(err.Error())
		return
	}

	statusMsg := fmt.Sprintf("'%s' has been updated", uuid) // 歌曲已保存
	w.WriteHeader(http.StatusOK)                            // 状态码: 200
	wr, err := w.Write([]byte(statusMsg))
	if err != nil {
		logger.ErrorContext(ctx, "Failed to write response", slog.Any("error", err))
		http.Error(w, "Failed to write response", http.StatusInternalServerError) // 状态码: 500
		return
	}
	logger.Info("HTTP save response written successfully", slog.Int("bytesWritten", wr))
	// 如果songURL不为空，删除songFilePath指向的文件
	if songURL != "" {
		if err := os.Remove(songPath); err != nil {
			logger.ErrorContext(ctx, "Failed to delete file", slog.Any("error", err))
		}
	}
}

// 删除歌曲
func handleHttpDelete(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()
	ctx := context.Background()
	// 检查请求方法是否为DELETE
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed) // 状态码: 405
		return
	}

	// 从请求参数中获取uuid
	uuid := r.FormValue("uuid")

	database, err := db.NewDBClient()
	if err != nil {
		logger.ErrorContext(ctx, "error connecting to DB", slog.Any("error", err))
		http.Error(w, "Error connecting to DB", http.StatusInternalServerError) // 状态码: 500
		return
	}
	defer database.Close()

	// 先检查歌曲是否存在
	_, songExists, err := database.GetSongByUUID(uuid)
	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "error checking song existence", slog.Any("error", err))
		http.Error(w, "Error checking song existence", http.StatusInternalServerError)
		return
	}
	if !songExists {
		statusMsg := fmt.Sprintf("No data found for uuid: %s", uuid)
		http.Error(w, statusMsg, http.StatusNotFound) // 状态码: 404
		return
	}

	err = database.DeleteSongByUUID(uuid)
	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "error deleting song", slog.Any("error", err))
		http.Error(w, "Error deleting song", http.StatusInternalServerError)
		return
	}

	statusMsg := fmt.Sprintf("'%s' was deleted", uuid) // 歌曲已删除
	w.WriteHeader(http.StatusOK)                       // 状态码: 200
	wr, err := w.Write([]byte(statusMsg))
	if err != nil {
		logger.ErrorContext(ctx, "Failed to write response", slog.Any("error", err))
		http.Error(w, "Failed to write response", http.StatusInternalServerError) // 状态码: 500
		return
	}
	logger.Info("HTTP delete response written successfully", slog.Int("bytesWritten", wr))
}

// 版权音乐部分
func handleHttpCopyrightSave(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()
	ctx := context.Background()
	// 检查请求方法是否为POST
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed) // 状态码: 405
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest) // 状态码: 400
		return
	}

	songPath := r.FormValue("songPath") // 本地文件路径
	songURL := r.FormValue("songUrl")   // 远程文件URL
	title := r.FormValue("title")       // 歌曲标题
	artist := r.FormValue("artist")     // 曲师
	uuid := r.FormValue("uuid")         // UUID
	// 检查所有字段是否都有值，songPath和songURL至少有一个
	if (songPath == "" && songURL == "") || (songPath != "" && songURL != "") {
		http.Error(w, "Either songPath or songUrl must be provided, but not both", http.StatusBadRequest) // 状态码: 400
		return
	}
	// 其他字段不能为空
	if title == "" || artist == "" || uuid == "" {
		http.Error(w, "Missing required field: title, artist, or uuid", http.StatusBadRequest) // 状态码: 400
		return
	}
	// 如果是URL，使用末尾的文件名作为音乐文件名，缓存到./urlSongTemp目录下
	if songURL != "" {
		//使用GET请求下载文件
		filePath, err := utils.DownloadFile(songURL, "./urlSongTemp", uuid)
		if err != nil {
			// 如果下载失败，返回错误信息
			http.Error(w, err.Error(), http.StatusInternalServerError) // 状态码: 500
			return
		}
		songPath = filePath
	}

	database, err := db.NewDBClient()
	if err != nil {
		logger.ErrorContext(ctx, "error connecting to DB", slog.Any("error", err))
		http.Error(w, "Error connecting to DB", http.StatusInternalServerError) // 状态码: 500
		return
	}
	defer database.Close()

	song, songExists, err := database.GetCopyrightSongByKey(utils.GenerateSongKey(title, artist))
	if err == nil {
		if songExists {
			statusMsg := fmt.Sprintf(
				"'%s' by '%s' already exists in the database (https://www.phi.zone/songs/%s)",
				song.Title, song.Artist, song.SongID) // Artist的Title已经存在于数据库中（https://www.phi.zone/songs/ID）
			http.Error(w, statusMsg, http.StatusConflict) // 状态码: 409
			return
		}
	} else {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "failed to get song by key", slog.Any("error", err)) // 通过key获取歌曲失败
		http.Error(w, "Failed to get song by key", http.StatusInternalServerError)    // 状态码: 500
		return
	}

	err = spotify.ProcessAndSaveCopyrightSong(songPath, title, artist, uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // 状态码: 500
		logger.Info(err.Error())
		return
	}

	statusMsg := fmt.Sprintf("'%s' by '%s' was saved", title, artist) // 歌曲已保存
	w.WriteHeader(http.StatusOK)                                      // 状态码: 200
	wr, err := w.Write([]byte(statusMsg))
	if err != nil {
		logger.ErrorContext(ctx, "Failed to write response", slog.Any("error", err))
		http.Error(w, "Failed to write response", http.StatusInternalServerError) // 状态码: 500
		return
	}
	logger.Info("HTTP save response written successfully", slog.Int("bytesWritten", wr))
	// 如果songURL不为空，删除songFilePath指向的文件
	if songURL != "" {
		if err := os.Remove(songPath); err != nil {
			logger.ErrorContext(ctx, "Failed to delete file", slog.Any("error", err))
		}
	}
}

func handleHttpCopyrightTotalSongs(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()
	ctx := context.Background()
	// 检查请求方法是否为GET
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	db, err := db.NewDBClient()
	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "error connecting to DB", slog.Any("error", err))
		http.Error(w, "Error connecting to DB", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	totalSongs, err := db.TotalCopyrightSongs()
	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "Log error getting total songs", slog.Any("error", err))
		http.Error(w, "Error getting total songs", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	wr, err := w.Write([]byte(fmt.Sprintf("Total songs: %d", totalSongs)))
	if err != nil {
		logger.ErrorContext(ctx, "Failed to write response", slog.Any("error", err))
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
	logger.Info("HTTP total songs response written successfully", slog.Int("bytesWritten", wr))
}

func handleHttpCopyrightFind(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()
	ctx := context.Background()

	// 检查请求方法是否为GET
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// 从请求参数中获取songFilePath和songUrl
	songPath := r.URL.Query().Get("songPath")
	songURL := r.URL.Query().Get("songUrl")

	if (songPath == "" && songURL == "") || (songPath != "" && songURL != "") {
		http.Error(w, "Either songFilePath or songUrl must be provided, but not both", http.StatusBadRequest) // 状态码: 400
		return
	}
	// 如果是URL，使用末尾的文件名作为音乐文件名，缓存到./urlSongTemp目录下
	if songURL != "" {
		filePath, err := utils.DownloadFile(songURL, "./urlSongTemp", utils.GenerateUniqueID())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Info(err.Error())
			return
		}
		songPath = filePath
	}

	wavFilePath, err := wav.ConvertToWAV(songPath, 1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Info(err.Error())
		return
	}

	wavInfo, err := wav.ReadWavInfo(wavFilePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Info(err.Error())
		return
	}

	samples, err := wav.WavBytesToSamples(wavInfo.Data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Info(err.Error())
		return
	}

	matches, _, err := shazam.FindCopyrightMatches(samples, wavInfo.Duration, wavInfo.SampleRate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Info(err.Error())
		return
	}

	var simplifiedMatches []map[string]interface{}
	for _, match := range matches {
		simplifiedMatches = append(simplifiedMatches, map[string]interface{}{
			"id":    match.UUID,
			"score": match.Score,
		})
	}
	// 如果是空的，返回状态码404
	if len(simplifiedMatches) == 0 {
		http.Error(w, "No match found", http.StatusNotFound)
		return
	}
	response := simplifiedMatches

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		logger.ErrorContext(ctx, "Error encoding response", slog.Any("error", err))
	}
	// 如果songURL不为空，删除songFilePath指向的文件
	if songURL != "" {
		if err := os.Remove(songPath); err != nil {
			logger.ErrorContext(ctx, "Failed to delete file", slog.Any("error", err))
		}
	}
}

func handleHttpCopyrightExists(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()
	ctx := context.Background()

	// 检查请求方法是否为GET
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// 从请求参数中获取UUID
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		http.Error(w, "Missing required field: uuid", http.StatusBadRequest)
		return
	}

	// 连接数据库
	db, err := db.NewDBClient()
	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "error connecting to DB", slog.Any("error", err))
		http.Error(w, "Error connecting to DB", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// 检查UUID对应的歌曲是否存在
	exists, err := db.CopyrightSongExistsByID(uuid)
	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "error checking song existence", slog.Any("error", err))
		http.Error(w, "Error checking song existence", http.StatusInternalServerError)
		return
	}
	// 如果exits为true，返回状态码200，否则返回状态码404
	if exists {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}

	// 空响应体
	wr, _ := w.Write([]byte(""))
	logger.Info("HTTP song exists response written successfully", slog.Int("bytesWritten", wr))
}

func handleHttpCopyrightUpdate(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()
	ctx := context.Background()
	// 检查请求方法是否为PATCH
	if r.Method != http.MethodPatch {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed) // 状态码: 405
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest) // 状态码: 400
		return
	}

	songPath := r.FormValue("songPath") // 本地文件路径
	songURL := r.FormValue("songUrl")   // 远程文件URL
	title := r.FormValue("title")       // 歌曲标题
	artist := r.FormValue("artist")     // 曲师
	uuid := r.FormValue("uuid")         // UUID
	// 检查所有字段是否都有值，songPath和songURL至少有一个
	if (songPath == "" && songURL == "") || (songPath != "" && songURL != "") {
		http.Error(w, "Either songPath or songUrl must be provided, but not both", http.StatusBadRequest) // 状态码: 400
		return
	}
	// 其他字段不能为空
	if title == "" || artist == "" || uuid == "" {
		http.Error(w, "Missing required field: title, artist, or uuid", http.StatusBadRequest) // 状态码: 400
		return
	}

	database, err := db.NewDBClient()
	if err != nil {
		logger.ErrorContext(ctx, "error connecting to DB", slog.Any("error", err))
		http.Error(w, "Error connecting to DB", http.StatusInternalServerError) // 状态码: 500
		return
	}
	defer database.Close()

	_, songExists, err := database.GetCopyrightSongByUUID(uuid)
	if err == nil {
		if !songExists {
			statusMsg := fmt.Sprintf("No data found for uuid: %s", uuid)
			http.Error(w, statusMsg, http.StatusNotFound) // 状态码: 404
			return
		}
	} else {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "failed to get song by key", slog.Any("error", err)) // 通过key获取歌曲失败
		http.Error(w, "Failed to get song by key", http.StatusInternalServerError)    // 状态码: 500
		return
	}

	// 如果是URL，使用末尾的文件名作为音乐文件名，缓存到./urlSongTemp目录下
	if songURL != "" {
		//使用GET请求下载文件
		filePath, err := utils.DownloadFile(songURL, "./urlSongTemp", uuid)
		if err != nil {
			// 如果下载失败，返回错误信息
			http.Error(w, err.Error(), http.StatusInternalServerError) // 状态码: 500
			return
		}
		songPath = filePath
	}

	err = spotify.ProcessAndUpdateCopyrightSong(songPath, title, artist, uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // 状态码: 500
		logger.Info(err.Error())
		return
	}

	statusMsg := fmt.Sprintf("'%s' has been updated", uuid) // 歌曲已保存
	w.WriteHeader(http.StatusOK)                            // 状态码: 200
	wr, err := w.Write([]byte(statusMsg))
	if err != nil {
		logger.ErrorContext(ctx, "Failed to write response", slog.Any("error", err))
		http.Error(w, "Failed to write response", http.StatusInternalServerError) // 状态码: 500
		return
	}
	logger.Info("HTTP save response written successfully", slog.Int("bytesWritten", wr))
	// 如果songURL不为空，删除songFilePath指向的文件
	if songURL != "" {
		if err := os.Remove(songPath); err != nil {
			logger.ErrorContext(ctx, "Failed to delete file", slog.Any("error", err))
		}
	}
}

// 删除歌曲
func handleHttpCopyrightDelete(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger()
	ctx := context.Background()
	// 检查请求方法是否为DELETE
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed) // 状态码: 405
		return
	}

	// 从请求参数中获取uuid
	uuid := r.FormValue("uuid")

	database, err := db.NewDBClient()
	if err != nil {
		logger.ErrorContext(ctx, "error connecting to DB", slog.Any("error", err))
		http.Error(w, "Error connecting to DB", http.StatusInternalServerError) // 状态码: 500
		return
	}
	defer database.Close()

	// 先检查歌曲是否存在
	_, songExists, err := database.GetCopyrightSongByUUID(uuid)
	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "error checking song existence", slog.Any("error", err))
		http.Error(w, "Error checking song existence", http.StatusInternalServerError)
		return
	}
	if !songExists {
		statusMsg := fmt.Sprintf("No data found for uuid: %s", uuid)
		http.Error(w, statusMsg, http.StatusNotFound) // 状态码: 404
		return
	}

	err = database.DeleteCopyrightSongByUUID(uuid)
	if err != nil {
		err := xerrors.New(err)
		logger.ErrorContext(ctx, "error deleting song", slog.Any("error", err))
		http.Error(w, "Error deleting song", http.StatusInternalServerError)
		return
	}

	statusMsg := fmt.Sprintf("'%s' was deleted", uuid) // 歌曲已删除
	w.WriteHeader(http.StatusOK)                       // 状态码: 200
	wr, err := w.Write([]byte(statusMsg))
	if err != nil {
		logger.ErrorContext(ctx, "Failed to write response", slog.Any("error", err))
		http.Error(w, "Failed to write response", http.StatusInternalServerError) // 状态码: 500
		return
	}
	logger.Info("HTTP delete response written successfully", slog.Int("bytesWritten", wr))
}
