package preprocess

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// TODO: increase code coverage with unit tests

type smartContractResults struct {
	*basePreProcess
	chRcvAllScrs                 chan bool
	onRequestSmartContractResult func(shardID uint32, txHashes [][]byte)
	scrForBlock                  txsForBlock
	scrPool                      dataRetriever.ShardedDataCacherNotifier
	storage                      dataRetriever.StorageService
	scrProcessor                 process.SmartContractResultProcessor
	accounts                     state.AccountsAdapter
}

// NewSmartContractResultPreprocessor creates a new smartContractResult preprocessor object
func NewSmartContractResultPreprocessor(
	scrDataPool dataRetriever.ShardedDataCacherNotifier,
	store dataRetriever.StorageService,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	scrProcessor process.SmartContractResultProcessor,
	shardCoordinator sharding.Coordinator,
	accounts state.AccountsAdapter,
	onRequestSmartContractResult func(shardID uint32, txHashes [][]byte),
) (*smartContractResults, error) {

	if hasher == nil {
		return nil, process.ErrNilHasher
	}
	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}
	if scrDataPool == nil {
		return nil, process.ErrNilUTxDataPool
	}
	if store == nil {
		return nil, process.ErrNilUTxStorage
	}
	if scrProcessor == nil {
		return nil, process.ErrNilTxProcessor
	}
	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}
	if accounts == nil {
		return nil, process.ErrNilAccountsAdapter
	}
	if onRequestSmartContractResult == nil {
		return nil, process.ErrNilRequestHandler
	}

	bpp := &basePreProcess{
		hasher:           hasher,
		marshalizer:      marshalizer,
		shardCoordinator: shardCoordinator,
	}

	scr := &smartContractResults{
		basePreProcess:               bpp,
		storage:                      store,
		scrPool:                      scrDataPool,
		onRequestSmartContractResult: onRequestSmartContractResult,
		scrProcessor:                 scrProcessor,
		accounts:                     accounts,
	}

	scr.chRcvAllScrs = make(chan bool)
	scr.scrPool.RegisterHandler(scr.receivedSmartContractResult)
	scr.scrForBlock.txHashAndInfo = make(map[string]*txInfo)

	return scr, nil
}

// waitForScrHashes waits for a call whether all the requested smartContractResults appeared
func (scr *smartContractResults) waitForScrHashes(waitTime time.Duration) error {
	select {
	case <-scr.chRcvAllScrs:
		return nil
	case <-time.After(waitTime):
		return process.ErrTimeIsOut
	}
}

// IsDataPrepared returns non error if all the requested smartContractResults arrived and were saved into the pool
func (scr *smartContractResults) IsDataPrepared(requestedScrs int, haveTime func() time.Duration) error {
	if requestedScrs > 0 {
		log.Info(fmt.Sprintf("requested %d missing scr\n", requestedScrs))
		err := scr.waitForScrHashes(haveTime())
		scr.scrForBlock.mutTxsForBlock.RLock()
		missingScrs := scr.scrForBlock.missingTxs
		scr.scrForBlock.mutTxsForBlock.RUnlock()
		log.Info(fmt.Sprintf("received %d missing scr\n", requestedScrs-missingScrs))
		if err != nil {
			return err
		}
	}
	return nil
}

// RemoveTxBlockFromPools removes smartContractResults and miniblocks from associated pools
func (scr *smartContractResults) RemoveTxBlockFromPools(body block.Body, miniBlockPool storage.Cacher) error {
	if body == nil {
		return process.ErrNilTxBlockBody
	}

	err := scr.removeDataFromPools(body, miniBlockPool, scr.scrPool, block.SmartContractResultBlock)

	return err
}

