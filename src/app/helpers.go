package main

import (
	"crypto/md5"

	farm "github.com/dgryski/go-farm"
	"github.com/zond/god/murmur"
)

func getMurMurHash(f []byte) []byte {
	return murmur.HashBytes(f)
}

func getMD5Hash(f []byte) [md5.Size]byte {
	return md5.Sum(f)
}

func getFarmHash(f []byte) uint32 {
	return farm.Hash32(f)
}
