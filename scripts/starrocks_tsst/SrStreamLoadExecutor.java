package com.clougence.cloudcanal.starrocks.worker.writer;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;

import org.apache.commons.codec.binary.Base64;
import org.apache.http.HttpEntity;
import org.apache.http.client.config.RequestConfig;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.client.methods.HttpPut;
import org.apache.http.entity.ByteArrayEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.DefaultRedirectStrategy;
import org.apache.http.impl.client.HttpClientBuilder;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.util.EntityUtils;

import com.clougence.cloudcanal.base.metadata.config.rdb.starrocks.SrOrDorisTableModel;
import com.clougence.cloudcanal.base.metadata.config.rdb.starrocks.StarRocksTargetConfig;
import com.clougence.cloudcanal.base.service.task.util.JacksonUtil;
import com.clougence.cloudcanal.starrocks.worker.util.StarRocksDelimiterHelper;
import com.clougence.utils.ExceptionUtils;
import com.clougence.utils.io.IOUtils;
import com.clougence.utils.json.JSON;

import lombok.extern.slf4j.Slf4j;

/**
 * StarRocks Stream Load 数据导入执行器
 * 功能：通过HTTP协议将CSV/JSON格式数据批量导入StarRocks表
 *
 * @author bucketli 2022/4/26 11:23:11
 */

@Slf4j
public class SrStreamLoadExecutor {

    // StarRocks返回状态常量
    private static final String RESULT_FAILED = "Fail";  // 导入失败
    private static final String RESULT_LABEL_EXISTED = "Label Already Exists";  // 标签已存在
    private static final String LAEBL_STATE_VISIBLE = "VISIBLE";  // 数据可见（导入成功）
    private static final String LAEBL_STATE_COMMITTED = "COMMITTED";  // 事务已提交
    private static final String RESULT_LABEL_PREPARE = "PREPARE";  // 准备中
    private static final String RESULT_LABEL_ABORTED = "ABORTED";  // 已中止
    private static final String RESULT_LABEL_UNKNOWN = "UNKNOWN";  // 未知状态

    // 配置参数
    private final StarRocksTargetConfig dstConfig;  // 目标StarRocks集群配置
    private final StarRocksValFormat format;  // 数据格式（CSV/JSON）
    private final SrOrDorisTableModel tableModel;  // 目标表模型
    private final byte[] lineDelimiter;  // 行分隔符（字节形式）

    // 超时和重试配置
    private Integer httpSoTimeout = 60 * 1000;  // Socket超时（默认60秒）
    private static int MAX_RETRY = 5;  // 最大重试次数
    private static int RETRY_SLEEP_MS = 5000;  // 重试间隔（默认5秒）

    /​**
     * 构造函数
     * @param dstConfig StarRocks目标配置（包含主机、端口、认证等信息）
     * @param format 数据格式枚举（CSV/JSON）
     * @param tableModel 表结构模型
     */
    public SrStreamLoadExecutor(StarRocksTargetConfig dstConfig, StarRocksValFormat format, SrOrDorisTableModel tableModel) {
        this.dstConfig = dstConfig;
        this.format = format;
        this.tableModel = tableModel;
        // 解析行分隔符（默认换行符\n）
        this.lineDelimiter = StarRocksDelimiterHelper.parse("\\" + dstConfig.getLineSeparator(), "\n").getBytes(StandardCharsets.UTF_8);

        // 覆盖默认配置（如果传入参数有效）
        if (dstConfig.getHttpSoTimeoutSec() != null && dstConfig.getHttpSoTimeoutSec() > 0) {
            this.httpSoTimeout = dstConfig.getHttpSoTimeoutSec() * 1000;
        }
        if (dstConfig.getRetryCount() != null) {
            MAX_RETRY = dstConfig.getRetryCount();
        }
        if (dstConfig.getRetryWaitTimeMs() != null) {
            RETRY_SLEEP_MS = dstConfig.getRetryWaitTimeMs();
        }
    }

