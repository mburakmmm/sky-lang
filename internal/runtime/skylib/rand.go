package skylib

import (
	crand "crypto/rand"
	"encoding/hex"
	"math/big"
	"math/rand"
	"time"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// Seed sets the random seed
func RandSeed(seed int64) {
	rng = rand.New(rand.NewSource(seed))
}

// IntN returns a random integer in [0, n)
func RandIntN(n int) int {
	if n <= 0 {
		return 0
	}
	return rng.Intn(n)
}

// Int64N returns a random int64 in [0, n)
func RandInt64N(n int64) int64 {
	if n <= 0 {
		return 0
	}
	return rng.Int63n(n)
}

// RandRange returns a random integer in [min, max)
func RandRange(min, max int) int {
	if min >= max {
		return min
	}
	return min + rng.Intn(max-min)
}

// Float64 returns a random float64 in [0.0, 1.0)
func RandFloat() float64 {
	return rng.Float64()
}

// Choice returns a random element from a slice
func RandChoice(items []interface{}) interface{} {
	if len(items) == 0 {
		return nil
	}
	return items[rng.Intn(len(items))]
}

// Shuffle shuffles a slice in place
func RandShuffle(items []interface{}) {
	rng.Shuffle(len(items), func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})
}

// UUID generates a UUID v4
func RandUUID() string {
	uuid := make([]byte, 16)
	crand.Read(uuid)

	// Set version (4) and variant bits
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return hex.EncodeToString(uuid[:4]) + "-" +
		hex.EncodeToString(uuid[4:6]) + "-" +
		hex.EncodeToString(uuid[6:8]) + "-" +
		hex.EncodeToString(uuid[8:10]) + "-" +
		hex.EncodeToString(uuid[10:])
}

// RandBytes generates n random bytes (crypto-safe)
func RandBytes(n int) []byte {
	b := make([]byte, n)
	crand.Read(b)
	return b
}

// RandBigInt generates a random big integer in [0, max)
func RandBigInt(max int64) int64 {
	if max <= 0 {
		return 0
	}

	n, err := crand.Int(crand.Reader, big.NewInt(max))
	if err != nil {
		return 0
	}
	return n.Int64()
}

// Bool returns a random boolean
func RandBool() bool {
	return rng.Intn(2) == 1
}
