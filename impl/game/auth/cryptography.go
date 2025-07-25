package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
)

var secretKey *rsa.PrivateKey
var publicKey *rsa.PublicKey

var secretArr []byte
var publicArr []byte

func NewCrypt() (secret []byte, public []byte) {

	// initialize keys
	if secretKey == nil || publicKey == nil {
		var err error
		secretKey, publicKey, err = generate()
		if err != nil {
			// Log the error but try to continue - this is critical for server startup
			fmt.Printf("CRITICAL: Failed to generate encryption keys: %v\n", err)
			// Retry once
			secretKey, publicKey, err = generate()
			if err != nil {
				fmt.Printf("FATAL: Second attempt to generate encryption keys failed: %v\n", err)
				// Return empty arrays - the calling code should handle this
				return nil, nil
			}
		}

		x509Secret := x509.MarshalPKCS1PrivateKey(secretKey)
		x509Public, _ := x509.MarshalPKIXPublicKey(publicKey)

		secretArr = x509Secret
		publicArr = x509Public
	}

	return secretArr, publicArr
}

func Encrypt(data []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, publicKey, data)
}

func Decrypt(data []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, secretKey, data)
}

func generate() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	secretKey := key
	publicKey := &key.PublicKey

	secretKey.Precompute()
	if err := secretKey.Validate(); err != nil {
		return nil, nil, fmt.Errorf("RSA key validation failed: %w", err)
	}

	return secretKey, publicKey, nil
}
