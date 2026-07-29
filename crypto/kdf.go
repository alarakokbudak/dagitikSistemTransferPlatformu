package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
)

// PBKDF2 implements the Password-Based Key Derivation Function 2
// using HMAC-SHA256 as the pseudorandom function.
// It mathematically stretches a simple password into a strong AES key.
func PBKDF2(password, salt []byte, iter, keyLen int) []byte {
	prf := hmac.New(sha256.New, password)
	hashLen := prf.Size()
	numBlocks := (keyLen + hashLen - 1) / hashLen

	var buf [4]byte
	dk := make([]byte, 0, numBlocks*hashLen)
	U := make([]byte, hashLen)
	T := make([]byte, hashLen)

	for block := 1; block <= numBlocks; block++ {
		// U_1 = PRF(Password, Salt || INT_32_BE(i))
		buf[0] = byte(block >> 24)
		buf[1] = byte(block >> 16)
		buf[2] = byte(block >> 8)
		buf[3] = byte(block)

		prf.Reset()
		prf.Write(salt)
		prf.Write(buf[:])
		U = prf.Sum(U[:0])
		copy(T, U)

		// U_c = PRF(Password, U_{c-1})
		for i := 2; i <= iter; i++ {
			prf.Reset()
			prf.Write(U)
			U = prf.Sum(U[:0])
			for x := range T {
				T[x] ^= U[x]
			}
		}
		dk = append(dk, T...)
	}
	return dk[:keyLen]
}
