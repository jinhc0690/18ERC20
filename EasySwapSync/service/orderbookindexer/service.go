package orderbookindexer

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ProjectsTask/EasySwapBase/chain/chainclient"
	"github.com/ProjectsTask/EasySwapBase/chain/types"
	"github.com/ProjectsTask/EasySwapBase/logger/xzap"
	"github.com/ProjectsTask/EasySwapBase/ordermanager"
	"github.com/ProjectsTask/EasySwapBase/stores/gdb"
	"github.com/ProjectsTask/EasySwapBase/stores/gdb/orderbookmodel/base"
	"github.com/ProjectsTask/EasySwapBase/stores/gdb/orderbookmodel/multi"
	"github.com/ProjectsTask/EasySwapBase/stores/xkv"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethereumTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/threading"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/ProjectsTask/EasySwapSync/service/comm"
	"github.com/ProjectsTask/EasySwapSync/service/config"
)

const (
	EventIndexType   = 6
	SleepInterval    = 10 // in seconds
	SyncBlockPeriod  = 10
	LogMakeTopic     = "0xfc37f2ff950f95913eb7182357ba3c14df60ef354bc7d6ab1ba2815f249fffe6"
	LogCancelTopic   = "0x0ac8bb53fac566d7afc05d8b4df11d7690a7b27bdc40b54e4060f9b21fb849bd"
	LogMatchTopic    = "0xf629aecab94607bc43ce4aebd564bf6e61c7327226a797b002de724b9944b20e"
	contractAbi      = `[{"inputs":[],"name":"CannotFindNextEmptyKey","type":"error"},{"inputs":[],"name":"CannotFindPrevEmptyKey","type":"error"},{"inputs":[{"internalType":"OrderKey","name":"orderKey","type":"bytes32"}],"name":"CannotInsertDuplicateOrder","type":"error"},{"inputs":[],"name":"CannotInsertEmptyKey","type":"error"},{"inputs":[],"name":"CannotInsertExistingKey","type":"error"},{"inputs":[],"name":"CannotRemoveEmptyKey","type":"error"},{"inputs":[],"name":"CannotRemoveMissingKey","type":"error"},{"inputs":[],"name":"EnforcedPause","type":"error"},{"inputs":[],"name":"ExpectedPause","type":"error"},{"inputs":[],"name":"InvalidInitialization","type":"error"},{"inputs":[],"name":"NotInitializing","type":"error"},{"inputs":[{"internalType":"address","name":"owner","type":"address"}],"name":"OwnableInvalidOwner","type":"error"},{"inputs":[{"internalType":"address","name":"account","type":"address"}],"name":"OwnableUnauthorizedAccount","type":"error"},{"inputs":[],"name":"ReentrancyGuardReentrantCall","type":"error"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"offset","type":"uint256"},{"indexed":false,"internalType":"bytes","name":"msg","type":"bytes"}],"name":"BatchMatchInnerError","type":"event"},{"anonymous":false,"inputs":[],"name":"EIP712DomainChanged","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint64","name":"version","type":"uint64"}],"name":"Initialized","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"OrderKey","name":"orderKey","type":"bytes32"},{"indexed":true,"internalType":"address","name":"maker","type":"address"}],"name":"LogCancel","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"},{"indexed":true,"internalType":"address","name":"by","type":"address"}],"name":"TokensMinted","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"from","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"},{"indexed":true,"internalType":"address","name":"by","type":"address"}],"name":"TokensBurned","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"from","type":"address"},{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"TokensTransferred","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"OrderKey","name":"orderKey","type":"bytes32"},{"indexed":true,"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"indexed":true,"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"indexed":true,"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"indexed":false,"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"indexed":false,"internalType":"Price","name":"price","type":"uint128"},{"indexed":false,"internalType":"uint64","name":"expiry","type":"uint64"},{"indexed":false,"internalType":"uint64","name":"salt","type":"uint64"}],"name":"LogMake","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"OrderKey","name":"makeOrderKey","type":"bytes32"},{"indexed":true,"internalType":"OrderKey","name":"takeOrderKey","type":"bytes32"},{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"indexed":false,"internalType":"structLibOrder.Order","name":"makeOrder","type":"tuple"},{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"indexed":false,"internalType":"structLibOrder.Order","name":"takeOrder","type":"tuple"},{"indexed":false,"internalType":"uint128","name":"fillPrice","type":"uint128"}],"name":"LogMatch","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"OrderKey","name":"orderKey","type":"bytes32"},{"indexed":false,"internalType":"uint64","name":"salt","type":"uint64"}],"name":"LogSkipOrder","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint128","name":"newProtocolShare","type":"uint128"}],"name":"LogUpdatedProtocolShare","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"recipient","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"LogWithdrawETH","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"account","type":"address"}],"name":"Paused","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"account","type":"address"}],"name":"Unpaused","type":"event"},{"inputs":[{"internalType":"OrderKey[]","name":"orderKeys","type":"bytes32[]"}],"name":"cancelOrders","outputs":[{"internalType":"bool[]","name":"successes","type":"bool[]"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"components":[{"internalType":"OrderKey","name":"oldOrderKey","type":"bytes32"},{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"internalType":"structLibOrder.Order","name":"newOrder","type":"tuple"}],"internalType":"structLibOrder.EditDetail[]","name":"editDetails","type":"tuple[]"}],"name":"editOrders","outputs":[{"internalType":"OrderKey[]","name":"newOrderKeys","type":"bytes32[]"}],"stateMutability":"payable","type":"function"},{"inputs":[],"name":"eip712Domain","outputs":[{"internalType":"bytes1","name":"fields","type":"bytes1"},{"internalType":"string","name":"name","type":"string"},{"internalType":"string","name":"version","type":"string"},{"internalType":"uint256","name":"chainId","type":"uint256"},{"internalType":"address","name":"verifyingContract","type":"address"},{"internalType":"bytes32","name":"salt","type":"bytes32"},{"internalType":"uint256[]","name":"extensions","type":"uint256[]"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"OrderKey","name":"","type":"bytes32"}],"name":"filledAmount","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"}],"name":"getBestOrder","outputs":[{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"internalType":"structLibOrder.Order","name":"orderResult","type":"tuple"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"collection","type":"address"},{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"}],"name":"getBestPrice","outputs":[{"internalType":"Price","name":"price","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"collection","type":"address"},{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"Price","name":"price","type":"uint128"}],"name":"getNextBestPrice","outputs":[{"internalType":"Price","name":"nextBestPrice","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"uint256","name":"count","type":"uint256"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"OrderKey","name":"firstOrderKey","type":"bytes32"}],"name":"getOrders","outputs":[{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"internalType":"structLibOrder.Order[]","name":"resultOrders","type":"tuple[]"},{"internalType":"OrderKey","name":"nextOrderKey","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint128","name":"newProtocolShare","type":"uint128"},{"internalType":"address","name":"newVault","type":"address"},{"internalType":"string","name":"EIP712Name","type":"string"},{"internalType":"string","name":"EIP712Version","type":"string"}],"name":"initialize","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"internalType":"structLibOrder.Order[]","name":"newOrders","type":"tuple[]"}],"name":"makeOrders","outputs":[{"internalType":"OrderKey[]","name":"newOrderKeys","type":"bytes32[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"internalType":"structLibOrder.Order","name":"sellOrder","type":"tuple"},{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"internalType":"structLibOrder.Order","name":"buyOrder","type":"tuple"}],"name":"matchOrder","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"internalType":"structLibOrder.Order","name":"sellOrder","type":"tuple"},{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"internalType":"structLibOrder.Order","name":"buyOrder","type":"tuple"},{"internalType":"uint256","name":"msgValue","type":"uint256"}],"name":"matchOrderWithoutPayback","outputs":[{"internalType":"uint128","name":"costValue","type":"uint128"}],"stateMutability":"payable","type":"function"},{"inputs":[{"components":[{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"internalType":"structLibOrder.Order","name":"sellOrder","type":"tuple"},{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"internalType":"structLibOrder.Order","name":"buyOrder","type":"tuple"}],"internalType":"structLibOrder.MatchDetail[]","name":"matchDetails","type":"tuple[]"}],"name":"matchOrders","outputs":[{"internalType":"bool[]","name":"successes","type":"bool[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"","type":"address"},{"internalType":"enumLibOrder.Side","name":"","type":"uint8"},{"internalType":"Price","name":"","type":"uint128"}],"name":"orderQueues","outputs":[{"internalType":"OrderKey","name":"head","type":"bytes32"},{"internalType":"OrderKey","name":"tail","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"OrderKey","name":"","type":"bytes32"}],"name":"orders","outputs":[{"components":[{"internalType":"enumLibOrder.Side","name":"side","type":"uint8"},{"internalType":"enumLibOrder.SaleKind","name":"saleKind","type":"uint8"},{"internalType":"address","name":"maker","type":"address"},{"components":[{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"address","name":"collection","type":"address"},{"internalType":"uint96","name":"amount","type":"uint96"}],"internalType":"structLibOrder.Asset","name":"nft","type":"tuple"},{"internalType":"Price","name":"price","type":"uint128"},{"internalType":"uint64","name":"expiry","type":"uint64"},{"internalType":"uint64","name":"salt","type":"uint64"}],"internalType":"structLibOrder.Order","name":"order","type":"tuple"},{"internalType":"OrderKey","name":"next","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"pause","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"paused","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"","type":"address"},{"internalType":"enumLibOrder.Side","name":"","type":"uint8"}],"name":"priceTrees","outputs":[{"internalType":"Price","name":"root","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"protocolShare","outputs":[{"internalType":"uint128","name":"","type":"uint128"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"renounceOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint128","name":"newProtocolShare","type":"uint128"}],"name":"setProtocolShare","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"newVault","type":"address"}],"name":"setVault","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"unpause","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"recipient","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"withdrawETH","outputs":[],"stateMutability":"nonpayable","type":"function"},{"stateMutability":"payable","type":"receive"}]`
	FixForCollection = 0
	FixForItem       = 1
	List             = 0
	Bid              = 1

	HexPrefix   = "0x"
	ZeroAddress = "0x0000000000000000000000000000000000000000"

	TokensMinted      = "0x969cd201f68f120baff2bf3c59bc3b534434e08b69a71a14ab85cb79cd3b63e4"
	TokensBurned      = "0x08009940fb138ae33fbb70c10b643e840c71f1654344cc173975a815e117e687"
	TokensTransferred = "0x1b89874203ff7f0bba87c969ada3f32fda22ed38a6706d35199d21280c7811b1"
)

