package utils

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/sha3"
)

// CryptoHandler menangani semua operasi kriptografi untuk Minecraft Bedrock
type CryptoHandler struct {
	privateKey      *ecdsa.PrivateKey
	publicKey       []byte
	keyPair         *KeyPair
	encryptionKey   []byte
	decryptionKey   []byte
	salt            []byte
}

// KeyPair represents ECDSA key pair
type KeyPair struct {
	Private *ecdsa.PrivateKey
	Public  []byte
}

// NewCryptoHandler membuat instance baru CryptoHandler
func NewCryptoHandler() (*CryptoHandler, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ECDSA key: %v", err)
	}

	publicKey := elliptic.MarshalCompressed(privateKey.Curve, privateKey.PublicKey.X, privateKey.PublicKey.Y)

	return &CryptoHandler{
		privateKey: privateKey,
		publicKey:  publicKey,
		keyPair: &KeyPair{
			Private: privateKey,
			Public:  publicKey,
		},
	}, nil
}

// GenerateSalt menghasilkan salt acak
func GenerateSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// ComputeMCPEKey computes Minecraft Pocket Edition encryption key
func ComputeMCPEKey(secret, salt []byte) []byte {
	// Minecraft uses a custom key derivation
	hash := sha256.New()
	hash.Write(secret)
	hash.Write(salt)
	return hash.Sum(nil)
}

// GenerateSharedSecret generates ECDH shared secret
func (ch *CryptoHandler) GenerateSharedSecret(peerPublicKey []byte) ([]byte, error) {
	x, y := elliptic.UnmarshalCompressed(elliptic.P256(), peerPublicKey)
	if x == nil {
		return nil, errors.New("invalid public key format")
	}

	// Compute shared secret using ECDH
	sharedX, _ := ch.privateKey.Curve.ScalarMult(x, y, ch.privateKey.D.Bytes())
	if sharedX == nil {
		return nil, errors.New("failed to compute shared secret")
	}

	// Convert to bytes (big-endian)
	sharedSecret := sharedX.Bytes()

	// Pad to 32 bytes if necessary
	if len(sharedSecret) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(sharedSecret):], sharedSecret)
		sharedSecret = padded
	}

	return sharedSecret, nil
}

// SetupEncryption mengatur enkripsi dengan kunci yang diberikan
func (ch *CryptoHandler) SetupEncryption(sharedSecret []byte) error {
	// Generate salt untuk key derivation
	salt, err := GenerateSalt(16)
	if err != nil {
		return err
	}
	ch.salt = salt

	// Compute encryption keys menggunakan metode Minecraft
	encryptionKey := ComputeMCPEKey(sharedSecret, salt)
	decryptionKey := ComputeMCPEKey(sharedSecret, salt) // Bisa berbeda di implementasi nyata

	// Untuk simplicity, kita gunakan kunci yang sama untuk enkripsi/dekripsi
	// Minecraft sebenarnya menggunakan derivasi yang berbeda
	ch.encryptionKey = encryptionKey[:16] // AES-128
	ch.decryptionKey = decryptionKey[:16]

	return nil
}

// Encrypt mengenkripsi data menggunakan AES-CFB
func (ch *CryptoHandler) Encrypt(data []byte) ([]byte, error) {
	if ch.encryptionKey == nil {
		return nil, errors.New("encryption not setup")
	}

	block, err := aes.NewCipher(ch.encryptionKey)
	if err != nil {
		return nil, err
	}

	// Generate IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// Encrypt using CFB mode
	stream := cipher.NewCFBEncrypter(block, iv)
	encrypted := make([]byte, len(data))
	stream.XORKeyStream(encrypted, data)

	// Prepend IV to encrypted data
	result := append(iv, encrypted...)
	return result, nil
}

// Decrypt mendekripsi data menggunakan AES-CFB
func (ch *CryptoHandler) Decrypt(data []byte) ([]byte, error) {
	if ch.decryptionKey == nil {
		return nil, errors.New("decryption not setup")
	}

	if len(data) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	block, err := aes.NewCipher(ch.decryptionKey)
	if err != nil {
		return nil, err
	}

	// Extract IV from beginning
	iv := data[:aes.BlockSize]
	ciphertext := data[aes.BlockSize:]

	// Decrypt using CFB mode
	stream := cipher.NewCFBDecrypter(block, iv)
	decrypted := make([]byte, len(ciphertext))
	stream.XORKeyStream(decrypted, ciphertext)

	return decrypted, nil
}

// ComputeHash computes SHA256 hash of data
func ComputeHash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// ComputeMinecraftHash computes Minecraft-specific hash (SHA3-256)
func ComputeMinecraftHash(data []byte) []byte {
	hash := sha3.New256()
	hash.Write(data)
	return hash.Sum(nil)
}

// SignData menandatangani data dengan private key ECDSA
func (ch *CryptoHandler) SignData(data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	
	r, s, err := ecdsa.Sign(rand.Reader, ch.privateKey, hash[:])
	if err != nil {
		return nil, err
	}

	// Encode signature as R + S (each 32 bytes)
	signature := make([]byte, 64)
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	
	copy(signature[32-len(rBytes):32], rBytes)
	copy(signature[64-len(sBytes):64], sBytes)
	
	return signature, nil
}

