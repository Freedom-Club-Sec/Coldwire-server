package utils

import (
    "math/big"
    "crypto/rand"
)

// helpers
func RandomUserId() (string, error) {
    digits := ""
    for i := 0; i < 16; i++ {
        n, err := rand.Int(rand.Reader, big.NewInt(10)) // 0-9
        if err != nil {
            return "", err
        }
        digits += n.String()
    }
    return digits, nil
}

func IsAllDigits(s string) bool {
    for _, r := range s {
        if r < '0' || r > '9' {
            return false
        }
    }
    return true
}

func SecureRandomBytes(n int) ([]byte, error) {
    b := make([]byte, n)
    _, err := rand.Read(b)
    if err != nil {
        return nil, err
    }
    return b, nil
}
