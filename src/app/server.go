package main

import (
	"net/http"
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
