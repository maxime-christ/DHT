package dht

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/nu7hatch/gouuid"
	"math/big"
)

func distance(a, b []byte, bits int) *big.Int {
	var ring big.Int
	ring.Exp(big.NewInt(2), big.NewInt(int64(bits)), nil)

	var a_int, b_int big.Int
	(&a_int).SetBytes(a)
	(&b_int).SetBytes(b)

	var dist big.Int
	(&dist).Sub(&b_int, &a_int)

	(&dist).Mod(&dist, &ring)
	return &dist
}

func between(id1, id2, key string, infIncluded, supIncluded bool) bool {
	// 0 if a==b, -1 if a < b, and +1 if a > b

	id1Bytes, _ := hex.DecodeString(id1)
	id2Bytes, _ := hex.DecodeString(id2)
	keyBytes, _ := hex.DecodeString(key)

	if bytes.Compare(keyBytes, id1Bytes) == 0 && infIncluded { // keyBytes == id1Bytes
		return true
	}

	if bytes.Compare(keyBytes, id2Bytes) == 0 && supIncluded { // keyBytes == id2Bytes
		return true
	}

	if bytes.Compare(id2Bytes, id1Bytes) == 1 { // id2Bytes > id1Bytes
		if bytes.Compare(keyBytes, id2Bytes) == -1 && bytes.Compare(keyBytes, id1Bytes) == 1 { // keyBytes < id2Bytes && keyBytes > id1Bytes
			return true
		} else {
			return false
		}
	} else { // id1Bytes > id2Bytes
		if bytes.Compare(keyBytes, id1Bytes) == 1 || bytes.Compare(keyBytes, id2Bytes) == -1 { // keyBytes > id1Bytes || keyBytes < id2Bytes
			return true
		} else {
			return false
		}
	}
}

// (n + 2^(k-1)) mod (2^m)
func calcFinger(n []byte, k int, m int) (string, []byte) {
	// fmt.Println("calulcating result = (n+2^(k-1)) mod (2^m)")

	// convert the n to a bigint
	nBigInt := big.Int{}
	nBigInt.SetBytes(n)

	// fmt.Printf("n            %s\n", nBigInt.String())

	// fmt.Printf("k            %d\n", k)

	// fmt.Printf("m            %d\n", m)

	// get the right addend, i.e. 2^(k-1)
	two := big.NewInt(2)
	addend := big.Int{}
	addend.Exp(two, big.NewInt(int64(k-1)), nil)

	//fmt.Printf("2^(k-1)      %s\n", addend.String())

	// calculate sum
	sum := big.Int{}
	sum.Add(&nBigInt, &addend)

	//fmt.Printf("(n+2^(k-1))  %s\n", sum.String())

	// calculate 2^m
	ceil := big.Int{}
	ceil.Exp(two, big.NewInt(int64(m)), nil)

	//fmt.Printf("2^m          %s\n", ceil.String())

	// apply the mod
	result := big.Int{}
	result.Mod(&sum, &ceil)

	//fmt.Printf("result       %s\n", result.String())

	resultBytes := result.Bytes()
	resultHex := fmt.Sprintf("%x", resultBytes)

	//fmt.Printf("result (hex) %s\n", resultHex)

	// when resultBytes is 0, resultHex is empty string
	if resultHex == "" {
		resultHex = "00"
	}

	return resultHex, resultBytes
}

func generateNodeId() string {
	u, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	// calculate sha-1 hash
	hasher := sha1.New()
	hasher.Write([]byte(u.String()))

	return fmt.Sprintf("%x", hasher.Sum(nil))
}
