package utils

import (
    "math/big"
    "crypto/rand"
     "net"
	"strconv"
	"strings"
	"unicode"
)

/*
var BlacklistedIPNets = []string{
	"127.0.0.0/8",
	"10.0.0.0/8",
}

var BlacklistedDomains = []string{
	"example.com",
	"badsite.org",
}
*/

// IsValidDomainOrIP checks whether s is a valid domain or IP (with optional port).
func IsValidDomainOrIP(s string, blacklistedIPNets []string, blacklistedDomains []string) bool {
	if s == "" || len(s) > 253 {
		return false
	}


	// crude port check: allow "host:port" only if exactly one colon
	if strings.Count(s, ":") == 1 {
		host, port, ok := strings.Cut(s, ":")
		if !ok {
			return false
		}
		p, err := strconv.Atoi(port)
		if err != nil || p <= 0 || p > 65535 {
			return false
		}
		s = host
	}

	// Try parsing as IP
	if ip := net.ParseIP(s); ip != nil {
		for _, cidr := range blacklistedIPNets {
			_, network, err := net.ParseCIDR(cidr)
			if err == nil && network.Contains(ip) {
				return false
			}
		}
		return true
	}

	// domain blacklist check
	sLower := strings.ToLower(s)
	for _, blk := range blacklistedDomains {
		if sLower == blk {
			return false
		}
	}

	// domain must have at least 2 labels
	labels := strings.Split(s, ".")
	if len(labels) < 2 {
		return false
	}

	for _, label := range labels {
		if len(label) < 1 || len(label) > 63 {
			return false
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, ch := range label {
			if ch > unicode.MaxASCII || !(unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '-') {
				return false
			}
		}
	}

	// check TLD
	tld := labels[len(labels)-1]
	if len(tld) < 2 {
		return false
	}
	for _, ch := range tld {
		if !unicode.IsLetter(ch) {
			return false
		}
	}

	return true
}

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
