# Claimask

#### 介绍

**项目背景**：为了回馈 NFT 社区用户，项目方决定对部分用户发放收益领取链接，并在指定的时间去进行白名单、空投资格发放，保证社区持续的热度。

#### 软件架构

后端技术栈

- Golang
- Hertz
- Redis
- MySQL
- Gorm
- Bwmarrin/snowflake 订单 ID 生层

前端技术栈

- React
- web3-react
- Javascript

#### 效果演示

1. 钱包链接

   1. ![Image text](https://github.com/Orlandoo24/claimask/blob/main/img/image-20240321220345790.png)

   2. ![Image text](https://github.com/Orlandoo24/claimask/blob/main/img/image-20240321220430461.png)

   3. ![Image text](https://github.com/Orlandoo24/claimask/blob/main/img/image-20240321220523373.png) 

2. claim 进行资格领取

   1. ![Image text](https://github.com/Orlandoo24/claimask/blob/main/img/image-20240321220621749.png)

3. 钱包断开

   1. ![Image text](https://github.com/Orlandoo24/claimask/blob/main/img/image-20240321220644503.png)

   2. ![Image text](https://github.com/Orlandoo24/claimask/blob/main/img/image-20240321220723172.png)

      

   

   



#### 使用说明

1.  克隆本项目或在发行版中下载源代码
2.  安装 node 和 golang 环境
3.  前端工程依赖：参考 claimask/web3-claimask/README.md 完成前端依赖和运行 
4.  后端工程依赖：进入 claimask/claim 运行 go mod tidy 完成依赖下载，再运行 go run claim.go 跑起后端服务
   

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
