// pkg/dogechain/rpc_client.go
package dogechain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"
)

// RPCClient Dogecoin RPC客户端
type RPCClient struct {
	endpoint   string
	httpClient *http.Client
	user       string
	password   string
}

// TxDetail 交易详情结构
type TxDetail struct {
	Txid          string     `json:"txid"`
	Hash          string     `json:"hash"`
	Version       int32      `json:"version"`
	Size          int32      `json:"size"`
	Vsize         int32      `json:"vsize"`
	Weight        int32      `json:"weight"`
	Vin           []TxInput  `json:"vin"`
	Vout          []TxOutput `json:"vout"`
	Hex           string     `json:"hex"`
	BlockHash     string     `json:"blockhash,omitempty"`
	Confirmations int64      `json:"confirmations,omitempty"`
	Time          int64      `json:"time,omitempty"`
	BlockTime     int64      `json:"blocktime,omitempty"`
}

// TxInput 交易输入
type TxInput struct {
	Txid      string     `json:"txid"`
	Vout      uint32     `json:"vout"`
	ScriptSig *ScriptSig `json:"scriptSig"`
	Sequence  uint32     `json:"sequence"`
	Addresses []string   `json:"addresses,omitempty"`
	Value     float64    `json:"value,omitempty"`
}

// ScriptSig 签名脚本
type ScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}

// TxOutput 交易输出
type TxOutput struct {
	Value        float64      `json:"value"`
	N            uint32       `json:"n"`
	ScriptPubKey ScriptPubKey `json:"scriptPubKey"`
}

// ScriptPubKey 公钥脚本
type ScriptPubKey struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex"`
	ReqSigs   int32    `json:"reqSigs,omitempty"`
	Type      string   `json:"type"`
	Addresses []string `json:"addresses,omitempty"`
}

// NewRPCClient 创建RPC客户端
func NewRPCClient(endpoint, user, password string) *RPCClient {
	return &RPCClient{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		user:     user,
		password: password,
	}
}

// GetAddressUTXOs 获取地址的UTXO列表
func (c *RPCClient) GetAddressUTXOs(address string) ([]UTXO, error) {
	req := map[string]interface{}{
		"jsonrpc": "1.0",
		"id":      "claimask",
		"method":  "listunspent",
		"params":  []interface{}{0, 9999999, []string{address}},
	}

	var resp struct {
		Result []struct {
			TxID   string  `json:"txid"`
			Vout   uint32  `json:"vout"`
			Amount float64 `json:"amount"`
		} `json:"result"`
		Error interface{} `json:"error"`
	}

	if err := c.rpcCall(req, &resp); err != nil {
		return nil, err
	}

	// 检查是否有错误
	if resp.Error != nil {
		return nil, fmt.Errorf("RPC error: %v", resp.Error)
	}

	utxos := make([]UTXO, len(resp.Result))
	for i, u := range resp.Result {
		utxos[i] = UTXO{
			TxHash: u.TxID,
			Index:  u.Vout,
			Value:  int64(u.Amount * 100000000), // 转换为ELON单位
		}
	}

	sort.Slice(utxos, func(i, j int) bool {
		return utxos[i].Value > utxos[j].Value
	})

	return utxos, nil
}

// GetTransaction 获取交易详情
func (c *RPCClient) GetTransaction(txid string) (*TxDetail, error) {
	req := map[string]interface{}{
		"jsonrpc": "1.0",
		"id":      "claimask",
		"method":  "getrawtransaction",
		"params":  []interface{}{txid, true},
	}

	var resp struct {
		Result TxDetail    `json:"result"`
		Error  interface{} `json:"error"`
	}

	if err := c.rpcCall(req, &resp); err != nil {
		return nil, err
	}

	// 检查是否有错误
	if resp.Error != nil {
		return nil, fmt.Errorf("RPC error: %v", resp.Error)
	}

	return &resp.Result, nil
}

// rpcCall 执行RPC调用
func (c *RPCClient) rpcCall(req interface{}, resp interface{}) error {
	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", c.endpoint, bytes.NewReader(body))
	httpReq.SetBasicAuth(c.user, c.password)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %s", httpResp.Status)
	}

	return json.NewDecoder(httpResp.Body).Decode(resp)
}