type Order struct {
	Side     uint8
	SaleKind uint8
	Maker    common.Address
	Nft      struct {
		TokenId        *big.Int
		CollectionAddr common.Address
		Amount         *big.Int
	}
	Price  *big.Int
	Expiry uint64
	Salt   uint64
}

type Service struct {
	ctx          context.Context
	cfg          *config.Config
	db           *gorm.DB
	kv           *xkv.Store
	orderManager *ordermanager.OrderManager
	chainClient  chainclient.ChainClient
	chainId      int64
	chain        string
	parsedAbi    abi.ABI
}

var MultiChainMaxBlockDifference = map[string]uint64{
	"eth":        1,
	"optimism":   2,
	"starknet":   1,
	"arbitrum":   2,
	"base":       2,
	"zksync-era": 2,
	"sepolia":    6,
	"basepolia":  6,
}

func New(ctx context.Context, cfg *config.Config, db *gorm.DB, xkv *xkv.Store, chainClient chainclient.ChainClient, chainId int64, chain string, orderManager *ordermanager.OrderManager) *Service {
	parsedAbi, _ := abi.JSON(strings.NewReader(contractAbi)) // 通过ABI实例化
	return &Service{
		ctx:          ctx,
		cfg:          cfg,
		db:           db,
		kv:           xkv,
		chainClient:  chainClient,
		orderManager: orderManager,
		chain:        chain,
		chainId:      chainId,
		parsedAbi:    parsedAbi,
	}
}

func (s *Service) Start() {
	//threading.GoSafe(s.SyncOrderBookEventLoop)
	//threading.GoSafe(s.UpKeepingCollectionFloorChangeLoop)
	threading.GoSafe(s.SyncErc20EventLoop)
	threading.GoSafe(s.SyncIntegralLoop)
}

func (s *Service) SyncOrderBookEventLoop() {
	var indexedStatus base.IndexedStatus
	if err := s.db.WithContext(s.ctx).Table(base.IndexedStatusTableName()).
		Where("chain_id = ? and index_type = ?", s.chainId, EventIndexType).
		First(&indexedStatus).Error; err != nil {
		xzap.WithContext(s.ctx).Error("failed on get listing index status",
			zap.Error(err))
		return
	}

	lastSyncBlock := uint64(indexedStatus.LastIndexedBlock)
	for {
		select {
		case <-s.ctx.Done():
			xzap.WithContext(s.ctx).Info("SyncOrderBookEventLoop stopped due to context cancellation")
			return
		default:
		}

		currentBlockNum, err := s.chainClient.BlockNumber() // 以轮询的方式获取当前区块高度
		if err != nil {
			xzap.WithContext(s.ctx).Error("failed on get current block number", zap.Error(err))
			time.Sleep(SleepInterval * time.Second)
			continue
		}

		if lastSyncBlock > currentBlockNum-MultiChainMaxBlockDifference[s.chain] { // 如果上次同步的区块高度大于当前区块高度，等待一段时间后再次轮询
			time.Sleep(SleepInterval * time.Second)
			continue
		}

		startBlock := lastSyncBlock
		endBlock := startBlock + SyncBlockPeriod
		if endBlock > currentBlockNum-MultiChainMaxBlockDifference[s.chain] { // 如果结束区块高度大于当前区块高度，将结束区块高度设置为当前区块高度
			endBlock = currentBlockNum - MultiChainMaxBlockDifference[s.chain]
		}

		query := types.FilterQuery{
			FromBlock: new(big.Int).SetUint64(startBlock),
			ToBlock:   new(big.Int).SetUint64(endBlock),
			Addresses: []string{s.cfg.ContractCfg.DexAddress},
		}

		logs, err := s.chainClient.FilterLogs(s.ctx, query) //同时获取多个（SyncBlockPeriod）区块的日志
		if err != nil {
			xzap.WithContext(s.ctx).Error("failed on get log", zap.Error(err))
			time.Sleep(SleepInterval * time.Second)
			continue
		}

		for _, log := range logs { // 遍历日志，根据不同的topic处理不同的事件
			ethLog := log.(ethereumTypes.Log)
			switch ethLog.Topics[0].String() {
			case LogMakeTopic:
				s.handleMakeEvent(ethLog)
			case LogCancelTopic:
				s.handleCancelEvent(ethLog)
			case LogMatchTopic:
				s.handleMatchEvent(ethLog)
			default:
			}
		}

		lastSyncBlock = endBlock + 1 // 更新最后同步的区块高度
		if err := s.db.WithContext(s.ctx).Table(base.IndexedStatusTableName()).
			Where("chain_id = ? and index_type = ?", s.chainId, EventIndexType).
			Update("last_indexed_block", lastSyncBlock).Error; err != nil {
			xzap.WithContext(s.ctx).Error("failed on update orderbook event sync block number",
				zap.Error(err))
			return
		}

		xzap.WithContext(s.ctx).Info("sync orderbook event ...",
			zap.Uint64("start_block", startBlock),
			zap.Uint64("end_block", endBlock))
	}
}

