package utils

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"
)

func GenerateUniqueID() uint32 {
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Uint32()

	return randomNumber
}

func GenerateSongKey(songTitle, songArtist string) string {
	return songTitle + "---" + songArtist
}

func GetEnv(key string, fallback ...string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	if len(fallback) > 0 {
		return fallback[0]
	}
	return ""
}

func DownloadFile(fileURL, destDir string) (string, error) {
	// 解析URL
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return "", err
	}

	// 从URL路径中提取文件名
	fileName := path.Base(parsedURL.Path)
	// 先检查目标目录是否存在
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		// 不存在则创建目录
		err = os.MkdirAll(destDir, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	// 创建目标文件
	destPath := path.Join(destDir, fileName)
	out, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// 发送HTTP GET请求
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 检查HTTP响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	// 将响应的内容写入文件
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	// 返回文件路径
	return path.Join(destDir, fileName), nil
}
