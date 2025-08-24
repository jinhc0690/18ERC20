const { ethers } = require("hardhat");

async function main() {
  console.log("开始与TestERC20合约交互...");

  // 合约地址（部署后需要更新）
  const CONTRACT_ADDRESS = "0x91967508183aefF7A6ff8F01Cf3fD3b7c35CF880"; // 请替换为实际部署的合约地址
  //const CONTRACT_ADDRESS = "0x99b5e4D23F5b5b928500934f55f51fd43f3BfB3E"; // 请替换为实际部署的合约地址
  
  // 获取账户
  const [owner, user1, user2] = await ethers.getSigners();
  console.log("合约所有者:", owner.address);
  console.log("用户1:", user1.address);
  console.log("用户2:", user2.address);

  // 连接到合约
  const TestERC20 = await ethers.getContractFactory("TestERC20");
  const testERC20 = TestERC20.attach(CONTRACT_ADDRESS);

  console.log("\n=== 初始状态 ===");
  const totalSupply = await testERC20.totalSupply();
  const ownerBalance = await testERC20.balanceOf(owner.address);
  console.log("总供应量:", ethers.utils.formatUnits(totalSupply, 18));
  console.log("所有者余额:", ethers.utils.formatUnits(ownerBalance, 18));

  try {
    // 1. 铸造代币给用户1
    console.log("\n=== 1. 铸造代币给用户1 ===");
    const mintAmount1 = ethers.utils.parseEther("1000"); // 1000代币
    const mintTx1 = await testERC20.mint(user1.address, mintAmount1);
    console.log("铸造交易哈希:", mintTx1.hash);
    await mintTx1.wait();
    
    const user1Balance = await testERC20.balanceOf(user1.address);
    console.log("用户1余额:", ethers.utils.formatUnits(user1Balance, 18));

    // 2. 用户1转移代币给用户2
    console.log("\n=== 2. 用户1转移代币给用户2 ===");
    const transferAmount = ethers.utils.parseEther("200"); // 200代币
    
    // 连接用户1账户
    const testERC20User1 = testERC20.connect(user1);
    const transferTx = await testERC20User1.transfer(user2.address, transferAmount);
    console.log("转移交易哈希:", transferTx.hash);
    await transferTx.wait();
    
    const user1BalanceAfter = await testERC20.balanceOf(user1.address);
    const user2BalanceAfter = await testERC20.balanceOf(user2.address);
    console.log("转移后用户1余额:", ethers.utils.formatUnits(user1BalanceAfter, 18));
    console.log("转移后用户2余额:", ethers.utils.formatUnits(user2BalanceAfter, 18));

    // 3. 销毁用户2的代币
    console.log("\n=== 3. 销毁用户2的代币 ===");
    const burnAmount = ethers.utils.parseEther("100"); // 销毁100代币
    const burnTx = await testERC20.burn(user2.address, burnAmount);
    console.log("销毁交易哈希:", burnTx.hash);
    await burnTx.wait();
    
    const user2BalanceAfter_1 = await testERC20.balanceOf(user2.address);
    console.log("销毁后用户3余额:", ethers.utils.formatUnits(user2BalanceAfter_1, 18));

    // 4. 用户1自己销毁一些代币
    console.log("\n=== 4. 用户1自己销毁代币 ===");
    const selfBurnAmount = ethers.utils.parseEther("50"); // 销毁50代币
    
    const testERC20User1_1 = testERC20.connect(user1);
    const selfBurnTx = await testERC20User1_1.burnSelf(selfBurnAmount);
    console.log("自销毁交易哈希:", selfBurnTx.hash);
    await selfBurnTx.wait();
    
    const user1BalanceAfterBurn = await testERC20.balanceOf(user1.address);
    console.log("自销毁后用户1余额:", ethers.utils.formatUnits(user1BalanceAfterBurn, 18));

    // 5. 最终状态
    console.log("\n=== 最终状态 ===");
    const finalTotalSupply = await testERC20.totalSupply();
    const finalOwnerBalance = await testERC20.balanceOf(owner.address);
    console.log("最终总供应量:", ethers.utils.formatUnits(finalTotalSupply, 18));
    console.log("最终所有者余额:", ethers.utils.formatUnits(finalOwnerBalance, 18));

  } catch (error) {
    console.error("交互过程中发生错误:", error);
  }
}

// 如果直接运行此脚本
if (require.main === module) {
  main()
    .then(() => {
      console.log("\n交互完成！");
      process.exit(0);
    })
    .catch((error) => {
      console.error("交互失败:", error);
      process.exit(1);
    });
}

module.exports = { main };
