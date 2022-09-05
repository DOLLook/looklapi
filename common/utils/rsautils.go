package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
)

// 生成RSA非对称加密密钥对
func RSAGenerateKeyPair() (base64PrivateKey string, base64PubKey string, err error) {
	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	base64PrivateKey = base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PrivateKey(rsaPrivateKey))
	base64PubKey = base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PublicKey(&rsaPrivateKey.PublicKey))
	return
}

// 加密
func RSAEncrypt(plain []byte, base64PubKey string) ([]byte, error) {
	if len(plain) < 1 {
		return nil, nil
	}
	if len(base64PubKey) < 1 {
		return nil, errors.New("invalid pubkey")
	}

	keyBytes, err := base64.StdEncoding.DecodeString(base64PubKey)
	if err != nil {
		return nil, err
	}

	publicKey, err := x509.ParsePKCS1PublicKey(keyBytes)
	if err != nil {
		return nil, err
	}

	if result, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, plain, nil); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

// 解密
func RSADecrypt(encryptBytes []byte, base64PrivateKey string) ([]byte, error) {
	if len(encryptBytes) < 1 {
		return nil, nil
	}
	if len(base64PrivateKey) < 1 {
		return nil, errors.New("invalid privateKey")
	}

	keyBytes, err := base64.StdEncoding.DecodeString(base64PrivateKey)
	if err != nil {
		return nil, err
	}

	rsaPrivateKey, err := x509.ParsePKCS1PrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}

	if result, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, rsaPrivateKey, encryptBytes, nil); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}
