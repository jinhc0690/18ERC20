# TestERC20 合约使用说明

## 概述

`TestERC20` 是一个功能完整的ERC20代币合约，包含铸造(mint)、销毁(burn)和转移(transfer)功能，专门用于测试和构造区块链事件。

## 合约特性

- **标准ERC20功能**: 完全兼容ERC20标准
- **铸造功能**: 所有者可以铸造新代币
- **销毁功能**: 所有者可以销毁用户代币，用户也可以自己销毁
- **批量操作**: 支持批量铸造和批量销毁
- **权限控制**: 使用OpenZeppelin的Ownable合约进行权限管理
- **事件记录**: 自定义事件记录所有操作

## 合约函数

### 核心函数

- `mint(address to, uint256 amount)`: 铸造代币给指定地址
- `burn(address from, uint256 amount)`: 销毁指定地址的代币
- `burnSelf(uint256 amount)`: 用户自己销毁代币
- `transfer(address to, uint256 amount)`: 转移代币
- `batchMint(address[] recipients, uint256[] amounts)`: 批量铸造
- `batchBurn(address[] froms, uint256[] amounts)`: 批量销毁

### 查询函数

- `name()`: 获取代币名称
- `symbol()`: 获取代币符号
- `decimals()`: 获取代币精度
- `totalSupply()`: 获取总供应量
- `balanceOf(address account)`: 获取账户余额
- `owner()`: 获取合约所有者

## 部署步骤

### 1. 编译合约

```bash
npx hardhat compile
```

### 2. 部署合约

```bash
npx hardhat run scripts/deploy_erc20.js --network <network_name>
```

部署参数:
- 代币名称: "Test Token"
- 代币符号: "TEST"
- 初始供应量: 1,000,000

### 3. 更新交互脚本

部署完成后，将合约地址更新到 `scripts/interact_erc20.js` 中的 `CONTRACT_ADDRESS` 变量。

### 4. 执行交互操作

```bash
npx hardhat run scripts/interact_erc20.js --network <network_name>
```

## 测试

运行测试套件:

```bash
npx hardhat test test/TestERC20.test.js
```

## 事件类型

### 1. TokensMinted 事件
```solidity
event TokensMinted(address indexed minter, address indexed to, uint256 amount);
```
- `minter`: 铸造者地址
- `to`: 接收地址
- `amount`: 铸造数量

### 2. TokensBurned 事件
```solidity
event TokensBurned(address indexed burner, address indexed from, uint256 amount);
```
- `burner`: 销毁者地址
- `from`: 被销毁代币的地址
- `amount`: 销毁数量

### 3. TokensTransferred 事件
```solidity
event TokensTransferred(address indexed from, address indexed to, uint256 amount);
```
- `from`: 发送地址
- `to`: 接收地址
- `amount`: 转移数量

## 使用示例

### 铸造代币
```javascript
// 铸造1000个代币给用户1
await testERC20.mint(user1.address, ethers.utils.parseEther("1000"));
```

### 销毁代币
```javascript
// 销毁用户1的100个代币
await testERC20.burn(user1.address, ethers.utils.parseEther("100"));
```

### 转移代币
```javascript
// 用户1转移200个代币给用户2
await testERC20.connect(user1).transfer(user2.address, ethers.utils.parseEther("200"));
```

### 批量操作
```javascript
// 批量铸造
const recipients = [user1.address, user2.address];
const amounts = [
  ethers.utils.parseEther("500"),
  ethers.utils.parseEther("300")
];
await testERC20.batchMint(recipients, amounts);
```

## 网络配置

确保在 `hardhat.config.js` 中配置了正确的网络参数。

## 注意事项

1. 只有合约所有者可以执行铸造和销毁操作
2. 用户只能销毁自己的代币
3. 所有操作都会发出相应的事件
4. 代币精度为18位小数
5. 初始供应量会在部署时分配给部署者

## 故障排除

如果遇到问题，请检查:
- 合约是否正确编译
- 网络配置是否正确
- 账户是否有足够的ETH支付gas费用
- 合约地址是否正确更新
