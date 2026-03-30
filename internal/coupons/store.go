package coupons

import "strings"

// Store answers whether a promo code is valid.
type Store interface {
	IsValid(code string) bool
}

type InMemoryStore struct {
	valid map[uint64]struct{} // hashed + normalized uppercase codes
}

func NewInMemoryStore(valid map[uint64]struct{}) *InMemoryStore {
	out := make(map[uint64]struct{}, len(valid))
	for h := range valid {
		out[h] = struct{}{}
	}
	return &InMemoryStore{valid: out}
}

func (s *InMemoryStore) IsValid(code string) bool {
	// Avoid allocating uppercase strings on every request.
	code = strings.TrimSpace(code)
	if code == "" {
		return false
	}
	h, ok := hashPromoCodeUpper(code)
	if !ok {
		return false
	}
	_, ok = s.valid[h]
	return ok
}

// hashPromoCodeUpper returns a case-insensitive hash for `code`.
// A "valid" promo code must be 8-10 chars and all alphanumeric.
func hashPromoCodeUpper(code string) (uint64, bool) {
	// Rule says “must be a string of length between 8 and 10 characters”.
	if len(code) < 8 || len(code) > 10 {
		return 0, false
	}

	// FNV-1a 64-bit
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)
	var h uint64 = offset64

	for i := 0; i < len(code); i++ {
		b := code[i]
		// Normalize ASCII letters to uppercase.
		if b >= 'a' && b <= 'z' {
			b = b - 'a' + 'A'
		}
		// Must be alphanumeric.
		if (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') {
			h ^= uint64(b)
			h *= prime64
			continue
		}
		return 0, false
	}

	return h, true
}

