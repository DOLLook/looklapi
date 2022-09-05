package utils

import (
	"crypto/cipher"
	"crypto/des"
	"errors"
)

// DES加密
func DesCBCEncrypt(data, key, iv []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
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
	data = pkcs7Padding(data, blockSize)
	//采用AES加密方法中CBC加密模式
	if spcIV {
		if blockSize != block.BlockSize() {
			return nil, errors.New("IV length must equal block size")
		}
		blockMode = cipher.NewCBCEncrypter(block, iv)
	} else {
		blockMode = cipher.NewCBCEncrypter(block, key[:blockSize])
	}
	crypted := make([]byte, len(data))
	//执行加密
	blockMode.CryptBlocks(crypted, data)
	return crypted, nil
}

// DES解密
func DesCBCDecrypt(cypted, key, iv []byte) (origData []byte, e error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

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

	origData = make([]byte, len(cypted))

	defer func() {
		if err := recover(); err != nil {
			if throw, ok := err.(error); ok {
				origData = nil
				e = throw
				return
			} else if msg, ok := err.(string); ok {
				origData = nil
				e = errors.New(msg)
				return
			}
		}
	}()
	blockMode.CryptBlocks(origData, cypted)

	// pkcs7填充
	origData, err = pkcs7UnPadding(origData)
	if err != nil {
		return nil, err
	}

	return origData, nil
}
