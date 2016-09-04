package main

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/tylerb/graceful"
)

type Server struct {
	httpServ *graceful.Server
	stopCh   chan struct{}
}

type ServOptions struct {
	FilesDir string
	Listen   string
}

func NewServer(opts ServOptions) *Server {
	m := http.NewServeMux()
	r := mux.NewRouter()
	r.HandleFunc("/api/upload", uploadHandle).Methods("POST")
	r.HandleFunc("/api/files/{name}", getFileHandle).Methods("GET")
	r.HandleFunc("/api/files/{name}", deleteHandle).Methods("DELETE")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(StaticFolder)))
	m.Handle("/", r)

	server := &Server{
		stopCh: make(chan struct{}, 1),
	}
	server.httpServ = &graceful.Server{
		Timeout: 360 * time.Second,
		Server: &http.Server{
			Addr:    opts.Listen,
			Handler: m,
		},
	}
	return server
}

func (s *Server) Start() error {
	if err := s.httpServ.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func (s *Server) Stop() {
	close(s.stopCh)
}

func uploadHandle(w http.ResponseWriter, r *http.Request) {
	var status int
	var err error
	defer func() {
		if err != nil {
			log.Printf("ERR: %v", err)
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
	fullpath := *filesDir + "/" + header.Filename
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
			log.Printf("ERR: %v", err)
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
			log.Printf("ERR: %v", err)
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
