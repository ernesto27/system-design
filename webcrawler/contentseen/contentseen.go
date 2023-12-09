package contentseen

import (
	"crypto/sha256"
	"encoding/hex"
)

type ContentSeen struct {
	html string
}

func New(html string) *ContentSeen {
	return &ContentSeen{html: html}
}

func (c *ContentSeen) CreateHash() string {
	hasher := sha256.New()
	hasher.Write([]byte(c.html))
	hashSum := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashSum)

	return hashString
}
