const { ethers } = require("hardhat");

async function main() {
  console.log("开始部署TestERC20合约...");

  // 获取部署账户
  const [deployer] = await ethers.getSigners();
  console.log("部署账户地址:", deployer.address);
  console.log("账户余额:", ethers.utils.formatEther(await deployer.getBalance()));

  // 部署合约参数
  const tokenName = "Test Token";
  const tokenSymbol = "TEST";
  const initialSupply = 1000000; // 100万代币

  console.log(`部署参数: 名称=${tokenName}, 符号=${tokenSymbol}, 初始供应量=${initialSupply}`);

  // 部署合约
  const TestERC20 = await ethers.getContractFactory("TestERC20");
  const testERC20 = await TestERC20.deploy(tokenName, tokenSymbol, initialSupply);
  await testERC20.deployed();

  console.log("TestERC20合约已部署到:", testERC20.address);
  console.log("交易哈希:", testERC20.deployTransaction.hash);

  // 等待几个区块确认
  console.log("等待区块确认...");
  await testERC20.deployTransaction.wait(5);

  // 验证部署
  const name = await testERC20.name();
  const symbol = await testERC20.symbol();
  const totalSupply = await testERC20.totalSupply();
  const owner = await testERC20.owner();
  const deployerBalance = await testERC20.balanceOf(deployer.address);

  console.log("\n=== 部署验证 ===");
  console.log("代币名称:", name);
  console.log("代币符号:", symbol);
  console.log("总供应量:", ethers.utils.formatUnits(totalSupply, 18));
  console.log("合约所有者:", owner);
  console.log("部署者余额:", ethers.utils.formatUnits(deployerBalance, 18));

  // 保存部署信息
  const deploymentInfo = {
    contractName: "TestERC20",
    contractAddress: testERC20.address,
    deployer: deployer.address,
    tokenName: tokenName,
    tokenSymbol: tokenSymbol,
    initialSupply: initialSupply,
    totalSupply: ethers.utils.formatUnits(totalSupply, 18),
    deploymentTime: new Date().toISOString(),
    network: (await ethers.provider.getNetwork()).name,
    chainId: (await ethers.provider.getNetwork()).chainId
  };

  console.log("\n=== 部署信息 ===");
  console.log(JSON.stringify(deploymentInfo, null, 2));

  return { testERC20, deploymentInfo };
}

// 如果直接运行此脚本
if (require.main === module) {
  main()
    .then(() => {
      console.log("\n部署完成！");
      process.exit(0);
    })
    .catch((error) => {
      console.error("部署失败:", error);
      process.exit(1);
    });
}

module.exports = { main };
