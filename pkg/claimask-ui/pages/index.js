import { useState, useEffect } from "react";
import { useWeb3React } from "@web3-react/core";
import { injected } from "../components/wallet/connectors";
import axios from "axios";

// 日志工具函数
function log(message, data = null) {
  const timestamp = new Date().toISOString();
  console.log(`[${timestamp}] ${message}`, data || "");
}

export default function Home() {
  const { active, account, activate, deactivate } = useWeb3React();
  const [prizes, setPrizes] = useState(null);

  // 创建一个新的axios实例
  const axiosInstance = axios.create({
    baseURL: 'http://127.0.0.1:8880',
    headers: {
      'User-Agent': 'Apifox/1.0.0 (https://apifox.com)',
      'Content-Type': 'application/json',
      'Accept': '*/*',
      'Host': '127.0.0.1:8870',
      'Connection': 'keep-alive',
    },
  });

  // 连接钱包的异步函数
  async function connect() {
    console.log("Connect function called"); // 测试日志
    try {
      await activate(injected, undefined, true);
      localStorage.setItem('isWalletConnected', true);
      log("Wallet connected successfully", { active, account });
    } catch (ex) {
      log("Wallet connection failed", ex);
    }
  }

  // 断开钱包连接的异步函数
  async function disconnect() {
    console.log("Disconnect function called"); // 测试日志
    try {
      deactivate();
      localStorage.setItem('isWalletConnected', false);
      log("Wallet disconnected successfully");
    } catch (ex) {
      log("Wallet disconnection failed", ex);
    }
  }

  // 在页面加载时尝试连接钱包的效果钩子
  useEffect(() => {
    console.log("useEffect triggered"); // 测试日志
    const connectWalletOnPageLoad = async () => {
      if (localStorage?.getItem('isWalletConnected') === 'true') {
        log("Attempting to connect wallet on page load...");
        try {
          await activate(injected);
          localStorage.setItem('isWalletConnected', true);
          log("Wallet connected on page load", { active, account });
        } catch (ex) {
          log("Wallet connection on page load failed", ex);
        }
      }
    };
    connectWalletOnPageLoad();
  }, [activate]);

  // 新增的claim函数
  async function claim() {
    console.log("Claim function called"); // 测试日志
    try {
      const response = await axiosInstance.post('/claim', { address: account });
      log("Claim request successful", response.data);
    } catch (ex) {
      log("Claim request failed", ex);
    }
  }

  // 新增的query函数
  async function query() {
    console.log("Query function called"); // 测试日志
    try {
      const response = await axiosInstance.get('/query');
      setPrizes(response.data.prizes);
      log("Query request successful", response.data);
    } catch (ex) {
      log("Query request failed", ex);
    }
  }

  // 在页面加载时尝试查询query接口的效果钩子
  useEffect(() => {
    console.log("useEffect for query triggered"); // 测试日志
    const queryOnPageLoad = async () => {
      try {
        const response = await axiosInstance.get('/query');
        setPrizes(response.data.prizes);
        log("Query on page load successful", response.data);
      } catch (ex) {
        log("Query on page load failed", ex);
      }
    };
    queryOnPageLoad();
  }, []);

  return (
      <div className="flex flex-col items-center justify-center">
        <button
            onClick={connect}
            className="py-2 mt-20 mb-4 text-lg font-bold text-white rounded-lg w-56 bg-blue-600 hover:bg-blue-800"
        >
          Connect to MetaMask
        </button>
        {active ? (
            <span>
          Connected with <b>{account}</b>
        </span>
        ) : (
            <span>Not connected</span>
        )}
        <button
            onClick={disconnect}
            className="py-2 mt-20 mb-4 text-lg font-bold text-white rounded-lg w-56 bg-blue-600 hover:bg-blue-800"
        >
          Disconnect
        </button>
        {active && (
            <button
                onClick={query}
                className="py-2 mt-20 mb-4 text-lg font-bold text-white rounded-lg w-56 bg-blue-600 hover:bg-blue-800"
            >
              Query: {prizes != null ? prizes : 'Loading...'}
            </button>
        )}
        {active && (
            <button
                onClick={claim}
                className="py-2 mt-20 mb-4 text-lg font-bold text-white rounded-lg w-56 bg-blue-600 hover:bg-blue-800"
            >
              QualClaim
            </button>
        )}
      </div>
  );
}