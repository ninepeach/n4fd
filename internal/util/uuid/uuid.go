package uuid

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
)

type UUID [16]byte

func Parse(s string) (UUID, error) {
	var u UUID
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "-", "")
	if len(s) != 32 {
		return u, errors.New("invalid UUID length")
	}
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 16 {
		return u, errors.New("invalid UUID hex")
	}
	copy(u[:], b)
	return u, nil
}

func MustParse(s string) UUID {
	u, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

func New() UUID {
	var u UUID
	_, _ = rand.Read(u[:])
	return u
}
