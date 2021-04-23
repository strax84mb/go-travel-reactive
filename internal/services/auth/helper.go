package auth

import (
	"crypto/sha512"
	"encoding/hex"
	"math/rand"
	"time"

	"github.com/strax84mb/go-travel-reactive/internal/entity"
)

type ctxIndex int

const (
	ctxPasswordIdx ctxIndex = iota + 1
)

func encodePassword(password string, salt []byte) string {
	h := sha512.New()
	_, _ = h.Write(salt)
	_, _ = h.Write([]byte(password))
	hashedPassword := h.Sum(salt)
	return hex.EncodeToString(hashedPassword)
}

func generateSalt() []byte {
	salt := make([]byte, 16)
	rand.Seed(time.Now().UnixNano())
	rand.Read(salt)
	return salt
}

func validRole(expectedRole, role, roleFromDb entity.UserRole) bool {
	if expectedRole == "ANY" {
		return true
	}
	return expectedRole == role && role == roleFromDb
}
