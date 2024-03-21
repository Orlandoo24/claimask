import { useWeb3React } from "@web3-react/core";
import { useEffect } from "react";
import { injected } from "../components/wallet/connectors";
import axios from 'axios'; // 引入axios库以发送HTTP请求

export default function Home() {
  const { active, account, library, connector, activate, deactivate } = useWeb3React();

  // 创建一个新的axios实例，并设置CORS相关的头部信息
  const axiosInstance = axios.create({
    baseURL: 'http://127.0.0.1:8870',
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

  return (
    <div className="flex flex-col items-center justify-center">
      <button onClick={connect} className="py-2 mt-20 mb-4 text-lg font-bold text-white rounded-lg w-56 bg-blue-600 hover:bg-blue-800">Connect to MetaMask</button>
      {active ? <span>Connected with <b>{account}</b></span> : <span>Not connected</span>}
      <button onClick={disconnect} className="py-2 mt-20 mb-4 text-lg font-bold text-white rounded-lg w-56 bg-blue-600 hover:bg-blue-800">Disconnect</button>
      {/* 新增的Claim按钮，只在钱包连接时显示 */}
      {active && <button onClick={claim} className="py-2 mt-20 mb-4 text-lg font-bold text-white rounded-lg w-56 bg-blue-600 hover:bg-blue-800">Claim</button>}
    </div>
  );
}
