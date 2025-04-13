#!/bin/bash
# 文件名：go_lint_collector.sh
# 功能：收集Go项目编译错误和静态检查结果

# 目标输出文件路径
OUTPUT_FILE="/Users/123jiaru/Desktop/project/my/astro/astro-rebuild/alpha/astro-orderx/lint_results.txt"

# 进入项目根目录
cd /Users/123jiaru/Desktop/project/my/astro/astro-rebuild/alpha/astro-orderx || {
    echo "错误：无法进入项目目录"
    exit 1
}

# 清空或创建输出文件
: > "$OUTPUT_FILE"

# 1. 执行go build编译
echo "=== Go Build 编译结果 ===" >> "$OUTPUT_FILE"
go build ./... 2>&1 | tee -a "$OUTPUT_FILE"
echo -e "\n\n" >> "$OUTPUT_FILE"

# 2. 执行go vet静态分析
echo "=== Go Vet 静态分析 ===" >> "$OUTPUT_FILE"
go vet ./... 2>&1 | tee -a "$OUTPUT_FILE"
echo -e "\n\n" >> "$OUTPUT_FILE"

# 3. 执行golangci-lint检查
if command -v golangci-lint &> /dev/null; then
    echo "=== GolangCI-Lint 检查 ===" >> "$OUTPUT_FILE"
    golangci-lint run ./... 2>&1 | tee -a "$OUTPUT_FILE"
else
    echo "GolangCI-Lint 未安装，跳过高级静态检查" >> "$OUTPUT_FILE"
fi

# 4. 收集测试错误（不运行测试）
echo "=== 测试编译检查 ===" >> "$OUTPUT_FILE"
go test -run=none 2>&1 | tee -a "$OUTPUT_FILE"

echo -e "\n\n检查完成，结果已保存到: $OUTPUT_FILE"