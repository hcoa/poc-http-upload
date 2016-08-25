package main

import (
	// "errors"
	"bytes"
	"errors"
	"io"
	"log"

	"github.com/gorilla/mux"
	// "mime/multipart"
	"net/http"
	"os"
	"strconv"
)

const (
	MAX32K    = 32 << 20 // 32 Mb
	filesDir  = "./files"
	filesMask = 0644 // -rw-r--r--
	CnListen  = "LISTEN"
	APIVer    = "/v1"
	StorFile  = "stor.json"
)

// type Config map[string]string

// // simple configuration
// func (c Config) init() {
// 	c[CnListen] = Getenv(CnListen, ":8080")
// }

var storage *Store

func init() {
	storage = newStorage(StorFile)

	// try to load data from disk
	err := storage.loadData()
	if err != nil {
		log.Printf("Nothing to load")
	}
}

func uploadHandle(w http.ResponseWriter, r *http.Request) {
	var status int
	var err error
	defer func() {
		if err != nil {
			log.Printf("Error %v", err)
			http.Error(w, err.Error(), status)
		}
	}()

	err = r.ParseMultipartForm(MAX32K)
	file, header, err := r.FormFile("upfile")
	if err != nil {
		status = http.StatusInternalServerError
		return
	}
	defer file.Close()

	buf := bytes.NewBuffer(nil)
	// buf.ReadFrom(file)
	_, err = io.Copy(buf, file)
	if err != nil {
		status = http.StatusInternalServerError
		return
	}

	sf := FileHash{
		MD5Hash:  getMD5Hash(buf.Bytes()),
		MurHash:  getMurMurHash(buf.Bytes()),
		FarmHash: getFarmHash(buf.Bytes()),
	}
	dupPath := storage.findDup(&sf)
	if dupPath != "" {
		storage.addFile(header.Filename, dupPath)
		w.Write([]byte("File " + header.Filename + " was uploaded."))
		return
	}
	fullpath := filesDir + "/" + header.Filename
	outFile, err := os.OpenFile(fullpath, os.O_WRONLY|os.O_CREATE, filesMask)
	if err != nil {
		status = http.StatusInternalServerError
		return
	}
	defer outFile.Close()
	written, err := io.Copy(outFile, buf)
	if err != nil {
		status = http.StatusInternalServerError
		return
	}
	storage.addHashes(fullpath, sf)
	storage.addFile(header.Filename, fullpath)

	// save data on the disk
	go storage.saveData()
	w.Write([]byte("File " + header.Filename + " was uploaded. Length: " + strconv.Itoa(int(written))))
}

func getFileHandle(w http.ResponseWriter, r *http.Request) {
	var status int
	var err error
	defer func() {
		if err != nil {
			log.Printf("Error %v", err)
			http.Error(w, err.Error(), status)
		}
	}()

	name, ok := mux.Vars(r)["name"]
	if !ok {
		status = http.StatusBadRequest
		err = errors.New("No name were provided")
		return
	}
	filePath, ok := storage.getFilePath(name)
	if !ok {
		status = http.StatusOK
		w.Write([]byte("File " + name + " was not found"))
		return
	}
	f, err := os.Open(filePath)
	if err != nil {
		status = http.StatusInternalServerError
		return
	}
	fi, err := f.Stat()
	if err != nil {
		status = http.StatusInternalServerError
		return
	}
	defer f.Close()
	// need only 512 bytes to determine contentType
	buffer := make([]byte, 512)
	_, err = f.Read(buffer)
	f.Seek(0, 0) // reset read pointer
	contentType := http.DetectContentType(buffer)
	length := strconv.FormatInt(fi.Size(), 10)
	w.Header().Set("Content-Disposition", "attachement; filename="+name)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", length)
	_, err = io.Copy(w, f)
	if err != nil {
		status = http.StatusInternalServerError
	}
}

func deleteHandle(w http.ResponseWriter, r *http.Request) {
	var status int
	var err error
	defer func() {
		if err != nil {
			log.Printf("Error %v", err)
			http.Error(w, err.Error(), status)
		}
	}()
	name, ok := mux.Vars(r)["name"]
	if !ok {
		status = http.StatusBadRequest
		err = errors.New("No name were provided")
		return
	}
	storage.deleteFile(name)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/v1/upload", uploadHandle).Methods("POST")
	r.HandleFunc("/v1/files/{name}", getFileHandle).Methods("GET")
	r.HandleFunc("/v1/files/{name}", deleteHandle).Methods("DELETE")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
