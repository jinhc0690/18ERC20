// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

/**
 * @title TestERC20
 * @dev 一个用于测试的ERC20代币合约，包含mint和burn功能
 */
contract TestERC20 is ERC20, Ownable {
    
    /**
     * @dev 构造函数，设置代币名称和符号
     * @param name 代币名称
     * @param symbol 代币符号
     * @param initialSupply 初始供应量
     */
    constructor(
        string memory name,
        string memory symbol,
        uint256 initialSupply
    ) ERC20(name, symbol) Ownable(msg.sender) {
        _mint(msg.sender, initialSupply * 10**decimals());
    }
    
    /**
     * @dev 铸造新代币（仅限所有者）
     * @param to 接收地址
     * @param amount 铸造数量
     */
    function mint(address to, uint256 amount) public onlyOwner {
        _mint(to, amount);
        emit TokensMinted(to, amount, owner());
    }
    
    /**
     * @dev 销毁代币（仅限所有者）
     * @param from 销毁地址
     * @param amount 销毁数量
     */
    function burn(address from, uint256 amount) public onlyOwner {
        _burn(from, amount);
        emit TokensBurned(from, amount, owner());
    }
    
    /**
     * @dev 用户自己销毁代币
     * @param amount 销毁数量
     */
    function burnSelf(uint256 amount) public {
        _burn(msg.sender, amount);
        emit TokensBurned(msg.sender, amount, msg.sender);
    }
    
    // 自定义事件
    event TokensMinted(address indexed to, uint256 amount, address indexed by);
    event TokensBurned(address indexed from, uint256 amount, address indexed by);
    event TokensTransferred(address indexed from, address indexed to, uint256 amount);
    
    // 重写transfer函数以发出自定义事件
    function transfer(address to, uint256 amount) public override returns (bool) {
        bool success = super.transfer(to, amount);
        if (success) {
            emit TokensTransferred(msg.sender, to, amount);
        }
        return success;
    }
    
    // 重写transferFrom函数以发出自定义事件
    function transferFrom(address from, address to, uint256 amount) public override returns (bool) {
        bool success = super.transferFrom(from, to, amount);
        if (success) {
            emit TokensTransferred(from, to, amount);
        }
        return success;
    }
}
