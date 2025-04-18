package starrocks_tsst

import (
	"context"
	"fmt"
	"time"
)

func main() {
	// 初始化配置
	config := &StarRocksConfig{
		Host:     "fe-host",
		Port:     8030,
		User:     "admin",
		Password: "starrocks_tsst",
		Timeout:  60 * time.Second,
		MaxMemMB: 2048,
	}

	// 定义表模型
	tableModel := &SrCommTableModel{
		Database: "test_db",
		Table:    "users",
		Columns: []ColumnMeta{
			{Name: "id", Type: "BIGINT"},
			{Name: "name", Type: "VARCHAR(255)"},
		},
	}

	// 准备数据
	data := [][]byte{
		[]byte(`{"id":1,"name":"Alice"}`),
		[]byte(`{"id":2,"name":"Bob"}`),
	}

	// 执行导入
	loader := NewStreamLoader(config, tableModel)
	if label, err := loader.Execute(context.Background(), "json", data); err == nil {
		fmt.Printf("Import success! Label: %s\n", label)
	} else {
		fmt.Printf("Import failed: %v\n", err)
	}
}
