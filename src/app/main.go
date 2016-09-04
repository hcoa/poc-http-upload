package main

import (
	// "errors"

	"flag"
	"log"
	"os/signal"
	"syscall"

	// "mime/multipart"

	"os"
)

const (
	MAX32K    = 32 << 20 // 32 Mb
	filesMask = 0644     // -rw-r--r--
	// APIVer    = "/v1"
	StorFile     = "stor.json"
	StaticFolder = "./static/"
)

var (
	filesDir   = flag.String("dir", "./files", "directory to store files")
	listenAddr = flag.String("addr", ":8080", "listen addr. default is :8080")
)

// type Config map[string]string

// // simple configuration
// func (c Config) init() {
// 	c[CnListen] = Getenv(CnListen, ":8080")
// }

var storage Storer

func init() {
	storage = newStorage(StorFile)

	// try to load data from disk
	err := storage.loadData()
	if err != nil {
		log.Printf("WARN: Nothing to load")
	}
}

func main() {
	flag.Parse()
	server := NewServer(ServOptions{
		FilesDir: *filesDir,
		Listen:   *listenAddr,
	})
	signalCh := make(chan os.Signal, syscall.SIGINT)
	signal.Notify(signalCh)
	errCh := make(chan error)
	go func() {
		if err := server.Start(); err != nil {
			errCh <- err
		}
	}()
	for {
		select {
		case err := <-errCh:
			log.Printf("ERR: %s", err)
			os.Exit(1)
		case sig := <-signalCh:
			switch sig {
			case syscall.SIGINT:
				server.Stop()
				os.Exit(4)
			}
		}
	}
}
