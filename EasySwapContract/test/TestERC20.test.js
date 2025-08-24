const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("TestERC20", function () {
  let TestERC20, testERC20;
  let owner, user1, user2, user3;
  let initialSupply = 1000000; // 100万代币

  beforeEach(async function () {
    // 获取账户
    [owner, user1, user2, user3] = await ethers.getSigners();

    // 部署合约
    TestERC20 = await ethers.getContractFactory("TestERC20");
    testERC20 = await TestERC20.deploy("Test Token", "TEST", initialSupply);
    await testERC20.deployed();
  });

  describe("部署", function () {
    it("应该正确设置代币名称和符号", async function () {
      expect(await testERC20.name()).to.equal("Test Token");
      expect(await testERC20.symbol()).to.equal("TEST");
    });

    it("应该正确设置初始供应量", async function () {
      const expectedSupply = ethers.utils.parseEther(initialSupply.toString());
      expect(await testERC20.totalSupply()).to.equal(expectedSupply);
    });

    it("应该将初始供应量分配给部署者", async function () {
      const expectedBalance = ethers.utils.parseEther(initialSupply.toString());
      expect(await testERC20.balanceOf(owner.address)).to.equal(expectedBalance);
    });

    it("应该正确设置所有者", async function () {
      expect(await testERC20.owner()).to.equal(owner.address);
    });
  });

  describe("铸造功能", function () {
    it("所有者应该能够铸造代币", async function () {
      const mintAmount = ethers.utils.parseEther("1000");
      const initialBalance = await testERC20.balanceOf(user1.address);
      
      await testERC20.mint(user1.address, mintAmount);
      
      const finalBalance = await testERC20.balanceOf(user1.address);
      expect(finalBalance).to.equal(initialBalance.add(mintAmount));
    });

    it("非所有者不能铸造代币", async function () {
      const mintAmount = ethers.utils.parseEther("1000");
      
      await expect(
        testERC20.connect(user1).mint(user2.address, mintAmount)
      ).to.be.revertedWithCustomError(testERC20, "OwnableUnauthorizedAccount");
    });

    it("应该发出正确的铸造事件", async function () {
      const mintAmount = ethers.utils.parseEther("1000");
      
      await expect(testERC20.mint(user1.address, mintAmount))
        .to.emit(testERC20, "TokensMinted")
        .withArgs(owner.address, user1.address, mintAmount);
    });

    it("应该能够批量铸造代币", async function () {
      const recipients = [user1.address, user2.address];
      const amounts = [
        ethers.utils.parseEther("500"),
        ethers.utils.parseEther("300")
      ];
      
      const initialBalance1 = await testERC20.balanceOf(user1.address);
      const initialBalance2 = await testERC20.balanceOf(user2.address);
      
      await testERC20.batchMint(recipients, amounts);
      
      const finalBalance1 = await testERC20.balanceOf(user1.address);
      const finalBalance2 = await testERC20.balanceOf(user2.address);
      
      expect(finalBalance1).to.equal(initialBalance1.add(amounts[0]));
      expect(finalBalance2).to.equal(initialBalance2.add(amounts[1]));
    });
  });

  describe("销毁功能", function () {
    beforeEach(async function () {
      // 先给用户铸造一些代币
      await testERC20.mint(user1.address, ethers.utils.parseEther("1000"));
      await testERC20.mint(user2.address, ethers.utils.parseEther("1000"));
    });

    it("所有者应该能够销毁用户的代币", async function () {
      const burnAmount = ethers.utils.parseEther("100");
      const initialBalance = await testERC20.balanceOf(user1.address);
      
      await testERC20.burn(user1.address, burnAmount);
      
      const finalBalance = await testERC20.balanceOf(user1.address);
      expect(finalBalance).to.equal(initialBalance.sub(burnAmount));
    });

    it("非所有者不能销毁代币", async function () {
      const burnAmount = ethers.utils.parseEther("100");
      
      await expect(
        testERC20.connect(user1).burn(user2.address, burnAmount)
      ).to.be.revertedWithCustomError(testERC20, "OwnableUnauthorizedAccount");
    });

    it("用户应该能够自己销毁代币", async function () {
      const burnAmount = ethers.utils.parseEther("100");
      const initialBalance = await testERC20.balanceOf(user1.address);
      
      await testERC20.connect(user1).burnSelf(burnAmount);
      
      const finalBalance = await testERC20.balanceOf(user1.address);
      expect(finalBalance).to.equal(initialBalance.sub(burnAmount));
    });

    it("应该发出正确的销毁事件", async function () {
      const burnAmount = ethers.utils.parseEther("100");
      
      await expect(testERC20.burn(user1.address, burnAmount))
        .to.emit(testERC20, "TokensBurned")
        .withArgs(owner.address, user1.address, burnAmount);
    });

    it("应该能够批量销毁代币", async function () {
      const burnFroms = [user1.address, user2.address];
      const burnAmounts = [
        ethers.utils.parseEther("100"),
        ethers.utils.parseEther("200")
      ];
      
      const initialBalance1 = await testERC20.balanceOf(user1.address);
      const initialBalance2 = await testERC20.balanceOf(user2.address);
      
      await testERC20.batchBurn(burnFroms, burnAmounts);
      
      const finalBalance1 = await testERC20.balanceOf(user1.address);
      const finalBalance2 = await testERC20.balanceOf(user2.address);
      
      expect(finalBalance1).to.equal(initialBalance1.sub(burnAmounts[0]));
      expect(finalBalance2).to.equal(initialBalance2.sub(burnAmounts[1]));
    });
  });

  describe("转移功能", function () {
    beforeEach(async function () {
      // 先给用户铸造一些代币
      await testERC20.mint(user1.address, ethers.utils.parseEther("1000"));
    });

    it("用户应该能够转移代币", async function () {
      const transferAmount = ethers.utils.parseEther("200");
      const initialBalance1 = await testERC20.balanceOf(user1.address);
      const initialBalance2 = await testERC20.balanceOf(user2.address);
      
      await testERC20.connect(user1).transfer(user2.address, transferAmount);
      
      const finalBalance1 = await testERC20.balanceOf(user1.address);
      const finalBalance2 = await testERC20.balanceOf(user2.address);
      
      expect(finalBalance1).to.equal(initialBalance1.sub(transferAmount));
      expect(finalBalance2).to.equal(initialBalance2.add(transferAmount));
    });

    it("应该发出正确的转移事件", async function () {
      const transferAmount = ethers.utils.parseEther("200");
      
      await expect(testERC20.connect(user1).transfer(user2.address, transferAmount))
        .to.emit(testERC20, "TokensTransferred")
        .withArgs(user1.address, user2.address, transferAmount);
    });

    it("不能转移超过余额的代币", async function () {
      const transferAmount = ethers.utils.parseEther("2000"); // 超过余额
      
      await expect(
        testERC20.connect(user1).transfer(user2.address, transferAmount)
      ).to.be.revertedWithCustomError(testERC20, "ERC20InsufficientBalance");
    });
  });

  describe("权限控制", function () {
    it("只有所有者能够铸造代币", async function () {
      const mintAmount = ethers.utils.parseEther("1000");
      
      await expect(
        testERC20.connect(user1).mint(user2.address, mintAmount)
      ).to.be.revertedWithCustomError(testERC20, "OwnableUnauthorizedAccount");
    });

    it("只有所有者能够销毁代币", async function () {
      const burnAmount = ethers.utils.parseEther("100");
      
      await expect(
        testERC20.connect(user1).burn(user2.address, burnAmount)
      ).to.be.revertedWithCustomError(testERC20, "OwnableUnauthorizedAccount");
    });

    it("只有所有者能够批量铸造代币", async function () {
      const recipients = [user1.address];
      const amounts = [ethers.utils.parseEther("100")];
      
      await expect(
        testERC20.connect(user1).batchMint(recipients, amounts)
      ).to.be.revertedWithCustomError(testERC20, "OwnableUnauthorizedAccount");
    });

    it("只有所有者能够批量销毁代币", async function () {
      const burnFroms = [user1.address];
      const burnAmounts = [ethers.utils.parseEther("100")];
      
      await expect(
        testERC20.connect(user1).batchBurn(burnFroms, burnAmounts)
      ).to.be.revertedWithCustomError(testERC20, "OwnableUnauthorizedAccount");
    });
  });

  describe("事件记录", function () {
    it("应该记录所有铸造事件", async function () {
      const mintAmount = ethers.utils.parseEther("1000");
      
      const tx = await testERC20.mint(user1.address, mintAmount);
      const receipt = await tx.wait();
      
      // OpenZeppelin v5会发出多个事件，包括Transfer和我们的自定义事件
      expect(receipt.events.length).to.be.greaterThan(0);
      const mintEvent = receipt.events.find(e => e.event === "TokensMinted");
      expect(mintEvent).to.not.be.undefined;
    });

    it("应该记录所有销毁事件", async function () {
      await testERC20.mint(user1.address, ethers.utils.parseEther("1000"));
      const burnAmount = ethers.utils.parseEther("100");
      
      const tx = await testERC20.burn(user1.address, burnAmount);
      const receipt = await tx.wait();
      
      // OpenZeppelin v5会发出多个事件，包括Transfer和我们的自定义事件
      expect(receipt.events.length).to.be.greaterThan(0);
      const burnEvent = receipt.events.find(e => e.event === "TokensBurned");
      expect(burnEvent).to.not.be.undefined;
    });

    it("应该记录所有转移事件", async function () {
      await testERC20.mint(user1.address, ethers.utils.parseEther("1000"));
      const transferAmount = ethers.utils.parseEther("200");
      
      const tx = await testERC20.connect(user1).transfer(user2.address, transferAmount);
      const receipt = await tx.wait();
      
      // OpenZeppelin v5会发出多个事件，包括Transfer和我们的自定义事件
      expect(receipt.events.length).to.be.greaterThan(0);
      const transferEvent = receipt.events.find(e => e.event === "TokensTransferred");
      expect(transferEvent).to.not.be.undefined;
    });
  });
});
