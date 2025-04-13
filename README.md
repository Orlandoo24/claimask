
# Claimask

## 介绍

**项目背景**：Claimask 项目致力于为NFT社区用户提供一个安全、高效的白名单领取和空投资格发放平台。通过根据用户持有的NFT数量和稀有度，进行meme币和ETH的空投，旨在增强社区活跃度和资产价值。

## 系统架构

### 整体架构

Claimask 采用前后端分离架构，后端基于 Go 语言微服务设计，前端使用 React 构建交互界面。

```
┌───────────────┐     ┌─────────────────────────────────┐     ┌────────────────┐
│               │     │           Backend Services       │     │                │
│  Web Frontend │────▶│                                 │◀───▶│  External APIs │
│  (React/Web3) │     │  ┌─────────┐      ┌─────────┐   │     │  (Blockchain)  │
│               │     │  │ClaimMask│      │ Monitor │   │     │                │
└───────────────┘     │  │ Service │      │ Service │   │     └────────────────┘
                      │  └─────────┘      └─────────┘   │           ▲  
                      │        │                │       │           │  
                      └────────┼────────────────┼───────┘           │  
                               │                │                   │  
                      ┌────────▼────────────────▼───────┐           │  
                      │       Data Storage Layer        │           │  
                      │                                 │           │  
                      │  ┌─────────┐      ┌─────────┐   │           │  
                      │  │  MySQL  │      │  Redis  │───┼───────────┘  
                      │  └─────────┘      └─────────┘   │              
                      │                                 │              
                      └─────────────────────────────────┘              
```

### 技术栈

#### 后端技术栈
- **Golang**: 核心开发语言
- **Gin**: Web框架
- **Gorm**: ORM数据库映射
- **Redis**: 缓存和分布式锁
- **MySQL**: 持久化存储
- **Bwmarrin/snowflake**: 分布式ID生成
- **Dogecoin RPC**: 区块链交互

#### 前端技术栈
- React
- Web3-react
- Javascript

### 核心模块

1. **ClaimMask 服务**
   - 负责奖品领取逻辑
   - 支持并发安全的领取机制
   - 实现分布式锁防止超发
   - 基于 Redis 的原子操作实现高性能处理

2. **Monitor 服务**
   - 区块链交易监控
   - NFT 交易跟踪与验证
   - 支持 WebSocket 和轮询双通道监控
   - 税收状态验证

3. **异步任务处理**
   - 基于优先级队列的任务处理系统
   - 支持速率限制与重试策略
   - 事务一致性保证

4. **数据访问层**
   - 订单 DAO
   - NFT DAO
   - UTXO DAO
   - 支持事务与批量操作

## 效果演示

1. 钱包链接
   1. ![钱包连接界面](https://github.com/Orlandoo24/claimask/pkg/claimask-ui/doc/img/image-20240321220345790.png)
   2. ![钱包选择界面](https://github.com/Orlandoo24/claimask/pkg/claimask-ui/doc/img/doc/img/image-20240321220430461.png)
   3. ![连接成功界面](https://github.com/Orlandoo24/claimask/pkg/claimask-ui/doc/img/img/image-20240321220523373.png)

2. Claim 资格领取
   1. ![Claim界面](https://github.com/Orlandoo24/claimask/pkg/claimask-ui/doc/img/image-20240321220621749.png)

3. 钱包断开
   1. ![断开提示](https://github.com/Orlandoo24/claimask/pkg/claimask-ui/doc/img/image-20240321220644503.png)
   2. ![断开后界面](https://github.com/Orlandoo24/claimask/pkg/claimask-ui/doc/img/image-20240321220723172.png)

## 关键流程

### 空投/奖品领取流程
1. 用户连接钱包
2. 系统验证NFT持有状态
3. 用户点击Claim按钮
4. 后端执行原子化扣减奖品库存
5. 生成唯一订单ID
6. 记录领取信息并异步发放奖励

### 交易监控流程
1. 系统通过WebSocket和定时轮询监听区块链交易
2. 解析交易数据，识别NFT相关操作
3. 验证交易有效性及税收状态
4. 更新NFT所有权信息
5. 触发相关业务流程

## 使用说明

1. 克隆本项目或在发行版中下载源代码
2. 安装 node 和 golang 环境
3. 前端工程依赖：参考 claimask/web3-claimask/README.md 完成前端依赖和运行
4. 后端工程依赖：进入 claimask/claim 运行 go mod tidy 完成依赖下载，再运行 go run claim.go 跑起后端服务

## 参与贡献

1. Fork 本仓库
2. 新建 Feat_xxx 分支
3. 提交代码
4. 新建 Pull Request

## 项目进度

- [x] 后端 Redis 事务监听
- [x] 前端 web3 钱包 Metamask 链接
- [x] 前端 claim 接口的交互
- [x] claim 测试接口的编写
- [x] 奖品数量的重置接口
- [x] 收益异步发放
- [x] 收益领取状态更新加锁
- [ ] 空投、或者白名单的分发
- [ ] 发放手续费签名

## API文档

详细API文档请参考项目中的 `docs/api.md` 文件。