    /​**​
     * 执行Stream Load导入
     * @param label 任务唯一标识（建议使用UUID）
     * @param dbName 目标数据库名
     * @param tableName 目标表名
     * @param columns 目标表列名列表（顺序需与数据对应）
     * @param rows 待导入数据行（字节数组形式）
     * @param totalBytes 数据总字节数（用于预分配缓冲区）
     * @param enableEasyMatchMode 是否启用简单匹配模式（JSON字段名直接映射列名）
     * @return 最终成功的label
     * @throws IOException 导入失败时抛出
     */
    public String doStreamLoad(String label, String dbName, String tableName, List<String> columns,
                              List<byte[]> rows, int totalBytes, boolean enableEasyMatchMode) throws IOException {
        // 构造Stream Load API地址
        String loadUrl = "http://" + dstConfig.getHttpHost() + "/api/" + dbName + "/" + tableName + "/_stream_load";
        boolean finished = false;
        int retry = 0;

        // 重试逻辑（最多MAX_RETRY次）
        while (!finished && retry <= MAX_RETRY) {
            try {
                // 发送HTTP PUT请求
                Map<String, Object> loadResult = doHttpPut(loadUrl, label, joinRows(rows, totalBytes), columns, enableEasyMatchMode);

                // 解析返回结果
                final String keyStatus = "Status";
                if (null == loadResult || !loadResult.containsKey(keyStatus)) {
                    log.error("Invalid response: " + JacksonUtil.toJson(loadResult));
                    throw new IOException("Unable to flush data: unknown result status.");
                }

                // 处理不同状态
                if (RESULT_FAILED.equals(loadResult.get(keyStatus))) {
                    throw new IOException("Import failed: " + JacksonUtil.toJson(loadResult));
                } else if (RESULT_LABEL_EXISTED.equals(loadResult.get(keyStatus))) {
                    // 标签冲突时检查任务最终状态
                    checkLabelState(dstConfig.getHttpHost(), label, dbName);
                } else {
                    // 导入成功
                    finished = true;
                }
            } catch (Exception e) {
                log.warn("Retry {}: Failed to import data. Error: {}", retry, ExceptionUtils.getRootCauseMessage(e), e);
                try {
                    Thread.sleep(RETRY_SLEEP_MS);
                } catch (InterruptedException ex) {
                    Thread.currentThread().interrupt();
                }
            } finally {
                retry++;
                // 每次重试生成新label避免冲突
                label = UUID.randomUUID().toString();
            }
        }

        if (!finished) {
            throw new IOException("Import failed after " + retry + " retries.");
        }
        return label;
    }

    /**
     * 合并多行数据为单个字节数组（根据格式添加分隔符）
     * @param rows 原始数据行列表
     * @param totalBytes 预估总字节数
     * @return 合并后的字节数组
     */
    private byte[] joinRows(List<byte[]> rows, int totalBytes) {
        if (format == StarRocksValFormat.csv) {
            // CSV格式：每行末尾添加换行符
            ByteBuffer buffer = ByteBuffer.allocate(totalBytes + rows.size() * lineDelimiter.length);
            for (byte[] row : rows) {
                buffer.put(row);
                buffer.put(lineDelimiter);
            }
            return buffer.array();
        } else if (format == StarRocksValFormat.json) {
            // JSON格式：包装为JSON数组（如 [{"a":1},{"a":2}]）
            ByteBuffer buffer = ByteBuffer.allocate(totalBytes + (rows.isEmpty() ? 2 : rows.size() + 1));
            buffer.put("[".getBytes(StandardCharsets.UTF_8));
            byte[] jsonDelimiter = ",".getBytes(StandardCharsets.UTF_8);
            boolean isFirstElement = true;
            for (byte[] row : rows) {
                if (!isFirstElement) {
                    buffer.put(jsonDelimiter);
                }
                buffer.put(row);
                isFirstElement = false;
            }
            buffer.put("]".getBytes(StandardCharsets.UTF_8));
            return buffer.array();
        } else {
            throw new RuntimeException("Unsupported format: " + format);
        }
    }

    /​**​
     * 检查指定label的导入状态（轮询）
     * @param host StarRocks主机地址
     * @param label 任务标识
     * @param dbName 数据库名
     * @throws IOException 状态异常时抛出
     */
    @SuppressWarnings("unchecked")
    protected void checkLabelState(String host, String label, String dbName) throws IOException {
        int retryCount = 0;
        while (true) {
            try {
                // 指数退避策略（最多等待5秒）
                TimeUnit.SECONDS.sleep(Math.min(++retryCount, 5));

                try (CloseableHttpClient httpclient = HttpClients.createDefault()) {
                    // 构造状态查询请求
                    HttpGet httpGet = new HttpGet("http://" + host + "/api/" + dbName + "/get_load_state?label=" + label);
                    httpGet.setHeader("Authorization", getBasicAuthHeader(dstConfig.getUserName(), dstConfig.getPassword()));
                    httpGet.setHeader("Connection", "close");

                    try (CloseableHttpResponse resp = httpclient.execute(httpGet)) {
                        HttpEntity respEntity = resp.getEntity();
                        if (resp.getStatusLine().getStatusCode() != 200) {
                            throw new IOException("Status check failed for label: " + label);
                        }

                        // 解析状态
                        Map<String, Object> result = (Map<String, Object>) JSON.parse(EntityUtils.toString(respEntity));
                        String labelState = (String) result.get("state");
                        if (labelState == null) {
                            throw new IOException("Invalid state response: " + result);
                        }

                        // 处理状态
                        switch (labelState) {
                            case LAEBL_STATE_VISIBLE:
                            case LAEBL_STATE_COMMITTED:
                                return;  // 成功
                            case RESULT_LABEL_PREPARE:
                                continue;  // 继续轮询
                            case RESULT_LABEL_ABORTED:
                                throw new RuntimeException("Import aborted for label: " + label);
                            default:
                                throw new IOException("Unknown state: " + labelState);
                        }
                    }
                }
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                throw new IOException("Interrupted during state check", e);
            }
        }
    }