// 处理挂单事件
func (s *Service) handleMakeEvent(log ethereumTypes.Log) {
	var event struct {
		OrderKey [32]byte
		Nft      struct {
			TokenId        *big.Int
			CollectionAddr common.Address
			Amount         *big.Int
		}
		Price  *big.Int
		Expiry uint64
		Salt   uint64
	}

	// Unpack data
	err := s.parsedAbi.UnpackIntoInterface(&event, "LogMake", log.Data) // 通过ABI解析日志数据
	if err != nil {
		xzap.WithContext(s.ctx).Error("Error unpacking LogMake event:", zap.Error(err))
		return
	}
	// Extract indexed fields from topics
	side := uint8(new(big.Int).SetBytes(log.Topics[1].Bytes()).Uint64())
	saleKind := uint8(new(big.Int).SetBytes(log.Topics[2].Bytes()).Uint64())
	maker := common.BytesToAddress(log.Topics[3].Bytes())

	var orderType int64
	if side == Bid { // 买单
		if saleKind == FixForCollection { // 针对集合的买单
			orderType = multi.CollectionBidOrder
		} else { // 针对某个具体NFT的买单
			orderType = multi.ItemBidOrder
		}
	} else { // 卖单
		orderType = multi.ListingOrder
	}
	newOrder := multi.Order{
		CollectionAddress: event.Nft.CollectionAddr.String(),
		MarketplaceId:     multi.MarketOrderBook,
		TokenId:           event.Nft.TokenId.String(),
		OrderID:           HexPrefix + hex.EncodeToString(event.OrderKey[:]),
		OrderStatus:       multi.OrderStatusActive,
		EventTime:         time.Now().Unix(),
		ExpireTime:        int64(event.Expiry),
		CurrencyAddress:   s.cfg.ContractCfg.EthAddress,
		Price:             decimal.NewFromBigInt(event.Price, 0),
		Maker:             maker.String(),
		Taker:             ZeroAddress,
		QuantityRemaining: event.Nft.Amount.Int64(),
		Size:              event.Nft.Amount.Int64(),
		OrderType:         orderType,
		Salt:              int64(event.Salt),
	}
	if err := s.db.WithContext(s.ctx).Table(multi.OrderTableName(s.chain)).Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&newOrder).Error; err != nil { // 将订单信息存入数据库
		xzap.WithContext(s.ctx).Error("failed on create order",
			zap.Error(err))
	}
	blockTime, err := s.chainClient.BlockTimeByNumber(s.ctx, big.NewInt(int64(log.BlockNumber)))
	if err != nil {
		xzap.WithContext(s.ctx).Error("failed to get block time", zap.Error(err))
		return
	}
	var activityType int
	if side == Bid {
		if saleKind == FixForCollection {
			activityType = multi.CollectionBid
		} else {
			activityType = multi.ItemBid
		}
	} else {
		activityType = multi.Listing
	}
	newActivity := multi.Activity{ // 将订单信息存入活动表
		ActivityType:      activityType,
		Maker:             maker.String(),
		Taker:             ZeroAddress,
		MarketplaceID:     multi.MarketOrderBook,
		CollectionAddress: event.Nft.CollectionAddr.String(),
		TokenId:           event.Nft.TokenId.String(),
		CurrencyAddress:   s.cfg.ContractCfg.EthAddress,
		Price:             decimal.NewFromBigInt(event.Price, 0),
		BlockNumber:       int64(log.BlockNumber),
		TxHash:            log.TxHash.String(),
		EventTime:         int64(blockTime),
	}
	if err := s.db.WithContext(s.ctx).Table(multi.ActivityTableName(s.chain)).Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&newActivity).Error; err != nil {
		xzap.WithContext(s.ctx).Warn("failed on create activity",
			zap.Error(err))
	}

	if err := s.orderManager.AddToOrderManagerQueue(&multi.Order{ // 将订单信息存入订单管理队列
		ExpireTime:        newOrder.ExpireTime,
		OrderID:           newOrder.OrderID,
		CollectionAddress: newOrder.CollectionAddress,
		TokenId:           newOrder.TokenId,
		Price:             newOrder.Price,
		Maker:             newOrder.Maker,
	}); err != nil {
		xzap.WithContext(s.ctx).Error("failed on add order to manager queue",
			zap.Error(err),
			zap.String("order_id", newOrder.OrderID))
	}
}

func (s *Service) handleMatchEvent(log ethereumTypes.Log) {
	var event struct {
		MakeOrder Order
		TakeOrder Order
		FillPrice *big.Int
	}

	err := s.parsedAbi.UnpackIntoInterface(&event, "LogMatch", log.Data)
	if err != nil {
		xzap.WithContext(s.ctx).Error("Error unpacking LogMatch event:", zap.Error(err))
		return
	}

	makeOrderId := HexPrefix + hex.EncodeToString(log.Topics[1].Bytes()) // 通过topic获取订单ID
	takeOrderId := HexPrefix + hex.EncodeToString(log.Topics[2].Bytes())
	var owner string
	var collection string
	var tokenId string
	var from string
	var to string
	var sellOrderId string
	var buyOrder multi.Order
	if event.MakeOrder.Side == Bid { // 买单， 由卖方发起交易撮合
		owner = strings.ToLower(event.MakeOrder.Maker.String())
		collection = event.TakeOrder.Nft.CollectionAddr.String()
		tokenId = event.TakeOrder.Nft.TokenId.String()
		from = event.TakeOrder.Maker.String()
		to = event.MakeOrder.Maker.String()
		sellOrderId = takeOrderId

		// 更新卖方订单状态
		if err := s.db.WithContext(s.ctx).Table(multi.OrderTableName(s.chain)).
			Where("order_id = ?", takeOrderId).
			Updates(map[string]interface{}{
				"order_status":       multi.OrderStatusFilled,
				"quantity_remaining": 0,
				"taker":              to,
			}).Error; err != nil {
			xzap.WithContext(s.ctx).Error("failed on update order status",
				zap.String("order_id", takeOrderId))
			return
		}

		// 查询买方订单信息，不存在则无需更新，说明不是从平台前端发起的交易
		if err := s.db.WithContext(s.ctx).Table(multi.OrderTableName(s.chain)).
			Where("order_id = ?", makeOrderId).
			First(&buyOrder).Error; err != nil {
			xzap.WithContext(s.ctx).Error("failed on get buy order",
				zap.Error(err))
			return
		}
		// 更新买方订单的剩余数量
		if buyOrder.QuantityRemaining > 1 {
			if err := s.db.WithContext(s.ctx).Table(multi.OrderTableName(s.chain)).
				Where("order_id = ?", makeOrderId).
				Update("quantity_remaining", buyOrder.QuantityRemaining-1).Error; err != nil {
				xzap.WithContext(s.ctx).Error("failed on update order quantity_remaining",
					zap.String("order_id", makeOrderId))
				return
			}
		} else {
			if err := s.db.WithContext(s.ctx).Table(multi.OrderTableName(s.chain)).
				Where("order_id = ?", makeOrderId).
				Updates(map[string]interface{}{
					"order_status":       multi.OrderStatusFilled,
					"quantity_remaining": 0,
				}).Error; err != nil {
				xzap.WithContext(s.ctx).Error("failed on update order status",
					zap.String("order_id", makeOrderId))
				return
			}
		}
	} else { // 卖单， 由买方发起交易撮合， 同理
		owner = strings.ToLower(event.TakeOrder.Maker.String())
		collection = event.MakeOrder.Nft.CollectionAddr.String()
		tokenId = event.MakeOrder.Nft.TokenId.String()
		from = event.MakeOrder.Maker.String()
		to = event.TakeOrder.Maker.String()
		sellOrderId = makeOrderId

		if err := s.db.WithContext(s.ctx).Table(multi.OrderTableName(s.chain)).
			Where("order_id = ?", makeOrderId).
			Updates(map[string]interface{}{
				"order_status":       multi.OrderStatusFilled,
				"quantity_remaining": 0,
				"taker":              to,
			}).Error; err != nil {
			xzap.WithContext(s.ctx).Error("failed on update order status",
				zap.String("order_id", makeOrderId))
			return
		}

		if err := s.db.WithContext(s.ctx).Table(multi.OrderTableName(s.chain)).
			Where("order_id = ?", takeOrderId).
			First(&buyOrder).Error; err != nil {
			xzap.WithContext(s.ctx).Error("failed on get buy order",
				zap.Error(err))
			return
		}
		if buyOrder.QuantityRemaining > 1 {
			if err := s.db.WithContext(s.ctx).Table(multi.OrderTableName(s.chain)).
				Where("order_id = ?", takeOrderId).
				Update("quantity_remaining", buyOrder.QuantityRemaining-1).Error; err != nil {
				xzap.WithContext(s.ctx).Error("failed on update order quantity_remaining",
					zap.String("order_id", takeOrderId))
				return
			}
		} else {
			if err := s.db.WithContext(s.ctx).Table(multi.OrderTableName(s.chain)).
				Where("order_id = ?", takeOrderId).
				Updates(map[string]interface{}{
					"order_status":       multi.OrderStatusFilled,
					"quantity_remaining": 0,
				}).Error; err != nil {
				xzap.WithContext(s.ctx).Error("failed on update order status",
					zap.String("order_id", takeOrderId))
				return
			}
		}
	}

	blockTime, err := s.chainClient.BlockTimeByNumber(s.ctx, big.NewInt(int64(log.BlockNumber)))
	if err != nil {
		xzap.WithContext(s.ctx).Error("failed to get block time", zap.Error(err))
		return
	}
	newActivity := multi.Activity{
		ActivityType:      multi.Sale,
		Maker:             event.MakeOrder.Maker.String(),
		Taker:             event.TakeOrder.Maker.String(),
		MarketplaceID:     multi.MarketOrderBook,
		CollectionAddress: collection,
		TokenId:           tokenId,
		CurrencyAddress:   s.cfg.ContractCfg.EthAddress,
		Price:             decimal.NewFromBigInt(event.FillPrice, 0),
		BlockNumber:       int64(log.BlockNumber),
		TxHash:            log.TxHash.String(),
		EventTime:         int64(blockTime),
	}
	if err := s.db.WithContext(s.ctx).Table(multi.ActivityTableName(s.chain)).Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&newActivity).Error; err != nil {
		xzap.WithContext(s.ctx).Warn("failed on create activity",
			zap.Error(err))
	}

	// 更新NFT的所有者
	if err := s.db.WithContext(s.ctx).Table(multi.ItemTableName(s.chain)).
		Where("collection_address = ? and token_id = ?", strings.ToLower(collection), tokenId).
		Update("owner", owner).Error; err != nil {
		xzap.WithContext(s.ctx).Error("failed to update item owner",
			zap.Error(err))
		return
	}

	if err := ordermanager.AddUpdatePriceEvent(s.kv, &ordermanager.TradeEvent{ // 将交易信息存入价格更新队列
		OrderId:        sellOrderId,
		CollectionAddr: collection,
		EventType:      ordermanager.Buy,
		TokenID:        tokenId,
		From:           from,
		To:             to,
	}, s.chain); err != nil {
		xzap.WithContext(s.ctx).Error("failed on add update price event",
			zap.Error(err),
			zap.String("type", "sale"),
			zap.String("order_id", sellOrderId))
	}
}

