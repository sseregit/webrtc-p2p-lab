package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const uploadDir = "./uploads"

func main() {
	err := os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		http.ServeFile(writer, request, "./static/index.html")
	})

	http.HandleFunc("/upload", func(writer http.ResponseWriter, request *http.Request) {
		file, handler, err := request.FormFile("file")
		if err != nil {
			http.Error(writer, "Failed to create file", http.StatusBadRequest)
			return
		}

		defer file.Close()

		fileName := request.FormValue("fileName")

		if fileName == "" {
			fileName = handler.Filename
		}

		destPath := filepath.Join(uploadDir, fileName)
		destFile, err := os.Create(destPath)
		if err != nil {
			http.Error(writer, "Failed to create destination", http.StatusInternalServerError)
			return
		}
		defer destFile.Close()

		buffer := make([]byte, 1024*1024)

		for {
			n, err := file.Read(buffer)
			if err != nil {
				http.Error(writer, "Failed to read destination", http.StatusInternalServerError)
				return
			}

			if n == 0 {
				break
			}

			_, err = destFile.Write(buffer[:n])
			if err != nil {
				http.Error(writer, "Failed to write destination", http.StatusInternalServerError)
				return
			}
		}

		fmt.Println("Success to create destination")

	})

	http.HandleFunc("/video", func(writer http.ResponseWriter, request *http.Request) {
		fileName := request.URL.Query().Get("fileName")
		if fileName == "" {
			http.Error(writer, "Invalid file name", http.StatusBadRequest)
		}

		filepath := filepath.Join(uploadDir, fileName)
		file, err := os.Open(filepath)
		if err != nil {
			http.Error(writer, "Invalid file", http.StatusInternalServerError)
			return
		}

		defer file.Close()

		rangeHeader := request.Header.Get("Range")
		if rangeHeader == "" {
			http.ServeFile(writer, request, filepath)
			return
		}

		// Range : byte=313072-621321
		// byte = <start> - <end>
		rangePars := strings.Split(strings.TrimPrefix(rangeHeader, "bytes="), "-")
		start, _ := strconv.ParseInt(rangePars[0], 10, 64)

		stat, _ := file.Stat()
		fileSize := stat.Size()

		var end int64

		if len(rangePars) > 1 && rangePars[1] != "" {
			end, _ = strconv.ParseInt(rangePars[1], 10, 64)
		} else {
			end = fileSize - 1
		}

		if end >= fileSize {
			end = fileSize - 1
		}

		writer.Header().Set("Content-Type", "video/mp4")
		writer.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
		writer.Header().Set("Accept-Ranges", "bytes")

		file.Seek(start, io.SeekStart)
		io.CopyN(writer, file, end-start+1)

	})
}
