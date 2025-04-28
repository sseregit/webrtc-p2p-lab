package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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

	})
}