func (s *Service) handleCancelEvent(log ethereumTypes.Log) {
	orderId := HexPrefix + hex.EncodeToString(log.Topics[1].Bytes())
	//maker := common.BytesToAddress(log.Topics[2].Bytes())
	if err := s.db.WithContext(s.ctx).Table(multi.OrderTableName(s.chain)).
		Where("order_id = ?", orderId).
		Update("order_status", multi.OrderStatusCancelled).Error; err != nil {
		xzap.WithContext(s.ctx).Error("failed on update order status",
			zap.String("order_id", orderId))
		return
	}

	var cancelOrder multi.Order
	if err := s.db.WithContext(s.ctx).Table(multi.OrderTableName(s.chain)).
		Where("order_id = ?", orderId).
		First(&cancelOrder).Error; err != nil {
		xzap.WithContext(s.ctx).Error("failed on get cancel order",
			zap.Error(err))
		return
	}

	blockTime, err := s.chainClient.BlockTimeByNumber(s.ctx, big.NewInt(int64(log.BlockNumber)))
	if err != nil {
		xzap.WithContext(s.ctx).Error("failed to get block time", zap.Error(err))
		return
	}
	var activityType int
	if cancelOrder.OrderType == multi.ListingOrder {
		activityType = multi.CancelListing
	} else if cancelOrder.OrderType == multi.CollectionBidOrder {
		activityType = multi.CancelCollectionBid
	} else {
		activityType = multi.CancelItemBid
	}
	newActivity := multi.Activity{
		ActivityType:      activityType,
		Maker:             cancelOrder.Maker,
		Taker:             ZeroAddress,
		MarketplaceID:     multi.MarketOrderBook,
		CollectionAddress: cancelOrder.CollectionAddress,
		TokenId:           cancelOrder.TokenId,
		CurrencyAddress:   s.cfg.ContractCfg.EthAddress,
		Price:             cancelOrder.Price,
		BlockNumber:       int64(log.BlockNumber),
		TxHash:            log.TxHash.String(),
		EventTime:         int64(blockTime),
	}
	if err := s.db.WithContext(s.ctx).Table(multi.ActivityTableName(s.chain)).Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&newActivity).Error; err != nil {
		xzap.WithContext(s.ctx).Warn("failed on create activity",
			zap.Error(err))
	}

	if err := ordermanager.AddUpdatePriceEvent(s.kv, &ordermanager.TradeEvent{
		OrderId:        cancelOrder.OrderID,
		CollectionAddr: cancelOrder.CollectionAddress,
		TokenID:        cancelOrder.TokenId,
		EventType:      ordermanager.Cancel,
	}, s.chain); err != nil {
		xzap.WithContext(s.ctx).Error("failed on add update price event",
			zap.Error(err),
			zap.String("type", "cancel"),
			zap.String("order_id", cancelOrder.OrderID))
	}
}

func (s *Service) UpKeepingCollectionFloorChangeLoop() {
	timer := time.NewTicker(comm.DaySeconds * time.Second)
	defer timer.Stop()
	updateFloorPriceTimer := time.NewTicker(comm.MaxCollectionFloorTimeDifference * time.Second)
	defer updateFloorPriceTimer.Stop()

	var indexedStatus base.IndexedStatus
	if err := s.db.WithContext(s.ctx).Table(base.IndexedStatusTableName()).
		Select("last_indexed_time").
		Where("chain_id = ? and index_type = ?", s.chainId, comm.CollectionFloorChangeIndexType).
		First(&indexedStatus).Error; err != nil {
		xzap.WithContext(s.ctx).Error("failed on get collection floor change index status",
			zap.Error(err))
		return
	}

	for {
		select {
		case <-s.ctx.Done():
			xzap.WithContext(s.ctx).Info("UpKeepingCollectionFloorChangeLoop stopped due to context cancellation")
			return
		case <-timer.C:
			if err := s.deleteExpireCollectionFloorChangeFromDatabase(); err != nil {
				xzap.WithContext(s.ctx).Error("failed on delete expire collection floor change",
					zap.Error(err))
			}
		case <-updateFloorPriceTimer.C:
			if s.cfg.ProjectCfg.Name == gdb.OrderBookDexProject {
				floorPrices, err := s.QueryCollectionsFloorPrice()
				if err != nil {
					xzap.WithContext(s.ctx).Error("failed on query collections floor change",
						zap.Error(err))
					continue
				}

				if err := s.persistCollectionsFloorChange(floorPrices); err != nil {
					xzap.WithContext(s.ctx).Error("failed on persist collections floor price",
						zap.Error(err))
					continue
				}
			}
		default:
		}
	}
}

func (s *Service) deleteExpireCollectionFloorChangeFromDatabase() error {
	stmt := fmt.Sprintf(`DELETE FROM %s where event_time < UNIX_TIMESTAMP() - %d`, gdb.GetMultiProjectCollectionFloorPriceTableName(s.cfg.ProjectCfg.Name, s.chain), comm.CollectionFloorTimeRange)

	if err := s.db.Exec(stmt).Error; err != nil {
		return errors.Wrap(err, "failed on delete expire collection floor price")
	}

	return nil
}

func (s *Service) QueryCollectionsFloorPrice() ([]multi.CollectionFloorPrice, error) {
	timestamp := time.Now().Unix()
	timestampMilli := time.Now().UnixMilli()
	var collectionFloorPrice []multi.CollectionFloorPrice
	sql := fmt.Sprintf(`SELECT co.collection_address as collection_address,min(co.price) as price
FROM %s as ci
         left join %s co on co.collection_address = ci.collection_address and co.token_id = ci.token_id
WHERE (co.order_type = ? and
       co.order_status = ? and expire_time > ? and co.maker = ci.owner) group by co.collection_address`, gdb.GetMultiProjectItemTableName(s.cfg.ProjectCfg.Name, s.chain), gdb.GetMultiProjectOrderTableName(s.cfg.ProjectCfg.Name, s.chain))
	if err := s.db.WithContext(s.ctx).Raw(
		sql,
		multi.ListingType,
		multi.OrderStatusActive,
		time.Now().Unix(),
	).Scan(&collectionFloorPrice).Error; err != nil {
		return nil, errors.Wrap(err, "failed on get collection floor price")
	}

	for i := 0; i < len(collectionFloorPrice); i++ {
		collectionFloorPrice[i].EventTime = timestamp
		collectionFloorPrice[i].CreateTime = timestampMilli
		collectionFloorPrice[i].UpdateTime = timestampMilli
	}

	return collectionFloorPrice, nil
}