    /​**​
     * 发送HTTP PUT请求执行Stream Load
     * @param loadUrl API地址
     * @param label 任务标识
     * @param data 待发送数据（字节数组）
     * @param columns 列名列表
     * @param enableEasyMatchMode 是否启用简单匹配模式
     * @return StarRocks返回的JSON结果（Map形式）
     * @throws IOException 请求失败时抛出
     */
    @SuppressWarnings("unchecked")
    private Map<String, Object> doHttpPut(String loadUrl, String label, byte[] data,
                                         List<String> columns, boolean enableEasyMatchMode) throws IOException {
        // 配置HTTP客户端（支持重定向）
        final HttpClientBuilder httpClientBuilder = HttpClients.custom()
                .setRedirectStrategy(new DefaultRedirectStrategy() {
                    @Override
                    protected boolean isRedirectable(String method) {
                        return true;
                    }
                });

        try (CloseableHttpClient httpclient = httpClientBuilder.build()) {
            HttpPut httpPut = new HttpPut(loadUrl);

            // 设置列映射（CSV或非简单匹配模式时必需）
            if ((format == StarRocksValFormat.csv || !enableEasyMatchMode) && columns != null && !columns.isEmpty()) {
                String colStr = format == StarRocksValFormat.json ?
                    String.join(",", columns) :  // JSON直接拼接列名
                    columns.stream().map(f -> String.format("`%s`", f.trim().replace("`", ""))).collect(Collectors.joining(","));  // CSV列名用反引号包裹
                httpPut.setHeader("columns", colStr);
            }

            // 基础头信息
            httpPut.setHeader("timezone", String.valueOf(dstConfig.getTimezone()));
            httpPut.setHeader("timeout", String.valueOf(dstConfig.getConnectionTimeoutSec()));
            httpPut.setHeader("exec_mem_limit", String.valueOf(dstConfig.getLoadExecMemLimitMb() * 1024 * 1024));
            httpPut.setHeader("Expect", "100-continue");  // HTTP 100 Continue
            httpPut.setHeader("label", label);
            httpPut.setHeader("Content-Type", "text/html");
            httpPut.setHeader("Authorization", getBasicAuthHeader(dstConfig.getUserName(), dstConfig.getPassword()));
            httpPut.setEntity(new ByteArrayEntity(data));
            httpPut.setConfig(RequestConfig.custom()
                    .setRedirectsEnabled(true)
                    .setSocketTimeout(httpSoTimeout)
                    .build());

            // 格式相关头信息
            if (format == StarRocksValFormat.csv) {
                httpPut.setHeader("column_separator", "\\" + dstConfig.getColumnSeparator());
                httpPut.setHeader("row_delimiter", "\\" + dstConfig.getLineSeparator());
            } else if (format == StarRocksValFormat.json) {
                httpPut.setHeader("format", "json");
                httpPut.setHeader("strip_outer_array", "true");  // 去除外层数组
                httpPut.setHeader("ignore_json_size", "true");  // 忽略JSON大小限制
            } else {
                throw new UnsupportedOperationException("Unsupported format: " + format);
            }

            // 执行请求并解析响应
            try (CloseableHttpResponse resp = httpclient.execute(httpPut)) {
                HttpEntity respEntity = resp.getEntity();
                if (resp.getStatusLine().getStatusCode() != 200) {
                    String errorMsg = respEntity != null ?
                        EntityUtils.toString(respEntity) : "Empty response";
                    throw new IOException("HTTP error: " + errorMsg);
                }
                return (Map<String, Object>) JSON.parse(EntityUtils.toString(respEntity));
            }
        }
    }

    /**
     * 生成Basic Auth认证头
     * @param username 用户名
     * @param password 密码
     * @return Base64编码的认证字符串
     */
    private String getBasicAuthHeader(String username, String password) {
        String auth = username + ":" + password;
        byte[] encodedAuth = Base64.encodeBase64(auth.getBytes(StandardCharsets.UTF_8));
        return "Basic " + new String(encodedAuth);
    }
}