// RestoreTxBlockIntoPools restores the smartContractResults and miniblocks to associated pools
func (scr *smartContractResults) RestoreTxBlockIntoPools(
	body block.Body,
	miniBlockPool storage.Cacher,
) (int, map[int][]byte, error) {
	if miniBlockPool == nil {
		return 0, nil, process.ErrNilMiniBlockPool
	}

	miniBlockHashes := make(map[int][]byte)

	scrRestored := 0
	for i := 0; i < len(body); i++ {
		miniBlock := body[i]
		if miniBlock.Type != block.SmartContractResultBlock {
			continue
		}

		strCache := process.ShardCacherIdentifier(miniBlock.SenderShardID, miniBlock.ReceiverShardID)
		scrBuff, err := scr.storage.GetAll(dataRetriever.UnsignedTransactionUnit, miniBlock.TxHashes)
		if err != nil {
			return scrRestored, miniBlockHashes, err
		}

		for txHash, txBuff := range scrBuff {
			tx := smartContractResult.SmartContractResult{}
			err = scr.marshalizer.Unmarshal(&tx, txBuff)
			if err != nil {
				return scrRestored, miniBlockHashes, err
			}

			scr.scrPool.AddData([]byte(txHash), &tx, strCache)
		}

		restoredHash, err := scr.restoreMiniBlock(miniBlock, miniBlockPool)
		if err != nil {
			return scrRestored, miniBlockHashes, err
		}

		miniBlockHashes[i] = restoredHash
		scrRestored += len(miniBlock.TxHashes)
	}

	return scrRestored, miniBlockHashes, nil
}