func (s *Service) persistCollectionsFloorChange(FloorPrices []multi.CollectionFloorPrice) error {
	for i := 0; i < len(FloorPrices); i += comm.DBBatchSizeLimit {
		end := i + comm.DBBatchSizeLimit
		if i+comm.DBBatchSizeLimit >= len(FloorPrices) {
			end = len(FloorPrices)
		}

		valueStrings := make([]string, 0)
		valueArgs := make([]interface{}, 0)

		for _, t := range FloorPrices[i:end] {
			valueStrings = append(valueStrings, "(?,?,?,?,?)")
			valueArgs = append(valueArgs, t.CollectionAddress, t.Price, t.EventTime, t.CreateTime, t.UpdateTime)
		}

		stmt := fmt.Sprintf(`INSERT INTO %s (collection_address,price,event_time,create_time,update_time)  VALUES %s
		ON DUPLICATE KEY UPDATE update_time=VALUES(update_time)`, gdb.GetMultiProjectCollectionFloorPriceTableName(s.cfg.ProjectCfg.Name, s.chain), strings.Join(valueStrings, ","))

		if err := s.db.Exec(stmt, valueArgs...).Error; err != nil {
			return errors.Wrap(err, "failed on persist collection floor price info")
		}
	}
	return nil
}

func (s *Service) SyncErc20EventLoop() {
	var indexedStatus base.IndexedStatus
	if err := s.db.WithContext(s.ctx).Table(base.IndexedStatusTableName()).
		Where("chain_id = ? and index_type = ?", s.chainId, EventIndexType).
		First(&indexedStatus).Error; err != nil {
		xzap.WithContext(s.ctx).Error("failed on get listing index status",
			zap.Error(err))
		return
	}

	lastSyncBlock := uint64(indexedStatus.LastIndexedBlock)
	for {
		select {
		case <-s.ctx.Done():
			xzap.WithContext(s.ctx).Info("SyncErc20EventLoop stopped due to context cancellation")
			return
		default:
		}

		currentBlockNum, err := s.chainClient.BlockNumber() // 以轮询的方式获取当前区块高度
		if err != nil {
			xzap.WithContext(s.ctx).Error("failed on get current block number", zap.Error(err))
			time.Sleep(SleepInterval * time.Second)
			continue
		}

		if lastSyncBlock > currentBlockNum-MultiChainMaxBlockDifference[s.chain] { // 如果上次同步的区块高度大于当前区块高度，等待一段时间后再次轮询
			time.Sleep(SleepInterval * time.Second)
			continue
		}

		startBlock := lastSyncBlock
		endBlock := startBlock + SyncBlockPeriod
		if endBlock > currentBlockNum-MultiChainMaxBlockDifference[s.chain] { // 如果结束区块高度大于当前区块高度，将结束区块高度设置为当前区块高度
			endBlock = currentBlockNum - MultiChainMaxBlockDifference[s.chain]
		}

		query := types.FilterQuery{
			FromBlock: new(big.Int).SetUint64(startBlock),
			ToBlock:   new(big.Int).SetUint64(endBlock),
			Addresses: []string{s.cfg.ContractCfg.DexAddress},
		}

		logs, err := s.chainClient.FilterLogs(s.ctx, query) //同时获取多个（SyncBlockPeriod）区块的日志
		if err != nil {
			xzap.WithContext(s.ctx).Error("failed on get log", zap.Error(err))
			time.Sleep(SleepInterval * time.Second)
			continue
		}

		for _, log := range logs { // 遍历日志，根据不同的topic处理不同的事件
			ethLog := log.(ethereumTypes.Log)
			xzap.WithContext(s.ctx).Info("=================================Topics=================")
			xzap.WithContext(s.ctx).Info(ethLog.Topics[0].String())
			switch ethLog.Topics[0].String() {
			case TokensMinted:
				xzap.WithContext(s.ctx).Info("=================================TokensMinted=================")
				xzap.WithContext(s.ctx).Info(ethLog.BlockHash.String())
				s.handleMintedEvent(ethLog)
			case TokensBurned:
				xzap.WithContext(s.ctx).Info("=================================TokensBurned=================")
				xzap.WithContext(s.ctx).Info(ethLog.BlockHash.String())
				s.handleBurnedEvent(ethLog)
			case TokensTransferred:
				xzap.WithContext(s.ctx).Info("=================================TokensTransferred=================")
				xzap.WithContext(s.ctx).Info(ethLog.BlockHash.String())
				s.handleTransferredEvent(ethLog)
			default:
			}
		}

		lastSyncBlock = endBlock + 1 // 更新最后同步的区块高度
		if err := s.db.WithContext(s.ctx).Table(base.IndexedStatusTableName()).
			Where("chain_id = ? and index_type = ?", s.chainId, EventIndexType).
			Update("last_indexed_block", lastSyncBlock).Error; err != nil {
			xzap.WithContext(s.ctx).Error("failed on update erc20 event sync block number",
				zap.Error(err))
			return
		}

		xzap.WithContext(s.ctx).Info("sync erc20 event ...",
			zap.Uint64("start_block", startBlock),
			zap.Uint64("end_block", endBlock))
	}
}

func BalanceTableName(chainName string) string {
	return fmt.Sprintf("erc_balance_%s", chainName)
}

type Balance struct {
	ID              int64  `gorm:"column:id" json:"id"` //  主键
	ChainId         int64  `gorm:"column:chain_id" json:"chain_id"`
	Owner           string `gorm:"column:owner" json:"owner"`
	Quantity        int64  `gorm:"column:quantity" json:"quantity"`
	ChangeTime      int64  `gorm:"column:change_time" json:"change_time"`
	EventType       string `gorm:"column:event_type" json:"event_type"`
	CreateTime      int64  `json:"create_time" gorm:"column:create_time;type:bigint(20);autoCreateTime:milli;comment:创建时间"` // 创建时间
	UpdateTime      int64  `json:"update_time" gorm:"column:update_time;type:bigint(20);autoUpdateTime:milli;comment:更新时间"` // 更新时间
	WhetherIntegral string `gorm:"column:whether_integral" json:"whether_integral"`
}

func BalanceSumTableName(chainName string) string {
	return fmt.Sprintf("erc_balance_sum_%s", chainName)
}

type BalanceSum struct {
	ChainId    int64  `gorm:"column:chain_id" json:"chain_id"`
	Owner      string `gorm:"column:owner" json:"owner"`
	Quantity   int64  `gorm:"column:quantity" json:"quantity"`
	ChangeTime int64  `gorm:"column:change_time" json:"change_time"`
	CreateTime int64  `json:"create_time" gorm:"column:create_time;type:bigint(20);autoCreateTime:milli;comment:创建时间"` // 创建时间
	UpdateTime int64  `json:"update_time" gorm:"column:update_time;type:bigint(20);autoUpdateTime:milli;comment:更新时间"` // 更新时间
}

