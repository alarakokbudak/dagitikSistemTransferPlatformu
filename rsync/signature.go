package rsync

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

const BlockSize = 1024 * 1024 // 1 MB sabit blok (Rsync standart yaklaşımı)

// BlockSignature temsil eder (B Bilgisayarının Haritası)
type BlockSignature struct {
	Index      int
	WeakHash   uint32
	StrongHash string
}

// GenerateSignatures, eski dosyayı sabit bloklara böler ve her bloğun
// hem hızlı (Adler32) hem güçlü (SHA256) özetini çıkararak haritayı oluşturur.
func GenerateSignatures(filePath string) ([]BlockSignature, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []BlockSignature{}, nil // Dosya yoksa harita boştur
		}
		return nil, err
	}
	defer file.Close()

	var signatures []BlockSignature
	buffer := make([]byte, BlockSize)
	index := 0

	for {
		bytesRead, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if bytesRead == 0 {
			break
		}

		data := buffer[:bytesRead]

		// Hızlı Özet (Weak Hash - Basit Kayan Özet)
		weak := computeWeakHash(data)

		// Güçlü Özet (Strong Hash - SHA256)
		strongHashBytes := sha256.Sum256(data)
		strong := hex.EncodeToString(strongHashBytes[:])

		signatures = append(signatures, BlockSignature{
			Index:      index,
			WeakHash:   weak,
			StrongHash: strong,
		})
		
		index++
	}

	return signatures, nil
}

// computeWeakHash calculates a simple rolling hash for a block
func computeWeakHash(data []byte) uint32 {
	var h uint32
	for _, b := range data {
		h = (h << 1) | (h >> 31)
		h ^= uint32(b)
	}
	return h
}

// RollWeakHash efficiently updates the weak hash when the window slides 1 byte
func RollWeakHash(oldHash uint32, oldByte, newByte byte, windowSize int) uint32 {
	h := oldHash
	
	// Eski byte'ın etkisini çıkar (Ters işlem: eski byte'ı windowSize kadar sola kaydırılmış haliyle XOR'la)
	// Rotate left calculation for the old byte:
	shift := windowSize % 32
	oldEffect := uint32(oldByte)
	oldEffect = (oldEffect << shift) | (oldEffect >> (32 - shift))
	h ^= oldEffect
	
	// Pencereyi kaydır (Sola döndür)
	h = (h << 1) | (h >> 31)
	
	// Yeni byte'ı ekle
	h ^= uint32(newByte)
	
	return h
}