// ProcessBlockTransactions processes all the smartContractResult from the block.Body, updates the state
func (scr *smartContractResults) ProcessBlockTransactions(body block.Body, round uint64, haveTime func() time.Duration) error {
	// basic validation already done in interceptors
	for i := 0; i < len(body); i++ {
		miniBlock := body[i]
		if miniBlock.Type != block.SmartContractResultBlock {
			continue
		}
		if miniBlock.ReceiverShardID != scr.shardCoordinator.SelfId() {
			continue
		}

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			if haveTime() < 0 {
				return process.ErrTimeIsOut
			}

			txHash := miniBlock.TxHashes[j]
			scr.scrForBlock.mutTxsForBlock.RLock()
			txInfo := scr.scrForBlock.txHashAndInfo[string(txHash)]
			scr.scrForBlock.mutTxsForBlock.RUnlock()
			if txInfo == nil || txInfo.tx == nil {
				return process.ErrMissingTransaction
			}

			currScr, ok := txInfo.tx.(*smartContractResult.SmartContractResult)
			if !ok {
				return process.ErrWrongTypeAssertion
			}

			err := scr.processSmartContractResult(
				txHash,
				currScr,
				round,
				miniBlock.SenderShardID,
				miniBlock.ReceiverShardID,
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// SaveTxBlockToStorage saves smartContractResults from body into storage
func (scr *smartContractResults) SaveTxBlockToStorage(body block.Body) error {
	for i := 0; i < len(body); i++ {
		miniBlock := (body)[i]
		if miniBlock.Type != block.SmartContractResultBlock || miniBlock.ReceiverShardID != scr.shardCoordinator.SelfId() {
			continue
		}

		err := scr.saveTxsToStorage(miniBlock.TxHashes, &scr.scrForBlock, scr.storage, dataRetriever.UnsignedTransactionUnit)
		if err != nil {
			return err
		}
	}

	return nil
}

// receivedSmartContractResult is a call back function which is called when a new smartContractResult
// is added in the smartContractResult pool
func (scr *smartContractResults) receivedSmartContractResult(txHash []byte) {
	receivedAllMissing := scr.baseReceivedTransaction(txHash, &scr.scrForBlock, scr.scrPool)

	if receivedAllMissing {
		scr.chRcvAllScrs <- true
	}
}

// CreateBlockStarted cleans the local cache map for processed/created smartContractResults at this round
func (scr *smartContractResults) CreateBlockStarted() {
	scr.scrForBlock.mutTxsForBlock.Lock()
	scr.scrForBlock.txHashAndInfo = make(map[string]*txInfo)
	scr.scrForBlock.mutTxsForBlock.Unlock()
}

// RequestBlockTransactions request for smartContractResults if missing from a block.Body
func (scr *smartContractResults) RequestBlockTransactions(body block.Body) int {
	requestedScrs := 0
	missingScrsForShards := scr.computeMissingAndExistingScrsForShards(body)

	scr.scrForBlock.mutTxsForBlock.Lock()
	for senderShardID, scrHashesInfo := range missingScrsForShards {
		txShardInfo := &txShardInfo{senderShardID: senderShardID, receiverShardID: scrHashesInfo.receiverShardID}
		for _, txHash := range scrHashesInfo.txHashes {
			scr.scrForBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: nil, txShardInfo: txShardInfo}
		}
	}
	scr.scrForBlock.mutTxsForBlock.Unlock()

	for senderShardID, scrHashesInfo := range missingScrsForShards {
		requestedScrs += len(scrHashesInfo.txHashes)
		scr.onRequestSmartContractResult(senderShardID, scrHashesInfo.txHashes)
	}

	return requestedScrs
}

// computeMissingAndExistingScrsForShards calculates what smartContractResults are available and what are missing from block.Body
func (scr *smartContractResults) computeMissingAndExistingScrsForShards(body block.Body) map[uint32]*txsHashesInfo {
	onlyScrFromOthersBody := block.Body{}
	for _, mb := range body {
		if mb.Type != block.SmartContractResultBlock {
			continue
		}
		if mb.SenderShardID == scr.shardCoordinator.SelfId() {
			continue
		}

		onlyScrFromOthersBody = append(onlyScrFromOthersBody, mb)
	}

	missingTxsForShard := scr.computeExistingAndMissing(onlyScrFromOthersBody, &scr.scrForBlock, scr.chRcvAllScrs, block.SmartContractResultBlock, scr.scrPool)

	return missingTxsForShard
}

// processAndRemoveBadSmartContractResults processed smartContractResults, if scr are with error it removes them from pool
func (scr *smartContractResults) processSmartContractResult(
	smartContractResultHash []byte,
	smartContractResult *smartContractResult.SmartContractResult,
	round uint64,
	sndShardId uint32,
	dstShardId uint32,
) error {

	err := scr.scrProcessor.ProcessSmartContractResult(smartContractResult)
	if err != nil {
		return err
	}

	txShardInfo := &txShardInfo{senderShardID: sndShardId, receiverShardID: dstShardId}
	scr.scrForBlock.mutTxsForBlock.Lock()
	scr.scrForBlock.txHashAndInfo[string(smartContractResultHash)] = &txInfo{tx: smartContractResult, txShardInfo: txShardInfo}
	scr.scrForBlock.mutTxsForBlock.Unlock()

	return nil
}

// RequestTransactionsForMiniBlock requests missing smartContractResults for a certain miniblock
func (scr *smartContractResults) RequestTransactionsForMiniBlock(mb block.MiniBlock) int {
	missingScrsForMiniBlock := scr.computeMissingScrsForMiniBlock(mb)
	scr.onRequestSmartContractResult(mb.SenderShardID, missingScrsForMiniBlock)

	return len(missingScrsForMiniBlock)
}

// computeMissingScrsForMiniBlock computes missing smartContractResults for a certain miniblock
func (scr *smartContractResults) computeMissingScrsForMiniBlock(mb block.MiniBlock) [][]byte {
	missingSmartContractResults := make([][]byte, 0)
	if mb.Type != block.SmartContractResultBlock {
		return missingSmartContractResults
	}

	for _, txHash := range mb.TxHashes {
		tx, _ := process.GetTransactionHandlerFromPool(
			mb.SenderShardID,
			mb.ReceiverShardID,
			txHash,
			scr.scrPool)

		if tx == nil {
			missingSmartContractResults = append(missingSmartContractResults, txHash)
		}
	}

	return missingSmartContractResults
}

// getAllScrsFromMiniBlock gets all the smartContractResults from a miniblock into a new structure
func (scr *smartContractResults) getAllScrsFromMiniBlock(
	mb *block.MiniBlock,
	haveTime func() bool,
) ([]*smartContractResult.SmartContractResult, [][]byte, error) {

	strCache := process.ShardCacherIdentifier(mb.SenderShardID, mb.ReceiverShardID)
	txCache := scr.scrPool.ShardDataStore(strCache)
	if txCache == nil {
		return nil, nil, process.ErrNilUTxDataPool
	}

	// verify if all smartContractResult exists
	smartContractResults := make([]*smartContractResult.SmartContractResult, 0)
	txHashes := make([][]byte, 0)
	for _, txHash := range mb.TxHashes {
		if !haveTime() {
			return nil, nil, process.ErrTimeIsOut
		}

		tmp, _ := txCache.Peek(txHash)
		if tmp == nil {
			return nil, nil, process.ErrNilSmartContractResult
		}

		tx, ok := tmp.(*smartContractResult.SmartContractResult)
		if !ok {
			return nil, nil, process.ErrWrongTypeAssertion
		}

		txHashes = append(txHashes, txHash)
		smartContractResults = append(smartContractResults, tx)
	}

	return smartContractResults, txHashes, nil
}

// CreateAndProcessMiniBlock creates the miniblock from storage and processes the smartContractResults added into the miniblock
func (scr *smartContractResults) CreateAndProcessMiniBlock(sndShardId, dstShardId uint32, spaceRemained int, haveTime func() bool, round uint64) (*block.MiniBlock, error) {
	return nil, nil
}

// ProcessMiniBlock processes all the smartContractResults from a and saves the processed smartContractResults in local cache complete miniblock
func (scr *smartContractResults) ProcessMiniBlock(miniBlock *block.MiniBlock, haveTime func() bool, round uint64) error {
	if miniBlock.Type != block.SmartContractResultBlock {
		return process.ErrWrongTypeInMiniBlock
	}

	miniBlockScrs, miniBlockTxHashes, err := scr.getAllScrsFromMiniBlock(miniBlock, haveTime)
	if err != nil {
		return err
	}

	for index := range miniBlockScrs {
		if !haveTime() {
			err = process.ErrTimeIsOut
			return err
		}

		err = scr.scrProcessor.ProcessSmartContractResult(miniBlockScrs[index])
		if err != nil {
			return err
		}
	}

	txShardInfo := &txShardInfo{senderShardID: miniBlock.SenderShardID, receiverShardID: miniBlock.ReceiverShardID}

	scr.scrForBlock.mutTxsForBlock.Lock()
	for index, txHash := range miniBlockTxHashes {
		scr.scrForBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: miniBlockScrs[index], txShardInfo: txShardInfo}
	}
	scr.scrForBlock.mutTxsForBlock.Unlock()

	return nil
}

// CreateMarshalizedData marshalizes smartContractResults and creates and saves them into a new structure
func (scr *smartContractResults) CreateMarshalizedData(txHashes [][]byte) ([][]byte, error) {
	mrsScrs, err := scr.createMarshalizedData(txHashes, &scr.scrForBlock)
	if err != nil {
		return nil, err
	}

	return mrsScrs, nil
}

// GetAllCurrentUsedTxs returns all the smartContractResults used at current creation / processing
func (scr *smartContractResults) GetAllCurrentUsedTxs() map[string]data.TransactionHandler {
	scrPool := make(map[string]data.TransactionHandler)

	scr.scrForBlock.mutTxsForBlock.RLock()
	for txHash, txInfo := range scr.scrForBlock.txHashAndInfo {
		scrPool[txHash] = txInfo.tx
	}
	scr.scrForBlock.mutTxsForBlock.RUnlock()

	return scrPool
}
