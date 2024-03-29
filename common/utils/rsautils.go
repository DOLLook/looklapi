package utils

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"errors"
)

// 生成RSA非对称加密密钥对
func RSAGenerateKeyPair(bits int) (base64PrivateKey string, base64PubKey string, err error) {
	if bits != 1024 && bits != 2048 {
		return "", "", errors.New("bits only support 1024 or 2048")
	}

	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", err
	}

	base64PrivateKey = base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PrivateKey(rsaPrivateKey))
	pubKey, err := x509.MarshalPKIXPublicKey(&rsaPrivateKey.PublicKey)
	if err != nil {
		return "", "", err
	}
	base64PubKey = base64.StdEncoding.EncodeToString(pubKey)

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

	pub, err := x509.ParsePKIXPublicKey(keyBytes)
	if err != nil {
		return nil, err
	}
	publicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("invalid pub key")
	}

	if result, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, plain); err != nil {
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

	if result, err := rsa.DecryptPKCS1v15(rand.Reader, rsaPrivateKey, encryptBytes); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

// 签名
func RSASign(plain []byte, base64PrivateKey string) ([]byte, error) {
	if len(plain) < 1 {
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

	// same with below
	hash := sha1.New()
	hash.Write(plain)

	//hash := crypto.Hash.New(crypto.SHA1)
	//hash.Write(plain)

	// same with below
	//rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, crypto.SHA1, hash.Sum(nil))

	if sign, err := rsaPrivateKey.Sign(rand.Reader, hash.Sum(nil), crypto.SHA1); err != nil {
		return nil, err
	} else {
		return sign, nil
	}
}

// 验签
func RSACheckSign(originalBytes, signBytes []byte, base64PubKey string) error {
	if len(originalBytes) < 1 {
		return errors.New("invalid originalBytes")
	}
	if len(signBytes) < 1 {
		return errors.New("invalid signBytes")
	}
	if len(base64PubKey) < 1 {
		return errors.New("invalid pubkey")
	}

	keyBytes, err := base64.StdEncoding.DecodeString(base64PubKey)
	if err != nil {
		return err
	}

	pub, err := x509.ParsePKIXPublicKey(keyBytes)
	if err != nil {
		return err
	}
	publicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return errors.New("invalid pub key")
	}

	hash := sha1.New()
	hash.Write(originalBytes)

	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA1, hash.Sum(nil), signBytes)
}

// 转换pkcs8私钥
func ConvertPKCS8PrivateKey(base64PKCS8PrivateKey string) (base64PKCS1PrivateKey, base64ECDSAPrivateKey, base64Ed25519PrivateKey string, err error) {
	base64PKCS1PrivateKey, base64ECDSAPrivateKey, base64Ed25519PrivateKey, err = "", "", "", nil

	if IsEmpty(base64PKCS8PrivateKey) {
		err = errors.New("invalid base64PKCS8PrivateKey")
		return
	}

	if keyBytes, err1 := base64.StdEncoding.DecodeString(base64PKCS8PrivateKey); err1 != nil {
		err = err1
	} else if pkcs8, err1 := x509.ParsePKCS8PrivateKey(keyBytes); err1 != nil {
		err = err1
	} else if pkcs1, ok := pkcs8.(*rsa.PrivateKey); ok {
		base64PKCS1PrivateKey = base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PrivateKey(pkcs1))
	} else if ecdsaPk, ok := pkcs8.(*ecdsa.PrivateKey); ok {
		if ecdsaBytes, err1 := x509.MarshalECPrivateKey(ecdsaPk); err1 != nil {
			err = err1
		} else {
			base64ECDSAPrivateKey = base64.StdEncoding.EncodeToString(ecdsaBytes)
		}
	} else if pk, ok := pkcs8.(ed25519.PrivateKey); ok {
		base64Ed25519PrivateKey = base64.StdEncoding.EncodeToString(pk)
	} else {
		err = errors.New("param named base64PKCS8PrivateKey is not a valid base64 pkcs8 private key")
	}

	return
}
