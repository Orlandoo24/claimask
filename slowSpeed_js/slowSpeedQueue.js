const logger = require('../tool/logger.js');

// 通用延时函数
function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

/**
* 串行化任务队列处理器
* 功能特点：
* 1. 先进先出顺序处理
* 2. 固定15秒间隔执行
* 3. 自动队列延续
*/
module.exports = class slowSpeedQueue {
/**
* @param {Function} fun - 实际处理函数，需支持异步操作
* @param {string} address - 相关地址
* @param {string} privateKey - 相关私钥
*/
constructor(fun, address, privateKey) {
  // 任务存储队列
  this.queue = [];
  // 处理状态锁
  this.isPending = false;
  // 依赖注入的业务函数
  this.fun = fun;
  this.address = address;
  this.privateKey = privateKey;
}

/**
* 将任务加入处理队列
* @param {Object} message - 需要处理的消息对象
*/
async enqueue(message) {
  // 使用arguments保持原始参数结构
  this.queue.push(arguments);
  // 触发队列处理（如果当前空闲）
  if (!this.isPending) {
    this.processQueue();
  }
}

/**
* 递归处理队列的核心方法
*/
async processQueue() {
  if (this.queue.length > 0 && !this.isPending) {
    this.isPending = true; // 上锁
    const message = this.queue.shift(); // 取出最早的任务

    try {
      // 执行实际业务逻辑
      await this.fun(...message);
    } catch (error) {
      logger.log("处理任务时出错:", error);
    }

    // 固定间隔15秒
    await sleep(15000);
    this.isPending = false; // 释放锁
    this.processQueue(); // 递归处理下一个任务
  }
}
}