// 处理铸币事件
func (s *Service) handleMintedEvent(log ethereumTypes.Log) {
	var event struct {
		To     common.Address
		Amount *big.Int
		By     common.Address
	}
	//fmt.Println("log.Data:==========================================", log.Data)
	//fmt.Println("event.Amount:==========================================", new(big.Int).SetBytes(log.Data[:32]))
	//fmt.Println("s.parsedAbi.Methods:==========================================", s.parsedAbi.Methods)
	//fmt.Println("s.parsedAbi.Events:==========================================", s.parsedAbi.Events)
	// Unpack data
	err := s.parsedAbi.UnpackIntoInterface(&event, "TokensMinted", log.Data) // 通过ABI解析日志数据
	if err != nil {
		xzap.WithContext(s.ctx).Error("Error unpacking TokensMinted event:", zap.Error(err))
		return
	}
	fmt.Println("Topics:==========================================", log.Topics)
	fmt.Println("event:==========================================", event)
	fmt.Println("To:==========================================", event.To)
	fmt.Println("Amount:==========================================", event.Amount)
	fmt.Println("By:==========================================", event.By)
	// Extract indexed fields from topics
	owner := common.BytesToAddress(log.Topics[1].Bytes())
	fmt.Println("owner:==========================================", owner)

	blockTime, err := s.chainClient.BlockTimeByNumber(s.ctx, big.NewInt(int64(log.BlockNumber)))
	if err != nil {
		xzap.WithContext(s.ctx).Error("failed to get block time", zap.Error(err))
		return
	}
	fmt.Println("blockTime:==========================================", blockTime)

	divisor := new(big.Int)
	divisor.SetString("1000000000000000000", 10)

	// 执行除法运算
	result := new(big.Int)
	result.Div(event.Amount, divisor)

	balance := Balance{
		ChainId:         s.chainId,
		Owner:           owner.String(),
		Quantity:        result.Int64(),
		ChangeTime:      int64(blockTime),
		EventType:       "Mint",
		WhetherIntegral: "N",
	}
	if err := s.db.WithContext(s.ctx).Table(BalanceTableName(s.chain)).Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&balance).Error; err != nil { // 将余额信息存入数据库
		xzap.WithContext(s.ctx).Error("failed on create balance",
			zap.Error(err))
	}

	balanceSum := BalanceSum{
		ChainId:    s.chainId,
		Owner:      owner.String(),
		Quantity:   result.Int64(),
		ChangeTime: int64(blockTime),
	}
	if err := s.db.WithContext(s.ctx).Debug().Table(BalanceSumTableName(s.chain)).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "owner"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"quantity": gorm.Expr("quantity + VALUES(quantity)"), "change_time": gorm.Expr("VALUES(change_time)")}),
	}).Create(&balanceSum).Error; err != nil { // 将总余额信息存入数据库
		xzap.WithContext(s.ctx).Error("failed on create balance_sum",
			zap.Error(err))
	}
}

// 处理销毁事件
func (s *Service) handleBurnedEvent(log ethereumTypes.Log) {
	var event struct {
		From   common.Address
		Amount *big.Int
		By     common.Address
	}
	//fmt.Println("log.Data:==========================================", log.Data)
	//fmt.Println("event.Amount:==========================================", new(big.Int).SetBytes(log.Data[:32]))
	//fmt.Println("s.parsedAbi.Methods:==========================================", s.parsedAbi.Methods)
	//fmt.Println("s.parsedAbi.Events:==========================================", s.parsedAbi.Events)
	// Unpack data
	err := s.parsedAbi.UnpackIntoInterface(&event, "TokensBurned", log.Data) // 通过ABI解析日志数据
	if err != nil {
		xzap.WithContext(s.ctx).Error("Error unpacking TokensBurned event:", zap.Error(err))
		return
	}
	fmt.Println("Topics:==========================================", log.Topics)
	fmt.Println("event:==========================================", event)
	fmt.Println("From:==========================================", event.From)
	fmt.Println("Amount:==========================================", event.Amount)
	fmt.Println("By:==========================================", event.By)
	// Extract indexed fields from topics
	owner := common.BytesToAddress(log.Topics[1].Bytes())
	fmt.Println("owner:==========================================", owner)

	blockTime, err := s.chainClient.BlockTimeByNumber(s.ctx, big.NewInt(int64(log.BlockNumber)))
	if err != nil {
		xzap.WithContext(s.ctx).Error("failed to get block time", zap.Error(err))
		return
	}
	fmt.Println("blockTime:==========================================", blockTime)

	divisor := new(big.Int)
	divisor.SetString("1000000000000000000", 10)

	// 执行除法运算
	result := new(big.Int)
	result.Div(event.Amount, divisor)

	balance := Balance{
		ChainId:         s.chainId,
		Owner:           owner.String(),
		Quantity:        -result.Int64(),
		ChangeTime:      int64(blockTime),
		EventType:       "Burn",
		WhetherIntegral: "N",
	}
	if err := s.db.WithContext(s.ctx).Table(BalanceTableName(s.chain)).Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&balance).Error; err != nil { // 将余额信息存入数据库
		xzap.WithContext(s.ctx).Error("failed on create balance",
			zap.Error(err))
	}

	balanceSum := BalanceSum{
		ChainId:    s.chainId,
		Owner:      owner.String(),
		Quantity:   -result.Int64(),
		ChangeTime: int64(blockTime),
	}
	if err := s.db.WithContext(s.ctx).Debug().Table(BalanceSumTableName(s.chain)).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "owner"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"quantity": gorm.Expr("quantity + VALUES(quantity)"), "change_time": gorm.Expr("VALUES(change_time)")}),
	}).Create(&balanceSum).Error; err != nil { // 将总余额信息存入数据库
		xzap.WithContext(s.ctx).Error("failed on create balance_sum",
			zap.Error(err))
	}
}

// 处理转移事件
func (s *Service) handleTransferredEvent(log ethereumTypes.Log) {
	var event struct {
		From   common.Address
		To     common.Address
		Amount *big.Int
	}
	//fmt.Println("log.Data:==========================================", log.Data)
	//fmt.Println("event.Amount:==========================================", new(big.Int).SetBytes(log.Data[:32]))
	//fmt.Println("s.parsedAbi.Methods:==========================================", s.parsedAbi.Methods)
	//fmt.Println("s.parsedAbi.Events:==========================================", s.parsedAbi.Events)
	// Unpack data
	err := s.parsedAbi.UnpackIntoInterface(&event, "TokensTransferred", log.Data) // 通过ABI解析日志数据
	if err != nil {
		xzap.WithContext(s.ctx).Error("Error unpacking TokensTransferred event:", zap.Error(err))
		return
	}
	fmt.Println("Topics:==========================================", log.Topics)
	fmt.Println("event:==========================================", event)
	fmt.Println("From:==========================================", event.From)
	fmt.Println("To:==========================================", event.To)
	fmt.Println("Amount:==========================================", event.Amount)
	// Extract indexed fields from topics
	owner := common.BytesToAddress(log.Topics[1].Bytes())
	fmt.Println("owner:==========================================", owner)
	receiver := common.BytesToAddress(log.Topics[2].Bytes())
	fmt.Println("receiver:==========================================", receiver)

	blockTime, err := s.chainClient.BlockTimeByNumber(s.ctx, big.NewInt(int64(log.BlockNumber)))
	if err != nil {
		xzap.WithContext(s.ctx).Error("failed to get block time", zap.Error(err))
		return
	}
	fmt.Println("blockTime:==========================================", blockTime)

	divisor := new(big.Int)
	divisor.SetString("1000000000000000000", 10)

	// 执行除法运算
	result := new(big.Int)
	result.Div(event.Amount, divisor)

	balance := []Balance{
		{ChainId: s.chainId,
			Owner:           owner.String(),
			Quantity:        -result.Int64(),
			ChangeTime:      int64(blockTime),
			EventType:       "Transfer",
			WhetherIntegral: "N"},
		{
			ChainId:         s.chainId,
			Owner:           receiver.String(),
			Quantity:        result.Int64(),
			ChangeTime:      int64(blockTime),
			EventType:       "Transfer",
			WhetherIntegral: "N",
		},
	}
	if err := s.db.Debug().WithContext(s.ctx).Table(BalanceTableName(s.chain)).Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&balance).Error; err != nil { // 将余额信息存入数据库
		xzap.WithContext(s.ctx).Error("failed on create balanceOwner",
			zap.Error(err))
	}

	balanceSum := []BalanceSum{
		{ChainId: s.chainId,
			Owner:      owner.String(),
			Quantity:   -result.Int64(),
			ChangeTime: int64(blockTime)},
		{
			ChainId:    s.chainId,
			Owner:      receiver.String(),
			Quantity:   result.Int64(),
			ChangeTime: int64(blockTime),
		},
	}
	if err := s.db.WithContext(s.ctx).Debug().Table(BalanceSumTableName(s.chain)).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "owner"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"quantity": gorm.Expr("quantity + VALUES(quantity)"), "change_time": gorm.Expr("VALUES(change_time)")}),
	}).Create(&balanceSum).Error; err != nil { // 将总余额信息存入数据库
		xzap.WithContext(s.ctx).Error("failed on create balance_sum",
			zap.Error(err))
	}
}

