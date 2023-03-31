package utils

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

func HashSha256(s string) string {
	s = strings.ToLower(s)
	h := sha256.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func HashSha256Match(clear, hash string, caseMatch bool) bool {
	if !caseMatch {
		return strings.ToLower(hash) == strings.ToLower(HashSha256(clear))
	}
	return hash == HashSha256(clear)
}
