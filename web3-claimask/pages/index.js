import { useState, useEffect } from "react";
import { useWeb3React } from "@web3-react/core";
import { injected } from "../components/wallet/connectors";
import axios from 'axios'; // 引入axios库以发送HTTP请求

export default function Home() {
  const { active, account, library, connector, activate, deactivate } = useWeb3React();
  const [prizes, setPrizes] = useState(null); // 用于保存prizes数量的状态

  // 创建一个新的axios实例，并设置CORS相关的头部信息
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
    try {
      await activate(injected);
      localStorage.setItem('isWalletConnected', true);
    } catch (ex) {
      console.log(ex);
    }
  }

  // 断开钱包连接的异步函数
  async function disconnect() {
    try {
      deactivate();
      localStorage.setItem('isWalletConnected', false);
    } catch (ex) {
      console.log(ex);
    }
  }

  // 在页面加载时尝试连接钱包的效果钩子
  useEffect(() => {
    const connectWalletOnPageLoad = async () => {
      if (localStorage?.getItem('isWalletConnected') === 'true') {
        try {
          await activate(injected);
          localStorage.setItem('isWalletConnected', true);
        } catch (ex) {
          console.log(ex);
        }
      }
    };
    connectWalletOnPageLoad();
  }, []);

  // 新增的claim函数，用于在已连接钱包的状态下请求claim接口
  async function claim() {
    try {
      // 使用axios实例发送POST请求到claim接口，并带上参数address
      const response = await axiosInstance.post('/claim', { address: account });
      console.log(response.data);
    } catch (ex) {
      console.log(ex);
    }
  }

  // 新增的query函数，用于发送GET请求到query接口，并更新prizes状态
  async function query() {
    try {
      // 使用axios实例发送GET请求到query接口
      const response = await axiosInstance.get('/query');
      setPrizes(response.data.prizes); // 更新prizes状态
      console.log(response.data);
    } catch (ex) {
      console.log(ex);
    }
  }

  // 在页面加载时尝试查询query接口的效果钩子
  useEffect(() => {
    const queryOnPageLoad = async () => {
      try {
        // 使用axios实例发送GET请求到query接口，并更新prizes状态
        const response = await axiosInstance.get('/query');
        setPrizes(response.data.prizes); // 更新prizes状态
        console.log(response.data);
      } catch (ex) {
        console.log(ex);
      }
    };
    queryOnPageLoad();
  }, []);

  return (
    <div className="flex flex-col items-center justify-center">   
      <button onClick={connect} className="py-2 mt-20 mb-4 text-lg font-bold text-white rounded-lg w-56 bg-blue-600 hover:bg-blue-800">Connect to MetaMask</button>
      {active ? <span> Connected with <b>{account}</b></span> : <span>Not connected</span>}
      <button onClick={disconnect} className="py-2 mt-20 mb-4 text-lg font-bold text-white rounded-lg w-56 bg-blue-600 hover:bg-blue-800">Disconnect</button>
      {active && (
        <button onClick={query} className="py-2 mt-20 mb-4 text-lg font-bold text-white rounded-lg w-56 bg-blue-600 hover:bg-blue-800">
          Query: {prizes != null ? prizes : 'Loading...'} {/* 显示prizes数量或加载状态 */}
        </button>
      )} 
      {/* 新增的Claim按钮，只在钱包连接时显示 */}
      {active && <button onClick={claim} className="py-2 mt-20 mb-4 text-lg font-bold text-white rounded-lg w-56 bg-blue-600 hover:bg-blue-800">QualClaim</button>}
    </div>
  );
}