func IntegralSumTableName(chainName string) string {
	return fmt.Sprintf("erc_integral_sum_%s", chainName)
}

type IntegralSum struct {
	ChainId      int64           `gorm:"column:chain_id" json:"chain_id"`
	Owner        string          `gorm:"column:owner" json:"owner"`
	Integral     decimal.Decimal `gorm:"column:integral" json:"integral"`
	DeadlineTime int64           `gorm:"column:deadline_time" json:"deadline_time"`
	CreateTime   int64           `json:"create_time" gorm:"column:create_time;type:bigint(20);autoCreateTime:milli;comment:创建时间"` // 创建时间
	UpdateTime   int64           `json:"update_time" gorm:"column:update_time;type:bigint(20);autoUpdateTime:milli;comment:更新时间"` // 更新时间
}

// GetNextHourTimestamp 获取给定时间戳的下一个整点时间戳
// deadlineTime: Unix时间戳（秒）
// loc: 时区，如果为nil则使用UTC时区
func GetNextHourTimestamp(deadlineTime int64, loc *time.Location) int64 {
	// 将时间戳转换为Time对象（UTC时间）
	t := time.Unix(deadlineTime, 0)

	// 转换为指定时区
	if loc != nil {
		t = t.In(loc)
	}

	// 计算下一个整点：当前时间 truncate到小时，然后增加1小时
	nextHour := t.Truncate(time.Hour).Add(time.Hour)

	// 返回Unix时间戳
	return nextHour.Unix()
}

// GetMinutesDifference 计算两个UNIX时间戳之间的分钟差（t2 - t1）
func GetMinutesDifference(t1, t2 int64) int64 {
	// 将时间戳转换为time.Time
	time1 := time.Unix(t1, 0).Truncate(time.Minute)
	time2 := time.Unix(t2, 0).Truncate(time.Minute)

	// 计算时间差并转换为分钟
	duration := time2.Sub(time1)
	minutes := int64(duration.Minutes())

	// 如果时间差是负数，需要特殊处理
	if minutes < 0 && duration.Seconds() > -60 {
		// 处理不足一分钟但跨越分钟边界的情况
		return -1
	}

	return minutes
}

func Int64SliceToString(slice []int64, sep string) string {
	// 将每个int64元素转换为字符串
	strSlice := make([]string, len(slice))
	for i, v := range slice {
		strSlice[i] = strconv.FormatInt(v, 10)
	}
	// 使用分隔符连接字符串切片
	return strings.Join(strSlice, sep)
}

// 从2025-08-21 19:46:12获取2025-08-21 19:00:00
func TruncateToHour(timestamp int64) int64 {
	t := time.Unix(timestamp-1, 0).UTC()
	truncated := t.Truncate(time.Hour)
	return truncated.Unix()
}

