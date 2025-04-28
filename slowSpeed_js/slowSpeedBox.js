const logger = require('../tool/logger.js');

// 通用延时函数
function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

/**
* 带速率限制和地址过滤的延时队列处理器
* 功能特点：
* 1. 每分钟批量处理一次队列
* 2. 24小时地址去重
* 3. 单次交易金额限制
*/
module.exports = class slowSpeedBox {
/**
* @param {Function} fun - 实际处理函数，需接收 address, privateKey, messageQueue 参数
* @param {string} address - 发送方钱包地址
* @param {string} privateKey - 发送方私钥
*/
constructor(fun, address, privateKey) {
    // 初始化等待处理的地址队列
    this.addressQueue = [];
    // 24小时内禁止重复操作的地址列表
    this.banAddress = [];
    // 依赖注入的实际业务处理函数
    this.fun = fun;
    this.address = address;
    this.privateKey = privateKey;

    // 定时处理器（每分钟触发）
    this.sender = setInterval(async () => {
        if (this.addressQueue.length == 0) return;

        try {
            // 批量处理队列中的所有地址
            this.fun(this.address, this.privateKey, this.addressQueue);
            // 清空队列并记录成功日志
            this.addressQueue = [];
            for (let group of this.addressQueue) {
                logger.log("成功发送", group["amount"], "$doge给地址", group["address"])
            }
        } catch (error) {
            logger.log("slowSpeedBox Error", error.message)
        }
    }, 60000); // 60秒间隔

    // 24小时清空禁止列表的定时器
    this.bander = setInterval(() => {
        this.banAddress = [];
    }, 86400000); // 24小时间隔
}

/**
* 将消息加入处理队列
* @param {Object} message - 包含地址和金额的消息对象
* @throws {Error} 包含具体错误原因
*/
async enqueue(message) {
    try {
        // 安全校验层
        if (message["value"] > 5000) throw new Error("失败，单次收益领取大于5000");
        if (this.banAddress.includes(message["address"])) throw new Error("失败，地址24小时内领取过收益");

        // 防止重复入队检查
        for (let group of this.addressQueue) {
            if (message["address"] == group["address"]) {
                throw new Error("失败，地址已经在领取队列中");
            }
        }

        // 通过所有检查后加入队列
        this.banAddress.push(message["address"]);
        this.addressQueue.push(message);
    } catch (error) {
        throw new Error(error.message);
        return false
    }
}
}