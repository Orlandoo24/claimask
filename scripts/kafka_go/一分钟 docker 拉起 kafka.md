### Kafka 3.5 集群 + Kafka UI 的完整安装步骤（使用 Docker Compose 和 SASL/SCRAM 认证）：

1. 环境准备
   1.1 安装 Docker 和 Docker Compose
# 添加阿里云镜像源（TencentOS/CentOS）
sudo yum-config-manager --add-repo https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo

# 安装 Docker
sudo yum install -y docker-ce-20.10.9-3.* docker-ce-cli-20.10.9-3.*
sudo systemctl start docker
sudo systemctl enable docker

# 安装 Docker Compose 插件
sudo yum install docker-compose-plugin
1.2 创建项目目录
mkdir kafka-cluster && cd kafka-cluster

2. 配置文件准备
   2.1 创建 docker-compose.yml
   cat > docker-compose.yml <<'EOF'
   version: "2"

services:
kafka-0:
image: bitnami/kafka:3.5-debian-12
container_name: kafka-0
ports:
- 19092:9092
- 19093:9093      
environment:
- KAFKA_CFG_NODE_ID=0
- KAFKA_CFG_PROCESS_ROLES=controller,broker
- KAFKA_CFG_CONTROLLER_QUORUM_VOTERS=0@kafka-0:9093
- KAFKA_CFG_LISTENERS=SASL_PLAINTEXT://:9092,CONTROLLER://:9093
- KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,SASL_PLAINTEXT:SASL_PLAINTEXT
- KAFKA_CFG_ADVERTISED_LISTENERS=SASL_PLAINTEXT://${HOST_CONFIG}:19092

      - KAFKA_CLIENT_USERS=admin
      - KAFKA_CLIENT_PASSWORDS=123456
      
      - KAFKA_CFG_CONTROLLER_LISTENER_NAMES=CONTROLLER
      - KAFKA_CFG_SASL_MECHANISM_CONTROLLER_PROTOCOL=PLAIN
      - KAFKA_CONTROLLER_USER=admin
      - KAFKA_CONTROLLER_PASSWORD=123456
      
      - KAFKA_CFG_INTER_BROKER_LISTENER_NAME=SASL_PLAINTEXT
      - KAFKA_CFG_SASL_MECHANISM_INTER_BROKER_PROTOCOL=SCRAM-SHA-256
      - KAFKA_INTER_BROKER_USER=admin
      - KAFKA_INTER_BROKER_PASSWORD=123456
    volumes:
      - ./data/kafka0:/root/bitnami/kafka
    networks:
      - mx-wk

kafkaui:
image: provectuslabs/kafka-ui:latest   
container_name: kafkaui
ports:
- 7080:8080
depends_on:
- kafka-0
environment:
- DYNAMIC_CONFIG_ENABLED=true
- KAFKA_CLUSTERS_0_NAME=mx
- KAFKA_CLUSTERS_0_BOOTSTRAPSERVERS=${HOST_CONFIG}:19092
- KAFKA_CLUSTERS_0_PROPERTIES_SECURITY_PROTOCOL=SASL_PLAINTEXT
- KAFKA_CLUSTERS_0_PROPERTIES_SASL_MECHANISM=SCRAM-SHA-256
- KAFKA_CLUSTERS_0_PROPERTIES_SASL_JAAS_CONFIG=org.apache.kafka.common.security.scram.ScramLoginModule required username="admin" password="123456";
networks:
- mx-wk

networks:
mx-wk:
external: true
EOF
2.2 创建安装脚本 install.sh
cat > install.sh <<'EOF'
#!/bin/bash

# 获取本机IP（自动适配不同系统）
HOST_CONFIG=$(hostname -I | awk '{print $1}')
echo "HOST_CONFIG=${HOST_CONFIG}" > .env

# 创建网络和数据目录
docker network create mx-wk
mkdir -p data/kafka0

# 启动服务
docker compose up -d

# 输出访问信息
echo -e "\n\033[32m[安装成功]\033[0m"
echo "Kafka UI 访问地址: http://${HOST_CONFIG}:7080"
echo "认证用户名: admin"
echo "认证密码: 123456"
echo "Kafka Broker 地址: ${HOST_CONFIG}:19092"
EOF

# 赋予执行权限
chmod +x install.sh

3. 执行安装
# 运行安装脚本
./install.sh

# 查看服务状态
docker ps -a

4. 验证服务
   4.1 访问 Kafka UI
   • 打开浏览器访问：http://<服务器IP>:7080
   • 使用账号 admin 和密码 123456 登录
   4.2 命令行测试
# 进入Kafka容器
docker exec -it kafka-0 bash

# 创建测试Topic（需认证）
kafka-topics.sh --create \
--bootstrap-server localhost:9092 \
--topic test \
--partitions 1 \
--replication-factor 1 \
--command-config /opt/bitnami/kafka/config/consumer.properties

# 生产消息
echo "Hello Kafka" | kafka-console-producer.sh \
--broker-list localhost:9092 \
--topic test \
--producer.config /opt/bitnami/kafka/config/producer.properties

# 消费消息
kafka-console-consumer.sh \
--bootstrap-server localhost:9092 \
--topic test \
--from-beginning \
--consumer.config /opt/bitnami/kafka/config/consumer.properties

5. 生产环境建议
1. 修改默认密码
   编辑 docker-compose.yml 中的以下变量：
   KAFKA_CLIENT_PASSWORDS=your_strong_password
   KAFKA_CONTROLLER_PASSWORD=your_strong_password
   KAFKA_INTER_BROKER_PASSWORD=your_strong_password
2. 启用SSL加密
   参考文档中的 SSL配置脚本 生成证书，并修改 docker-compose.yml 的监听器配置。
3. 扩展集群节点
   复制 kafka-0 服务配置，修改为 kafka-1 和 kafka-2，并调整以下参数：
   KAFKA_CFG_NODE_ID=1
   KAFKA_CFG_CONTROLLER_QUORUM_VOTERS=0@kafka-0:9093,1@kafka-1:9093,2@kafka-2:9093

6. 卸载清理
# 删除容器和数据
docker compose down
docker network rm mx-wk
sudo rm -rf data/

# 删除镜像（可选）
docker rmi bitnami/kafka:3.5-debian-12 provectuslabs/kafka-ui:latest

通过以上步骤，您已成功部署了一个 支持 SASL/SCRAM 认证的 Kafka 3.5 单节点集群，并可通过 Web UI 管理。如需扩展为多节点集群或启用 SSL，请参考注释中的高级配置。




ps:
# Kafka 配合 SSL
原文见 https://github.com/bitnami/containers/blob/main/bitnami/kafka/README.md#security
采用官方提供的生成脚本，生成 SSL 公钥/密钥。
https://raw.githubusercontent.com/confluentinc/confluent-platform-security-tools/master/kafka-generate-ssl.sh