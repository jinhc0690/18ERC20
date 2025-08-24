# TestERC20 代币合约项目

## 🎯 项目概述

这是一个功能完整的ERC20代币合约项目，专门用于测试和构造区块链事件。合约包含铸造(mint)、销毁(burn)和转移(transfer)等核心功能，并会发出相应的事件记录。

## ✨ 主要特性

- **标准ERC20兼容**: 完全符合ERC20标准
- **铸造功能**: 支持单个和批量铸造代币
- **销毁功能**: 支持所有者和用户自销毁
- **转移功能**: 标准代币转移
- **权限控制**: 基于OpenZeppelin的Ownable合约
- **事件记录**: 自定义事件记录所有操作
- **批量操作**: 支持批量铸造和销毁
- **完整测试**: 包含全面的测试套件

## 📁 项目结构

```
EasySwapContract/
├── contracts/
│   └── TestERC20.sol              # 主合约文件
├── scripts/
│   ├── deploy_erc20.js            # 部署脚本
│   ├── deploy_local.js            # 本地部署脚本
│   ├── interact_erc20.js          # 交互脚本
│   ├── demo_events.js             # 事件演示脚本
│   └── quick_start.js             # 快速启动脚本
├── test/
│   └── TestERC20.test.js          # 测试文件
├── docs/
│   └── ERC20_Usage.md             # 使用说明文档
└── README_ERC20.md                # 项目说明文档
```

## 🚀 快速开始

### 1. 安装依赖

```bash
npm install
```

### 2. 编译合约

```bash
npx hardhat compile
```

### 3. 运行快速启动脚本

```bash
npx hardhat run scripts/quick_start.js
```

这个脚本会自动：
- 编译合约
- 部署到本地网络
- 测试基本功能
- 构造各种事件
- 显示结果统计

## 📋 详细使用说明

### 部署合约

```bash
# 部署到本地网络
npx hardhat run scripts/deploy_local.js

# 部署到测试网络（需要配置网络）
npx hardhat run scripts/deploy_erc20.js --network sepolia
```

### 运行测试

```bash
# 运行所有测试
npx hardhat test

# 运行特定测试文件
npx hardhat test test/TestERC20.test.js
```

### 演示事件构造

```bash
# 运行完整的事件演示
npx hardhat run scripts/demo_events.js
```

## 🔧 合约功能

### 核心函数

| 函数 | 描述 | 权限 |
|------|------|------|
| `mint(address to, uint256 amount)` | 铸造代币 | 仅所有者 |
| `burn(address from, uint256 amount)` | 销毁代币 | 仅所有者 |
| `burnSelf(uint256 amount)` | 自销毁代币 | 任何用户 |
| `transfer(address to, uint256 amount)` | 转移代币 | 任何用户 |
| `batchMint(address[] recipients, uint256[] amounts)` | 批量铸造 | 仅所有者 |
| `batchBurn(address[] froms, uint256[] amounts)` | 批量销毁 | 仅所有者 |

### 事件类型

1. **TokensMinted**: 铸造事件
   ```solidity
   event TokensMinted(address indexed minter, address indexed to, uint256 amount);
   ```

2. **TokensBurned**: 销毁事件
   ```solidity
   event TokensBurned(address indexed burner, address indexed from, uint256 amount);
   ```

3. **TokensTransferred**: 转移事件
   ```solidity
   event TokensTransferred(address indexed from, address indexed to, uint256 amount);
   ```

## 🧪 测试覆盖

测试套件覆盖以下功能：

- ✅ 合约部署和初始化
- ✅ 铸造功能（单个和批量）
- ✅ 销毁功能（所有者和自销毁）
- ✅ 转移功能
- ✅ 权限控制
- ✅ 事件发出
- ✅ 错误处理
- ✅ 边界条件

## 🌐 网络配置

项目支持以下网络：

- **localhost**: 本地开发网络
- **sepolia**: Sepolia测试网络
- **base**: Base测试网络
- **mainnet**: 以太坊主网

## 📊 事件构造统计

通过运行演示脚本，您将看到类似以下的统计：

```
🎯 === 事件构造统计 ===
✅ 成功构造了以下事件:
   🪙 铸造事件: 4 个
   🔥 销毁事件: 3 个
   🔄 转移事件: 3 个
   📊 总事件数: 10 个
```

## 🔍 查询事件

### 查询铸造事件
```javascript
const mintEvents = await testERC20.queryFilter(
  testERC20.filters.TokensMinted(),
  fromBlock,
  toBlock
);
```

### 查询销毁事件
```javascript
const burnEvents = await testERC20.queryFilter(
  testERC20.filters.TokensBurned(),
  fromBlock,
  toBlock
);
```

### 查询转移事件
```javascript
const transferEvents = await testERC20.queryFilter(
  testERC20.filters.TokensTransferred(),
  fromBlock,
  toBlock
);
```

## 🛠️ 开发工具

- **Hardhat**: 开发框架
- **OpenZeppelin**: 安全合约库
- **Ethers.js**: 以太坊交互库
- **Chai**: 测试框架
- **Solidity Coverage**: 代码覆盖率

## 📚 学习资源

- [OpenZeppelin文档](https://docs.openzeppelin.com/)
- [Hardhat文档](https://hardhat.org/docs)
- [ERC20标准](https://eips.ethereum.org/EIPS/eip-20)
- [Solidity文档](https://docs.soliditylang.org/)

## 🤝 贡献

欢迎提交Issue和Pull Request来改进这个项目！

## 📄 许可证

MIT License

## 🆘 常见问题

### Q: 如何修改代币名称和符号？
A: 在部署脚本中修改 `tokenName` 和 `tokenSymbol` 变量。

### Q: 如何增加初始供应量？
A: 在部署脚本中修改 `initialSupply` 变量。

### Q: 如何添加新的功能？
A: 在 `TestERC20.sol` 合约中添加新函数，并在测试文件中添加相应测试。

### Q: 如何部署到其他网络？
A: 在 `hardhat.config.js` 中添加网络配置，然后使用 `--network` 参数指定网络。

---

🎉 **开始使用TestERC20合约来构造您的第一个区块链事件吧！**