// VerifySignature memverifikasi signature ECDSA
func (ch *CryptoHandler) VerifySignature(data, signature, publicKey []byte) bool {
	if len(signature) != 64 {
		return false
	}

	x, y := elliptic.UnmarshalCompressed(elliptic.P256(), publicKey)
	if x == nil {
		return false
	}

	publicKeyECDSA := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	hash := sha256.Sum256(data)
	r := new(ecdsa.Int).SetBytes(signature[:32])
	s := new(ecdsa.Int).SetBytes(signature[32:])

	return ecdsa.Verify(publicKeyECDSA, hash[:], r, s)
}

// GenerateRandomBytes menghasilkan byte acak
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	return bytes, nil
}

// PKCS7Pad menambahkan padding PKCS7
func PKCS7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// PKCS7Unpad menghapus padding PKCS7
func PKCS7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}
	
	padding := int(data[len(data)-1])
	if padding > len(data) {
		return nil, errors.New("invalid padding")
	}
	
	return data[:len(data)-padding], nil
}

// ComputeHandshakeHash computes hash for Minecraft handshake
func ComputeHandshakeHash(serverKey, clientToken, salt []byte) []byte {
	hash := sha256.New()
	hash.Write(serverKey)
	hash.Write(clientToken)
	hash.Write(salt)
	return hash.Sum(nil)
}

// GenerateClientToken generates random client token
func GenerateClientToken() ([]byte, error) {
	return GenerateRandomBytes(16)
}

// XORCipher melakukan enkripsi/dekripsi XOR sederhana
func XORCipher(data, key []byte) []byte {
	result := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		result[i] = data[i] ^ key[i%len(key)]
	}
	return result
}

// HexEncode mengencode byte ke hex string
func HexEncode(data []byte) string {
	return hex.EncodeToString(data)
}

// HexDecode mendecode hex string ke byte
func HexDecode(hexStr string) ([]byte, error) {
	return hex.DecodeString(hexStr)
}

// GetPublicKey returns the public key in compressed format
func (ch *CryptoHandler) GetPublicKey() []byte {
	return ch.publicKey
}

// GetKeyPair returns the key pair
func (ch *CryptoHandler) GetKeyPair() *KeyPair {
	return ch.keyPair
}

// GenerateServerID generates server ID for encryption
func GenerateServerID() ([]byte, error) {
	// Server ID is typically a random 8-byte value
	return GenerateRandomBytes(8)
}

// ComputeChainHash computes hash for login chain data
func ComputeChainHash(chainData [][]byte) []byte {
	hash := sha256.New()
	for _, data := range chainData {
		hash.Write(data)
	}
	return hash.Sum(nil)
}

// VerifyLoginChain verifies Minecraft login chain data
func (ch *CryptoHandler) VerifyLoginChain(chainData [][]byte, clientPublicKey []byte) bool {
	if len(chainData) == 0 {
		return false
	}

	// Verify each link in the chain
	for i := 0; i < len(chainData)-1; i++ {
		// This is simplified - real implementation would verify JWT tokens
		// and the chain of trust from Mojang/Microsoft
		if !ch.verifyChainLink(chainData[i], chainData[i+1]) {
			return false
		}
	}

	return true
}

// verifyChainLink verifies a single link in the login chain
func (ch *CryptoHandler) verifyChainLink(link, nextLink []byte) bool {
	// Simplified verification
	// In real implementation, this would verify JWT signatures
	// and check expiration times
	return len(link) > 0 && len(nextLink) > 0
}

// GenerateToken generates random authentication token
func GenerateToken() string {
	token, _ := GenerateRandomBytes(32)
	return HexEncode(token)
}

// ComputeSkinHash computes hash for skin data
func ComputeSkinHash(skinData []byte) []byte {
	return ComputeMinecraftHash(skinData)
}

// SimpleEncrypt provides simple XOR encryption for testing
func SimpleEncrypt(data, key []byte) []byte {
	encrypted := make([]byte, len(data))
	keyLen := len(key)
	for i := 0; i < len(data); i++ {
		encrypted[i] = data[i] ^ key[i%keyLen]
	}
	return encrypted
}

// SimpleDecrypt provides simple XOR decryption for testing
func SimpleDecrypt(data, key []byte) []byte {
	return SimpleEncrypt(data, key) // XOR is symmetric
}

// GenerateECKeyPair generates new ECDSA key pair
func GenerateECKeyPair() (*ecdsa.PrivateKey, []byte, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	publicKey := elliptic.MarshalCompressed(privateKey.Curve, privateKey.PublicKey.X, privateKey.PublicKey.Y)
	return privateKey, publicKey, nil
}

// LoadPrivateKeyFromPEM loads private key from PEM format (for future use)
func LoadPrivateKeyFromPEM(pemData []byte) (*ecdsa.PrivateKey, error) {
	// Implementation for loading from PEM
	// This would be used for persistent key storage
	return nil, errors.New("not implemented")
}

// SavePrivateKeyToPEM saves private key to PEM format (for future use)
func SavePrivateKeyToPEM(privateKey *ecdsa.PrivateKey) ([]byte, error) {
	// Implementation for saving to PEM
	// This would be used for persistent key storage
	return nil, errors.New("not implemented")
}