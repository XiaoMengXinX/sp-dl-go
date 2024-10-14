package playplay

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"io"
)

// createAESCTRCipher 创建 AES-CTR 解密流
func createAESCTRCipher(key []byte) (cipher.Stream, error) {
	if len(key) != 16 {
		return nil, fmt.Errorf("invalid key length: expected 16 bytes, got %d", len(key))
	}

	nonce, _ := hex.DecodeString("72e067fbddcbcf77")
	iv, _ := hex.DecodeString("ebe8bc643f630d93")
	// 合并 nonce 和 iv 为 16 字节的初始化向量
	ctrIV := append(nonce, iv...)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %v", err)
	}

	// 创建 CTR 模式的流
	stream := cipher.NewCTR(block, ctrIV)
	return stream, nil
}

// DecryptFileStream 使用文件流进行解密，并跳过前 167 字节
func DecryptFileStream(key []byte, input io.Reader, output io.Writer) error {
	// 创建 AES-CTR 解密流
	stream, err := createAESCTRCipher(key)
	if err != nil {
		return err
	}

	// 读取整个输入文件到内存中（根据您的需求）
	data, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("failed to read input: %v", err)
	}

	// 解密数据
	decrypted := make([]byte, len(data))
	stream.XORKeyStream(decrypted, data)

	// 写入解密后的数据，从第 167 字节开始
	if len(decrypted) > 167 {
		if _, err := output.Write(decrypted[167:]); err != nil {
			return fmt.Errorf("failed to write output: %v", err)
		}
	} else {
		return fmt.Errorf("decrypted data is too short")
	}

	return nil
}
