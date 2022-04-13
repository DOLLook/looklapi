package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

//高级加密标准（Adevanced Encryption Standard ,AES）
//16,24,32位字符串，分别对应AES-128，AES-192，AES-256 加密方法

//PKCS7 填充模式
func pkcs7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	//Repeat()函数的功能是把切片[]byte{byte(padding)}复制padding个，然后合并成新的字节切片返回
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//填充的反向操作，删除填充字符串
func pkcs7UnPadding(origData []byte) ([]byte, error) {
	//获取数据长度
	length := len(origData)
	if length == 0 {
		return nil, errors.New("invalid encrypt data")
	} else {
		//获取填充字符串长度
		unpadding := int(origData[length-1])
		//截取切片，删除填充字节，并且返回明文
		return origData[:(length - unpadding)], nil
	}
}

//实现加密
func AesEcrypt(origData []byte, key []byte, iv []byte) ([]byte, error) {
	//创建加密算法实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	spcIV := len(iv) > 0
	blockSize := 0
	if spcIV {
		blockSize = len(iv)
	} else {
		//获取块的大小
		blockSize = block.BlockSize()
	}

	var blockMode cipher.BlockMode
	//对数据进行填充，让数据长度满足需求
	origData = pkcs7Padding(origData, blockSize)
	//采用AES加密方法中CBC加密模式
	if spcIV {
		if blockSize != block.BlockSize() {
			return nil, errors.New("IV length must equal block size")
		}
		blockMode = cipher.NewCBCEncrypter(block, iv)
	} else {
		blockMode = cipher.NewCBCEncrypter(block, key[:blockSize])
	}
	crypted := make([]byte, len(origData))
	//执行加密
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

//实现解密
func AesDecrypt(cypted []byte, key []byte, iv []byte) ([]byte, error) {
	//创建加密算法实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	//采用AES加密方法中CBC加密模式
	var blockMode cipher.BlockMode
	if len(iv) > 0 {
		if len(iv) != block.BlockSize() {
			return nil, errors.New("IV length must equal block size")
		}
		// 指定iv
		blockMode = cipher.NewCBCDecrypter(block, iv)
	} else {
		blockMode = cipher.NewCBCDecrypter(block, key[:block.BlockSize()])
	}

	origData := make([]byte, len(cypted))
	//这个函数也可以用来解密
	blockMode.CryptBlocks(origData, cypted)
	//去除填充字符串
	origData, err = pkcs7UnPadding(origData)
	if err != nil {
		return nil, err
	}
	return origData, err
}

//加密base64
func AesEcrypt2Base64(content string, key string, iv []byte) (string, error) {
	contentBytes, keyBytes := []byte(content), []byte(key)
	result, err := AesEcrypt(contentBytes, keyBytes, iv)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(result), err
}

//解密base64字符串
func AesDecryptBase64(base64Content string, key string, iv []byte) (string, error) {
	keyBytes := []byte(key)
	contentBytes, err := base64.StdEncoding.DecodeString(base64Content)
	if err != nil {
		return "", err
	}
	if decryptBytes, err := AesDecrypt(contentBytes, keyBytes, iv); err != nil {
		return "", err
	} else {
		return string(decryptBytes), nil
	}
}
