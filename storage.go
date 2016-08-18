package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
)

type SaveLoader interface {
	saveData() error
	loadData() error
}

type Storer interface {
	SaveLoader
	addFile(name, path string)
	getFilePath(name string) (string, bool)
	deleteFile(name string) error
}

type FileHash struct {
	MD5Hash  [md5.Size]byte `json:"md5,omitempty"`
	MurHash  []byte         `json:"mur,omitempty"`
	FarmHash uint32         `json:"farm,omitempty"`
}

type Store struct {
	Files       map[string]string   `json:"files,omitepmty"`  // will contain name -> path
	Hashes      map[string]FileHash `json:"hashes,omitempty"` // path -> hashes
	FilePath    string              `json:"-"`
	*sync.Mutex `json:"-"`
}

func newStorage(filePath string) *Store {
	return &Store{
		Files:    make(map[string]string),
		Hashes:   make(map[string]FileHash),
		FilePath: filePath,
		Mutex:    new(sync.Mutex),
	}
}

func (s *Store) saveData() error {
	buffer := new(bytes.Buffer)
	s.Lock()
	defer s.Unlock()
	err := json.NewEncoder(buffer).Encode(s)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(s.FilePath, buffer.Bytes(), filesMask)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) loadData() error {
	loadedStore := &Store{}
	s.Lock()
	defer s.Unlock()
	f, err := os.Open(s.FilePath)
	if err != nil {
		return err
	}
	// buf := bufio.NewReader(f)
	err = json.NewDecoder(f).Decode(loadedStore)
	if err != nil {
		return err
	}
	s.Files = loadedStore.Files
	return nil
}

func (s *Store) addFile(name, path string) {
	s.Lock()
	s.Files[name] = path
	s.Unlock()
}

func (s *Store) getFilePath(name string) (string, bool) {
	s.Lock()
	defer s.Unlock()
	v, k := s.Files[name]
	return v, k
}

func (s *Store) deleteFile(name string) error {
	return nil
}

func (s *Store) addHashes(path string, sf FileHash) {
	s.Lock()
	s.Hashes[path] = sf
	s.Unlock()
}

// not very efficient search. it takes O(n)
func (s *Store) findDup(sf *FileHash) string {
	s.Lock()
	defer s.Unlock()
	for k, v := range s.Hashes {
		score := 0
		if bytes.Compare(v.MurHash, sf.MurHash) == 0 {
			score++
		}
		if v.FarmHash == sf.FarmHash {
			score++
		}
		if v.MD5Hash == sf.MD5Hash {
			score++
		}

		if score >= 2 {
			return k
		}
	}
	return ""
}
