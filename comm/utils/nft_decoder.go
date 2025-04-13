package utils

import (
	"bytes"
	"claimask/comm/constant"
	"encoding/hex"
	"errors"

	"github.com/btcsuite/btcd/txscript"
)

// 参考原decodeElon.js实现
func DecodeElonScript(scriptHex string) (string, []byte, error) {
	scriptBytes, err := hex.DecodeString(scriptHex)
	if err != nil {
		return "", nil, err
	}

	// 解析操作码
	tokenizer := txscript.MakeScriptTokenizer(0, scriptBytes)

	// 验证NFT前缀
	if !tokenizer.Next() || tokenizer.Opcode() != txscript.OP_RETURN {
		return "", nil, errors.New("invalid NFT script: missing OP_RETURN")
	}

	// 检查第二个操作码是否是数据推送
	if !tokenizer.Next() || !tokenizer.Done() {
		// 获取前缀数据
		prefixBytes := tokenizer.Data()
		prefix := string(prefixBytes)
		if prefix != constant.NFT_PREFIX {
			return "", nil, errors.New("not a doginal")
		}
	} else {
		return "", nil, errors.New("invalid NFT script: missing prefix")
	}

	// 收集后续数据块
	var buffer bytes.Buffer
	for tokenizer.Next() && !tokenizer.Done() {
		// 检查是否为数据推送操作码
		if txscript.IsSmallInt(tokenizer.Opcode()) {
			continue
		}

		// 获取数据并添加到缓冲区
		data := tokenizer.Data()
		if len(data) > 0 {
			buffer.Write(data)
		}
	}

	return constant.NFT_CONTENT_TYPE, buffer.Bytes(), nil
}
