package sync

import (
	"crypto/rand"
	"fmt"
)

func sessionID() string {
	buf := make([]byte, 16)
	rand.Read(buf)
	return fmt.Sprintf("%x", buf)
}