func (s *Service) SyncIntegralLoop() {
	for {
		select {
		case <-s.ctx.Done():
			xzap.WithContext(s.ctx).Info("SyncIntegralLoop stopped due to context cancellation")
			return
		default:
		}

		var integralSum []IntegralSum
		if err := s.db.WithContext(s.ctx).Table(IntegralSumTableName(s.chain)).
			Find(&integralSum).Error; err != nil {
			xzap.WithContext(s.ctx).Error("failed on get listing integral_sum",
				zap.Error(err))
			return
		}

		var balance []Balance
		if err := s.db.WithContext(s.ctx).Table(BalanceTableName(s.chain)).
			Select("owner, min(change_time) as change_time").
			Where("whether_integral = ?", "N").
			Group("owner").
			Find(&balance).Error; err != nil {
			xzap.WithContext(s.ctx).Error("failed on get listing balance",
				zap.Error(err))
			return
		}

		if integralSum == nil || len(integralSum) == 0 {
			if balance == nil || len(balance) == 0 {
				return
			}
			for _, balanceEach := range balance {
				owner := balanceEach.Owner
				changeTime := balanceEach.ChangeTime

				// 使用系统本地时区（如CST）
				localNextHour := GetNextHourTimestamp(changeTime, time.Local)
				fmt.Printf("Local next hour: %d (%s)\n",
					localNextHour,
					time.Unix(localNextHour, 0).In(time.Local).Format("2006-01-02 15:04:05"))

				var balanceOwner []Balance
				if err := s.db.WithContext(s.ctx).Table(BalanceTableName(s.chain)).
					Where("change_time >= ? and change_time <= ? and owner = ?", changeTime, localNextHour, owner).
					Order("change_time asc").
					Find(&balanceOwner).Error; err != nil {
					xzap.WithContext(s.ctx).Error("failed on get listing balance",
						zap.Error(err))
					return
				}

				var (
					quantityLast   int64
					changeTimeLast int64
					integral       decimal.Decimal
				)
				ids := []int64{}
				for index, balanceOwnerEach := range balanceOwner {
					id := balanceOwnerEach.ID
					changeTime := balanceOwnerEach.ChangeTime
					var balanceSum Balance
					if err := s.db.WithContext(s.ctx).Table(BalanceTableName(s.chain)).
						Select("sum(quantity) quantity").
						Where("change_time <= ? and owner = ?", changeTime, owner).
						Find(&balanceSum).Error; err != nil {
						xzap.WithContext(s.ctx).Error("failed on get listing balance",
							zap.Error(err))
						return
					}
					quantity := balanceSum.Quantity
					if index != 0 {
						diffMin := GetMinutesDifference(changeTimeLast, changeTime)
						integral = integral.Add(
							decimal.NewFromInt(quantityLast).
								Mul(decimal.NewFromInt(diffMin)))
					}
					if index == len(balanceOwner)-1 {
						diffMin := GetMinutesDifference(changeTime, localNextHour)
						integral = integral.Add(
							decimal.NewFromInt(quantity).
								Mul(decimal.NewFromInt(diffMin)))
					}
					quantityLast = quantity
					changeTimeLast = balanceOwnerEach.ChangeTime

					ids = append(ids, id)
				}

				integralSum := IntegralSum{
					ChainId: s.chainId,
					Owner:   owner,
					Integral: integral.
						Mul(decimal.NewFromFloat(0.05)).
						Div(decimal.NewFromInt(60)).Round(2),
					DeadlineTime: localNextHour,
				}

				if err := s.db.WithContext(s.ctx).Table(IntegralSumTableName(s.chain)).
					Create(&integralSum).Error; err != nil { // 将总积分信息存入数据库
					xzap.WithContext(s.ctx).Error("failed on create integral_sum",
						zap.Error(err))
					return
				}

				if err := s.db.WithContext(s.ctx).Table(BalanceTableName(s.chain)).
					Where("id in ?", ids).
					Update("whether_integral", "Y").Error; err != nil { // 将余额信息更新数据库
					xzap.WithContext(s.ctx).Error("failed on update balance",
						zap.String("ids", Int64SliceToString(ids, ",")))
					return
				}

				xzap.WithContext(s.ctx).Info("sync integral ...",
					zap.Int64("start_time", changeTime),
					zap.Int64("end_time", localNextHour))

			}
		} else {
			for _, integralSumEach := range integralSum {
				owner := integralSumEach.Owner
				deadlineTime := integralSumEach.DeadlineTime
				// 使用系统本地时区（如CST）
				localNextHour := GetNextHourTimestamp(deadlineTime, time.Local)
				fmt.Printf("Local next hour: %d (%s)\n",
					localNextHour,
					time.Unix(localNextHour, 0).In(time.Local).Format("2006-01-02 15:04:05"))

				whetherIntegral := false

				// 可能存在因为异常好几天没更新的账户
				if balance != nil && len(balance) > 0 {
					for _, balanceEach := range balance {
						if balanceEach.Owner == owner {
							changeTime := balanceEach.ChangeTime
							if changeTime < deadlineTime {
								deadlineTime = TruncateToHour(changeTime)
								whetherIntegral = true
							}
							break
						}

					}
				}

				var balanceOwner []Balance
				if whetherIntegral {
					if err := s.db.WithContext(s.ctx).Table(BalanceTableName(s.chain)).
						Where("change_time <= ? and owner = ?", localNextHour, owner).
						Order("change_time asc").
						Find(&balanceOwner).Error; err != nil {
						xzap.WithContext(s.ctx).Error("failed on get listing balance",
							zap.Error(err))
						return
					}
				} else {
					if err := s.db.WithContext(s.ctx).Table(BalanceTableName(s.chain)).
						Where("change_time >= ? and change_time <= ? and owner = ?", deadlineTime, localNextHour, owner).
						Order("change_time asc").
						Find(&balanceOwner).Error; err != nil {
						xzap.WithContext(s.ctx).Error("failed on get listing balance",
							zap.Error(err))
						return
					}
				}

				var (
					quantityLast   int64
					changeTimeLast int64
					integral       decimal.Decimal
				)

				// 账户后续一直没有交易
				if balanceOwner == nil || len(balanceOwner) == 0 {
					var balanceSum Balance
					if err := s.db.WithContext(s.ctx).Table(BalanceTableName(s.chain)).
						Select("sum(quantity) quantity").
						Where("change_time <= ? and owner = ?", deadlineTime, owner).
						Find(&balanceSum).Error; err != nil {
						xzap.WithContext(s.ctx).Error("failed on get listing balance",
							zap.Error(err))
						return
					}

					integral = decimal.NewFromInt(balanceSum.Quantity).
						Mul(decimal.NewFromInt(GetMinutesDifference(deadlineTime, localNextHour)))
				}
				ids := []int64{}
				for index, balanceOwnerEach := range balanceOwner {
					id := balanceOwnerEach.ID
					changeTime := balanceOwnerEach.ChangeTime
					var balanceSum Balance
					if err := s.db.WithContext(s.ctx).Table(BalanceTableName(s.chain)).
						Select("sum(quantity) quantity").
						Where("change_time <= ? and owner = ?", changeTime, owner).
						Find(&balanceSum).Error; err != nil {
						xzap.WithContext(s.ctx).Error("failed on get listing balance",
							zap.Error(err))
						return
					}
					quantity := balanceSum.Quantity
					if index != 0 {
						diffMin := GetMinutesDifference(changeTimeLast, changeTime)
						integral = integral.Add(
							decimal.NewFromInt(quantityLast).
								Mul(decimal.NewFromInt(diffMin)))
					}
					if index == len(balanceOwner)-1 {
						diffMin := GetMinutesDifference(changeTime, localNextHour)
						integral = integral.Add(
							decimal.NewFromInt(quantity).
								Mul(decimal.NewFromInt(diffMin)))
					}
					quantityLast = quantity
					changeTimeLast = balanceOwnerEach.ChangeTime

					ids = append(ids, id)
				}

				integralSum := IntegralSum{
					Owner: owner,
					Integral: integral.
						Mul(decimal.NewFromFloat(0.05)).
						Div(decimal.NewFromInt(60)).Round(2),
					DeadlineTime: localNextHour,
				}

				if whetherIntegral {
					if err := s.db.WithContext(s.ctx).Debug().Table(IntegralSumTableName(s.chain)).Clauses(clause.OnConflict{
						Columns:   []clause.Column{{Name: "owner"}},
						DoUpdates: clause.Assignments(map[string]interface{}{"integral": gorm.Expr("VALUES(integral)"), "deadline_time": gorm.Expr("VALUES(deadline_time)")}),
					}).Create(&integralSum).Error; err != nil { // 将总积分信息存入数据库
						xzap.WithContext(s.ctx).Error("failed on create integral_sum",
							zap.Error(err))
						return
					}
				} else {
					if err := s.db.WithContext(s.ctx).Debug().Table(IntegralSumTableName(s.chain)).Clauses(clause.OnConflict{
						Columns:   []clause.Column{{Name: "owner"}},
						DoUpdates: clause.Assignments(map[string]interface{}{"integral": gorm.Expr("integral + VALUES(integral)"), "deadline_time": gorm.Expr("VALUES(deadline_time)")}),
					}).Create(&integralSum).Error; err != nil { // 将总积分信息存入数据库
						xzap.WithContext(s.ctx).Error("failed on create integral_sum",
							zap.Error(err))
						return
					}
				}

				if ids != nil && len(ids) > 0 {
					if err := s.db.WithContext(s.ctx).Table(BalanceTableName(s.chain)).
						Where("id in ?", ids).
						Update("whether_integral", "Y").Error; err != nil { // 将余额信息更新数据库
						xzap.WithContext(s.ctx).Error("failed on update balance",
							zap.String("ids", Int64SliceToString(ids, ",")))
						return
					}
				}

				xzap.WithContext(s.ctx).Info("sync integral ...",
					zap.Int64("start_time", deadlineTime),
					zap.Int64("end_time", localNextHour))

			}

			// 处理可能存在第一次进行交易的账户
			if balance != nil && len(balance) > 0 {
				for _, balanceEach := range balance {
					owner := balanceEach.Owner
					changeTime := balanceEach.ChangeTime

					whetherFirst := true
					for _, integralSumEach := range integralSum {
						if integralSumEach.Owner == owner {
							whetherFirst = false
							break
						}
					}

					if whetherFirst {
						// 使用系统本地时区（如CST）
						localNextHour := GetNextHourTimestamp(changeTime, time.Local)
						fmt.Printf("Local next hour: %d (%s)\n",
							localNextHour,
							time.Unix(localNextHour, 0).In(time.Local).Format("2006-01-02 15:04:05"))

						var balanceOwner []Balance
						if err := s.db.WithContext(s.ctx).Table(BalanceTableName(s.chain)).
							Where("change_time >= ? and change_time <= ? and owner = ?", changeTime, localNextHour, owner).
							Order("change_time asc").
							Find(&balanceOwner).Error; err != nil {
							xzap.WithContext(s.ctx).Error("failed on get listing balance",
								zap.Error(err))
							return
						}

						var (
							quantityLast   int64
							changeTimeLast int64
							integral       decimal.Decimal
						)
						ids := []int64{}
						for index, balanceOwnerEach := range balanceOwner {
							id := balanceOwnerEach.ID
							changeTime := balanceOwnerEach.ChangeTime
							var balanceSum Balance
							if err := s.db.WithContext(s.ctx).Table(BalanceTableName(s.chain)).
								Select("sum(quantity) quantity").
								Where("change_time <= ? and owner = ?", changeTime, owner).
								Find(&balanceSum).Error; err != nil {
								xzap.WithContext(s.ctx).Error("failed on get listing balance",
									zap.Error(err))
								return
							}
							quantity := balanceSum.Quantity
							if index != 0 {
								diffMin := GetMinutesDifference(changeTimeLast, changeTime)
								integral = integral.Add(
									decimal.NewFromInt(quantityLast).
										Mul(decimal.NewFromInt(diffMin)))
							}
							if index == len(balanceOwner)-1 {
								diffMin := GetMinutesDifference(changeTime, localNextHour)
								integral = integral.Add(
									decimal.NewFromInt(quantity).
										Mul(decimal.NewFromInt(diffMin)))
							}
							quantityLast = quantity
							changeTimeLast = balanceOwnerEach.ChangeTime

							ids = append(ids, id)
						}

						integralSum := IntegralSum{
							ChainId: s.chainId,
							Owner:   owner,
							Integral: integral.
								Mul(decimal.NewFromFloat(0.05)).
								Div(decimal.NewFromInt(60)).Round(2),
							DeadlineTime: localNextHour,
						}

						if err := s.db.WithContext(s.ctx).Table(IntegralSumTableName(s.chain)).
							Create(&integralSum).Error; err != nil { // 将总积分信息存入数据库
							xzap.WithContext(s.ctx).Error("failed on create integral_sum",
								zap.Error(err))
							return
						}

						if err := s.db.WithContext(s.ctx).Table(BalanceTableName(s.chain)).
							Where("id in ?", ids).
							Update("whether_integral", "Y").Error; err != nil { // 将余额信息更新数据库
							xzap.WithContext(s.ctx).Error("failed on update balance",
								zap.String("ids", Int64SliceToString(ids, ",")))
							return
						}

						xzap.WithContext(s.ctx).Info("sync integral ...",
							zap.Int64("start_time", changeTime),
							zap.Int64("end_time", localNextHour))
					}
				}
			}
		}
		time.Sleep(3600000 * time.Millisecond)
	}
}
