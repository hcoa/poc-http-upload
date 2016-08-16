package main

import (
	// "errors"
	"github.com/gorilla/mux"
	"io"
	"log"
	// "mime/multipart"
	"net/http"
	"os"
	"strconv"
)

const (
	MAX32K    = 32 << 20 // 32 Mb
	filesDir  = "./files"
	filesMask = 0644 // -rw-r--r--
)

func UploadHandle(w http.ResponseWriter, r *http.Request) {
	var status int
	var err error
	defer func() {
		if err != nil {
			http.Error(w, err.Error(), status)
		}
	}()

	r.ParseMultipartForm(MAX32K)
	file, handler, err := r.FormFile("upfile")
	if err != nil {
		status = http.StatusInternalServerError
		return
	}
	defer file.Close()
	outFile, err := os.OpenFile(filesDir+"/"+handler.Filename, os.O_WRONLY|os.O_CREATE, filesMask)
	if err != nil {
		status = http.StatusInternalServerError
		return
	}
	defer outFile.Close()
	written, err := io.Copy(outFile, file)
	if err != nil {
		status = http.StatusInternalServerError
		return
	}
	w.Write([]byte("File " + handler.Filename + " was uploaded. Length: " + strconv.Itoa(int(written))))
}

func docHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is docHandler\n"))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/v1/upload", UploadHandle)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
