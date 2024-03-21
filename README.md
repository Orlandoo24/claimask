# Claimask

#### 介绍

**项目背景**：为了回馈 NFT 持有用户会对某些用户发放收益领取链接并告知在指定的时间去进行白名单、空投资格领取，保证社区持续的热度。

#### 软件架构

后端技术栈

- Golang
- Heartz
- Redis
- MySQL
- Gorm
- Bwmarrin/snowflake 雪花算法生成分布式 ID

前端技术栈

- React
- web3-react
- Javascript

#### 效果演示

1. 钱包链接
2. 钱包断开
3. claim 进行资格领取



#### 使用说明

1.  克隆本项目或在发行版中下载源代码
2.  运行 sql 脚本，并启动 Nacos 与 Zipkin
3.  启动 Exam-Backstage 下的服务
4.  前端项目 npm install 安装所需依赖
5.  运行前端项目即可完成项目启动

#### 参与贡献

1.  Fork 本仓库
2.  新建 Feat_xxx 分支
3.  提交代码
4.  新建 Pull Request

#### To-Do

- [x] 后端 Redis 事务监听
- [x] 前端 web3 钱包 Metamask 链接
- [x] 前端 claim 接口的交互
- [x] claim 测试接口的编写
- [x] 奖品数量的重置接口
- [ ] 空投、或者白名单的分发
