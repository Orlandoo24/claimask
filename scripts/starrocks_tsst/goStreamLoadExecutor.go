package starrocks_tsst

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	maxRetry       = 5
	retryInterval  = 5 * time.Second
	visibleState   = "VISIBLE"
	committedState = "COMMITTED"
)

// SrCommTableModel 通用表模型
type SrCommTableModel struct {
	Database   string       // 数据库名
	Table      string       // 表名
	Columns    []ColumnMeta // 列定义
	Partition  PartitionInfo
	Properties map[string]string // 副本数等参数
}

// ColumnMeta 列元数据
type ColumnMeta struct {
	Name         string // 列名
	Type         string // 数据类型 (如 "BIGINT")
	IsNullable   bool   // 是否允许NULL
	DefaultValue string // 默认值
}

// PartitionInfo 分区信息
type PartitionInfo struct {
	Type    string   // RANGE/LIST
	Columns []string // 分区列
	Defs    []PartitionDef
}

// DistributionInfo 分桶信息
type DistributionInfo struct {
	Type    string // HASH/RANDOM
	Columns []string
	Buckets int
}

// PartitionDef 分区定义
type PartitionDef struct {
	Name       string            // 分区名称（如 p202301）
	Value      string            // 分区键值（如 "2023-01-01"）
	Buckets    int               // 分桶数量（如 10）
	Properties map[string]string `json:"properties,omitempty"`
}

// StarRocksConfig config
type StarRocksConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Timeout  time.Duration
	MaxMemMB int
}

type StreamLoader struct {
	config     *StarRocksConfig
	tableModel *SrCommTableModel
	client     *http.Client
}

// NewStreamLoader new StreamLoader
func NewStreamLoader(cfg *StarRocksConfig, model *SrCommTableModel) *StreamLoader {
	return &StreamLoader{
		config:     cfg,
		tableModel: model,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// Execute 核心导入方法
func (s *StreamLoader) Execute(ctx context.Context, format string, data [][]byte) (string, error) {
	label := generateLabel()
	url := fmt.Sprintf("http://%s:%d/api/%s/%s/_stream_load",
		s.config.Host, s.config.Port, s.tableModel.Database, s.tableModel.Table)

	body := s.buildPayload(format, data)

	for retry := 0; retry <= maxRetry; retry++ {
		req, _ := http.NewRequest("PUT", url, bytes.NewReader(body))
		s.setHeaders(req, label, format)

		resp, err := s.client.Do(req)
		if err := s.handleResponse(resp, err, label); err == nil {
			return label, nil
		}

		time.Sleep(retryInterval)
		label = generateLabel()
	}
	return "", fmt.Errorf("exceeded max retries")
}

// buildPayload 构建数据包
func (s *StreamLoader) buildPayload(format string, rows [][]byte) []byte {
	var buf bytes.Buffer
	switch format {
	case "csv":
		for _, row := range rows {
			buf.Write(row)
			buf.WriteByte('\n')
		}
	case "json":
		buf.WriteByte('[')
		for i, row := range rows {
			if i > 0 {
				buf.WriteByte(',')
			}
			buf.Write(row)
		}
		buf.WriteByte(']')
	}
	return buf.Bytes()
}

// 设置HTTP头
func (s *StreamLoader) setHeaders(req *http.Request, label string, format string) {
	req.Header.Set("Authorization", s.getAuthHeader())
	req.Header.Set("label", label)
	req.Header.Set("timeout", s.config.Timeout.String())

	switch format {
	case "csv":
		req.Header.Set("column_separator", ",")
		req.Header.Set("row_delimiter", "\n")
	case "json":
		req.Header.Set("format", "json")
		req.Header.Set("strip_outer_array", "true")
		req.Header.Set("ignore_json_size", "true")
	}
}

// 处理响应
func (s *StreamLoader) handleResponse(resp *http.Response, err error, label string) error {
	if resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&result)
			if state := result["state"]; state == visibleState || state == committedState {
				return nil
			}
		}
	}
	return fmt.Errorf("import failed for label %s", label)
}

// 生成唯一Label
func generateLabel() string {
	return fmt.Sprintf("load_%d", time.Now().UnixNano())
}

// 生成认证头
func (s *StreamLoader) getAuthHeader() string {
	auth := s.config.User + ":" + s.config.Password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}
