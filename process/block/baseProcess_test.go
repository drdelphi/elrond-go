package block_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	blproc "github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
)

func haveTime() time.Duration {
	return time.Duration(2000 * time.Millisecond)
}

func createTestBlockchain() *mock.BlockChainMock {
	return &mock.BlockChainMock{}
}

func generateTestCache() storage.Cacher {
	cache, _ := storageUnit.NewCache(storageUnit.LRUCache, 1000, 1)
	return cache
}

func generateTestUnit() storage.Storer {
	memDB, _ := memorydb.New()

	storer, _ := storageUnit.NewStorageUnit(
		generateTestCache(),
		memDB,
	)

	return storer
}

func initDataPool(testHash []byte) *mock.PoolsHolderStub {
	sdp := &mock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{
				RegisterHandlerCalled: func(i func(key []byte)) {},
				ShardDataStoreCalled: func(id string) (c storage.Cacher) {
					return &mock.CacherStub{
						PeekCalled: func(key []byte) (value interface{}, ok bool) {
							if reflect.DeepEqual(key, testHash) {
								return &transaction.Transaction{Nonce: 10}, true
							}
							return nil, false
						},
						KeysCalled: func() [][]byte {
							return [][]byte{[]byte("key1"), []byte("key2")}
						},
						LenCalled: func() int {
							return 0
						},
					}
				},
				RemoveSetOfDataFromPoolCalled: func(keys [][]byte, id string) {},
				SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
					if reflect.DeepEqual(key, []byte("tx1_hash")) {
						return &transaction.Transaction{Nonce: 10}, true
					}
					return nil, false
				},
				AddDataCalled: func(key []byte, data interface{}, cacheId string) {
				},
			}
		},
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{
				RegisterHandlerCalled: func(i func(key []byte)) {},
				ShardDataStoreCalled: func(id string) (c storage.Cacher) {
					return &mock.CacherStub{
						PeekCalled: func(key []byte) (value interface{}, ok bool) {
							if reflect.DeepEqual(key, testHash) {
								return &transaction.Transaction{Nonce: 10}, true
							}
							return nil, false
						},
						KeysCalled: func() [][]byte {
							return [][]byte{[]byte("key1"), []byte("key2")}
						},
						LenCalled: func() int {
							return 0
						},
					}
				},
				RemoveSetOfDataFromPoolCalled: func(keys [][]byte, id string) {},
				SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
					if reflect.DeepEqual(key, []byte("tx1_hash")) {
						return &transaction.Transaction{Nonce: 10}, true
					}
					return nil, false
				},
				AddDataCalled: func(key []byte, data interface{}, cacheId string) {
				},
			}
		},
		HeadersNoncesCalled: func() dataRetriever.Uint64SyncMapCacher {
			return &mock.Uint64SyncMapCacherStub{
				MergeCalled: func(u uint64, syncMap dataRetriever.ShardIdHashMap) {},
				HasCalled: func(nonce uint64, shardId uint32) bool {
					return true
				},
			}
		},
		MetaBlocksCalled: func() storage.Cacher {
			return &mock.CacherStub{
				GetCalled: func(key []byte) (value interface{}, ok bool) {
					if reflect.DeepEqual(key, []byte("tx1_hash")) {
						return &transaction.Transaction{Nonce: 10}, true
					}
					return nil, false
				},
				KeysCalled: func() [][]byte {
					return nil
				},
				LenCalled: func() int {
					return 0
				},
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					if reflect.DeepEqual(key, []byte("tx1_hash")) {
						return &transaction.Transaction{Nonce: 10}, true
					}
					return nil, false
				},
				RegisterHandlerCalled: func(i func(key []byte)) {},
			}
		},
		MiniBlocksCalled: func() storage.Cacher {
			cs := &mock.CacherStub{}
			cs.RegisterHandlerCalled = func(i func(key []byte)) {
			}
			cs.GetCalled = func(key []byte) (value interface{}, ok bool) {
				if bytes.Equal([]byte("bbb"), key) {
					return make(block.MiniBlockSlice, 0), true
				}

				return nil, false
			}
			cs.PeekCalled = func(key []byte) (value interface{}, ok bool) {
				if bytes.Equal([]byte("bbb"), key) {
					return make(block.MiniBlockSlice, 0), true
				}

				return nil, false
			}
			cs.RegisterHandlerCalled = func(i func(key []byte)) {}
			cs.RemoveCalled = func(key []byte) {}
			return cs
		},
		HeadersCalled: func() storage.Cacher {
			cs := &mock.CacherStub{}
			cs.RegisterHandlerCalled = func(i func(key []byte)) {
			}
			return cs
		},
	}
	return sdp
}

func initMetaDataPool() *mock.MetaPoolsHolderStub {
	mdp := &mock.MetaPoolsHolderStub{
		MetaChainBlocksCalled: func() storage.Cacher {
			return &mock.CacherStub{
				GetCalled: func(key []byte) (value interface{}, ok bool) {
					if reflect.DeepEqual(key, []byte("tx1_hash")) {
						return &transaction.Transaction{Nonce: 10}, true
					}
					return nil, false
				},
				KeysCalled: func() [][]byte {
					return nil
				},
				LenCalled: func() int {
					return 0
				},
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					if reflect.DeepEqual(key, []byte("tx1_hash")) {
						return &transaction.Transaction{Nonce: 10}, true
					}
					return nil, false
				},
				RegisterHandlerCalled: func(i func(key []byte)) {},
			}
		},
		MiniBlockHashesCalled: func() dataRetriever.ShardedDataCacherNotifier {
			sdc := &mock.ShardedDataStub{}
			sdc.RegisterHandlerCalled = func(i func(key []byte)) {
			}
			sdc.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
				if bytes.Equal([]byte("bbb"), key) {
					return make(block.MiniBlockSlice, 0), true
				}

				return nil, false
			}
			sdc.RegisterHandlerCalled = func(i func(key []byte)) {}
			sdc.RemoveDataCalled = func(key []byte, cacheId string) {}
			return sdc
		},
		ShardHeadersCalled: func() storage.Cacher {
			cs := &mock.CacherStub{}
			cs.RegisterHandlerCalled = func(i func(key []byte)) {
			}
			cs.PeekCalled = func(key []byte) (value interface{}, ok bool) {
				if bytes.Equal([]byte("hdr_hash1"), key) {
					return &block.Header{Nonce: 1}, true
				}
				return nil, false
			}
			cs.LenCalled = func() int {
				return 0
			}
			cs.RemoveCalled = func(key []byte) {}
			cs.KeysCalled = func() [][]byte {
				return nil
			}
			return cs
		},
		HeadersNoncesCalled: func() dataRetriever.Uint64SyncMapCacher {
			cs := &mock.Uint64SyncMapCacherStub{}
			cs.MergeCalled = func(u uint64, syncMap dataRetriever.ShardIdHashMap) {}
			cs.RemoveCalled = func(nonce uint64, shardId uint32) {}
			return cs
		},
	}
	return mdp
}

func initStore() *dataRetriever.ChainStorer {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, generateTestUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, generateTestUnit())
	store.AddStorer(dataRetriever.MetaBlockUnit, generateTestUnit())
	store.AddStorer(dataRetriever.PeerChangesUnit, generateTestUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, generateTestUnit())
	store.AddStorer(dataRetriever.ShardHdrNonceHashDataUnit, generateTestUnit())
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, generateTestUnit())
	return store
}

func createDummyMetaBlock(destShardId uint32, senderShardId uint32, miniBlockHashes ...[]byte) data.HeaderHandler {
	metaBlock := &block.MetaBlock{
		ShardInfo: []block.ShardData{
			{
				ShardMiniBlockHeaders: make([]block.ShardMiniBlockHeader, len(miniBlockHashes)),
			},
		},
	}

	for idx, mbHash := range miniBlockHashes {
		metaBlock.ShardInfo[0].ShardMiniBlockHeaders[idx].ReceiverShardId = destShardId
		metaBlock.ShardInfo[0].ShardMiniBlockHeaders[idx].SenderShardId = senderShardId
		metaBlock.ShardInfo[0].ShardMiniBlockHeaders[idx].Hash = mbHash
	}

	return metaBlock
}

func createDummyMiniBlock(
	txHash string,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	destShardId uint32,
	senderShardId uint32) (*block.MiniBlock, []byte) {

	miniblock := &block.MiniBlock{
		TxHashes:        [][]byte{[]byte(txHash)},
		ReceiverShardID: destShardId,
		SenderShardID:   senderShardId,
	}

	buff, _ := marshalizer.Marshal(miniblock)
	hash := hasher.Compute(string(buff))

	return miniblock, hash
}

func isInTxHashes(searched []byte, list [][]byte) bool {
	for _, txHash := range list {
		if bytes.Equal(txHash, searched) {
			return true
		}
	}
	return false
}

type wrongBody struct {
}

func (wr wrongBody) IntegrityAndValidity() error {
	return nil
}

func TestBlockProcessor_CheckBlockValidity(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte(""))
	bp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		initStore(),
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	blkc := createTestBlockchain()
	body := &block.Body{}
	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.Round = 1
	hdr.TimeStamp = 0
	hdr.PrevHash = []byte("X")
	err := bp.CheckBlockValidity(blkc, hdr, body)
	assert.Equal(t, process.ErrBlockHashDoesNotMatch, err)

	hdr.PrevHash = []byte("")
	err = bp.CheckBlockValidity(blkc, hdr, body)
	assert.Nil(t, err)

	hdr.Nonce = 2
	err = bp.CheckBlockValidity(blkc, hdr, body)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)

	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &block.Header{Round: 1, Nonce: 1}
	}
	hdr = &block.Header{}

	err = bp.CheckBlockValidity(blkc, hdr, body)
	assert.Equal(t, process.ErrLowerRoundInBlock, err)

	hdr.Round = 2
	hdr.Nonce = 1
	err = bp.CheckBlockValidity(blkc, hdr, body)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)

	hdr.Nonce = 2
	hdr.PrevRandSeed = []byte("X")
	err = bp.CheckBlockValidity(blkc, hdr, body)
	assert.Equal(t, process.ErrRandSeedMismatch, err)

	hdr.PrevRandSeed = []byte("")
	hdr.PrevHash = []byte("X")
	err = bp.CheckBlockValidity(blkc, hdr, body)
	assert.Equal(t, process.ErrBlockHashDoesNotMatch, err)

	hdr.Nonce = 3
	hdr.PrevHash = []byte("")
	err = bp.CheckBlockValidity(blkc, hdr, body)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)

	hdr.Nonce = 2
	marshalizerMock := mock.MarshalizerMock{}
	hasherMock := mock.HasherMock{}
	prevHeader, _ := marshalizerMock.Marshal(blkc.GetCurrentBlockHeader())
	hdr.PrevHash = hasherMock.Compute(string(prevHeader))

	err = bp.CheckBlockValidity(blkc, hdr, body)
	assert.Nil(t, err)
}

func TestVerifyStateRoot_ShouldWork(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte(""))
	rootHash := []byte("root hash to be tested")
	accounts := &mock.AccountsStub{
		RootHashCalled: func() ([]byte, error) {
			return rootHash, nil
		},
	}
	store := initStore()

	bp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		store,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		accounts,
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	assert.True(t, bp.VerifyStateRoot(rootHash))
}

//------- ComputeNewNoncePrevHash

func TestBlockProcessor_computeHeaderHashMarshalizerFail1ShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte(""))
	marshalizer := &mock.MarshalizerStub{}
	bp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		initStore(),
		&mock.HasherStub{},
		marshalizer,
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	hdr, txBlock := createTestHdrTxBlockBody()
	expectedError := errors.New("marshalizer fail")
	marshalizer.MarshalCalled = func(obj interface{}) (bytes []byte, e error) {
		if hdr == obj {
			return nil, expectedError
		}

		if reflect.DeepEqual(txBlock, obj) {
			return []byte("txBlockBodyMarshalized"), nil
		}
		return nil, nil
	}
	_, err := bp.ComputeHeaderHash(hdr)
	assert.Equal(t, expectedError, err)
}

func TestBlockPorcessor_ComputeNewNoncePrevHashShouldWork(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte(""))
	marshalizer := &mock.MarshalizerStub{}
	hasher := &mock.HasherStub{}
	bp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		initStore(),
		hasher,
		marshalizer,
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewOneShardCoordinatorMock()),
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	hdr, txBlock := createTestHdrTxBlockBody()
	marshalizer.MarshalCalled = func(obj interface{}) (bytes []byte, e error) {
		if hdr == obj {
			return []byte("hdrHeaderMarshalized"), nil
		}
		if reflect.DeepEqual(txBlock, obj) {
			return []byte("txBlockBodyMarshalized"), nil
		}
		return nil, nil
	}
	hasher.ComputeCalled = func(s string) []byte {
		if s == "hdrHeaderMarshalized" {
			return []byte("hdr hash")
		}
		if s == "txBlockBodyMarshalized" {
			return []byte("tx block body hash")
		}
		return nil
	}
	_, err := bp.ComputeHeaderHash(hdr)
	assert.Nil(t, err)
}

func TestBlockPorcessor_DisplayHeaderShouldWork(t *testing.T) {
	lines := blproc.DisplayHeader(&block.Header{})
	assert.Equal(t, 10, len(lines))
}

func TestBaseProcessor_SetLastNotarizedHeadersSliceNil(t *testing.T) {
	t.Parallel()

	base := blproc.NewBaseProcessor(mock.NewOneShardCoordinatorMock())

	err := base.SetLastNotarizedHeadersSlice(nil)

	assert.Equal(t, process.ErrNotarizedHdrsSliceIsNil, err)
}

func TestBaseProcessor_SetLastNotarizedHeadersSliceNotEnoughHeaders(t *testing.T) {
	t.Parallel()

	base := blproc.NewBaseProcessor(mock.NewOneShardCoordinatorMock())

	err := base.SetLastNotarizedHeadersSlice(make(map[uint32]data.HeaderHandler, 0))

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestBaseProcessor_SetLastNotarizedHeadersSliceOneShardWrongType(t *testing.T) {
	t.Parallel()

	base := blproc.NewBaseProcessor(mock.NewOneShardCoordinatorMock())

	lastNotHdrs := createGenesisBlocks(mock.NewOneShardCoordinatorMock())
	lastNotHdrs[0] = &block.MetaBlock{}
	err := base.SetLastNotarizedHeadersSlice(lastNotHdrs)

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestBaseProcessor_SetLastNotarizedHeadersSliceOneShardGood(t *testing.T) {
	t.Parallel()

	base := blproc.NewBaseProcessor(mock.NewOneShardCoordinatorMock())

	lastNotHdrs := createGenesisBlocks(mock.NewOneShardCoordinatorMock())
	err := base.SetLastNotarizedHeadersSlice(lastNotHdrs)

	assert.Nil(t, err)
}

func TestBaseProcessor_SetLastNotarizedHeadersSliceOneShardMetaMissing(t *testing.T) {
	t.Parallel()

	base := blproc.NewBaseProcessor(mock.NewOneShardCoordinatorMock())

	lastNotHdrs := createGenesisBlocks(mock.NewOneShardCoordinatorMock())
	lastNotHdrs[sharding.MetachainShardId] = nil
	err := base.SetLastNotarizedHeadersSlice(lastNotHdrs)

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestBaseProcessor_SetLastNotarizedHeadersSliceOneShardMetaWrongType(t *testing.T) {
	t.Parallel()

	base := blproc.NewBaseProcessor(mock.NewOneShardCoordinatorMock())

	lastNotHdrs := createGenesisBlocks(mock.NewOneShardCoordinatorMock())
	lastNotHdrs[sharding.MetachainShardId] = &block.Header{}
	err := base.SetLastNotarizedHeadersSlice(lastNotHdrs)

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestBaseProcessor_SetLastNotarizedHeadersSliceMultiShardGood(t *testing.T) {
	t.Parallel()

	base := blproc.NewBaseProcessor(mock.NewMultiShardsCoordinatorMock(5))

	lastNotHdrs := createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(5))
	err := base.SetLastNotarizedHeadersSlice(lastNotHdrs)

	assert.Nil(t, err)
}

func TestBaseProcessor_SetLastNotarizedHeadersSliceMultiShardNotEnough(t *testing.T) {
	t.Parallel()

	base := blproc.NewBaseProcessor(mock.NewMultiShardsCoordinatorMock(5))

	lastNotHdrs := createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(4))
	lastNotHdrs[sharding.MetachainShardId] = nil
	err := base.SetLastNotarizedHeadersSlice(lastNotHdrs)

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func createShardProcessHeadersToSaveLastNoterized(
	highestNonce uint64,
	genesisHdr data.HeaderHandler,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
) []data.HeaderHandler {
	rootHash := []byte("roothash")
	processedHdrs := make([]data.HeaderHandler, 0)

	headerMarsh, _ := marshalizer.Marshal(genesisHdr)
	headerHash := hasher.Compute(string(headerMarsh))

	for i := uint64(1); i <= highestNonce; i++ {
		hdr := &block.Header{
			Nonce:         i,
			Round:         uint64(i),
			Signature:     rootHash,
			RandSeed:      rootHash,
			PrevRandSeed:  rootHash,
			PubKeysBitmap: rootHash,
			RootHash:      rootHash,
			PrevHash:      headerHash}
		processedHdrs = append(processedHdrs, hdr)

		headerMarsh, _ = marshalizer.Marshal(hdr)
		headerHash = hasher.Compute(string(headerMarsh))
	}

	return processedHdrs
}

func createMetaProcessHeadersToSaveLastNoterized(
	highestNonce uint64,
	genesisHdr data.HeaderHandler,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
) []data.HeaderHandler {
	rootHash := []byte("roothash")
	processedHdrs := make([]data.HeaderHandler, 0)

	headerMarsh, _ := marshalizer.Marshal(genesisHdr)
	headerHash := hasher.Compute(string(headerMarsh))

	for i := uint64(1); i <= highestNonce; i++ {
		hdr := &block.MetaBlock{
			Nonce:         i,
			Round:         uint64(i),
			Signature:     rootHash,
			RandSeed:      rootHash,
			PrevRandSeed:  rootHash,
			PubKeysBitmap: rootHash,
			RootHash:      rootHash,
			PrevHash:      headerHash}
		processedHdrs = append(processedHdrs, hdr)

		headerMarsh, _ = marshalizer.Marshal(hdr)
		headerHash = hasher.Compute(string(headerMarsh))
	}

	return processedHdrs
}

func TestBaseProcessor_SaveLastNoterizedHdrLastNotSliceNotSet(t *testing.T) {
	t.Parallel()

	base := blproc.NewBaseProcessor(mock.NewMultiShardsCoordinatorMock(5))
	base.SetHasher(mock.HasherMock{})
	base.SetMarshalizer(&mock.MarshalizerMock{})
	prHdrs := createShardProcessHeadersToSaveLastNoterized(10, &block.Header{}, mock.HasherMock{}, &mock.MarshalizerMock{})

	err := base.SaveLastNotarizedHeader(2, prHdrs)

	assert.Equal(t, process.ErrNotarizedHdrsSliceIsNil, err)
}

func TestBaseProcessor_SaveLastNoterizedHdrLastNotShardIdMissmatch(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	base := blproc.NewBaseProcessor(shardCoordinator)
	base.SetHasher(mock.HasherMock{})
	base.SetMarshalizer(&mock.MarshalizerMock{})
	_ = base.SetLastNotarizedHeadersSlice(createGenesisBlocks(shardCoordinator))
	prHdrs := createShardProcessHeadersToSaveLastNoterized(10, &block.Header{}, mock.HasherMock{}, &mock.MarshalizerMock{})

	err := base.SaveLastNotarizedHeader(6, prHdrs)

	assert.Equal(t, process.ErrShardIdMissmatch, err)
}

func TestBaseProcessor_SaveLastNoterizedHdrLastNotHdrNil(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	base := blproc.NewBaseProcessor(shardCoordinator)
	base.SetHasher(mock.HasherMock{})
	base.SetMarshalizer(&mock.MarshalizerMock{})

	// make it wrong
	shardId := uint32(2)
	genesisBlock := createGenesisBlocks(shardCoordinator)
	genesisBlock[shardId] = nil

	_ = base.SetLastNotarizedHeadersSlice(genesisBlock)
	prHdrs := createShardProcessHeadersToSaveLastNoterized(10, &block.Header{}, mock.HasherMock{}, &mock.MarshalizerMock{})

	err := base.SaveLastNotarizedHeader(shardId, prHdrs)

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestBaseProcessor_SaveLastNoterizedHdrLastNotWrongTypeShard(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	base := blproc.NewBaseProcessor(shardCoordinator)
	base.SetHasher(mock.HasherMock{})
	base.SetMarshalizer(&mock.MarshalizerMock{})

	// make it wrong
	shardId := uint32(2)
	genesisBlock := createGenesisBlocks(shardCoordinator)
	genesisBlock[shardId] = &block.MetaBlock{Nonce: 0}

	_ = base.SetLastNotarizedHeadersSlice(genesisBlock)
	prHdrs := createShardProcessHeadersToSaveLastNoterized(10, &block.Header{}, mock.HasherMock{}, &mock.MarshalizerMock{})

	err := base.SaveLastNotarizedHeader(shardId, prHdrs)

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestBaseProcessor_SaveLastNoterizedHdrLastNotWrongTypeMeta(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	base := blproc.NewBaseProcessor(shardCoordinator)
	base.SetHasher(mock.HasherMock{})
	base.SetMarshalizer(&mock.MarshalizerMock{})

	// make it wrong
	genesisBlock := createGenesisBlocks(shardCoordinator)
	genesisBlock[sharding.MetachainShardId] = &block.Header{Nonce: 0}

	_ = base.SetLastNotarizedHeadersSlice(genesisBlock)
	prHdrs := createMetaProcessHeadersToSaveLastNoterized(10, &block.Header{}, mock.HasherMock{}, &mock.MarshalizerMock{})

	err := base.SaveLastNotarizedHeader(sharding.MetachainShardId, prHdrs)

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestBaseProcessor_SaveLastNoterizedHdrShardWrongProcessed(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	base := blproc.NewBaseProcessor(shardCoordinator)
	base.SetHasher(mock.HasherMock{})
	base.SetMarshalizer(&mock.MarshalizerMock{})
	_ = base.SetLastNotarizedHeadersSlice(createGenesisBlocks(shardCoordinator))
	highestNonce := uint64(10)
	prHdrs := createMetaProcessHeadersToSaveLastNoterized(highestNonce, &block.Header{}, mock.HasherMock{}, &mock.MarshalizerMock{})

	shardId := uint32(0)
	err := base.SaveLastNotarizedHeader(shardId, prHdrs)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)

	notarizedHdrs := base.NotarizedHdrs()
	assert.Equal(t, uint64(0), notarizedHdrs[shardId][0].GetNonce())
}

func TestBaseProcessor_SaveLastNoterizedHdrMetaWrongProcessed(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	base := blproc.NewBaseProcessor(shardCoordinator)
	base.SetHasher(mock.HasherMock{})
	base.SetMarshalizer(&mock.MarshalizerMock{})
	_ = base.SetLastNotarizedHeadersSlice(createGenesisBlocks(shardCoordinator))
	highestNonce := uint64(10)
	prHdrs := createShardProcessHeadersToSaveLastNoterized(highestNonce, &block.Header{}, mock.HasherMock{}, &mock.MarshalizerMock{})

	err := base.SaveLastNotarizedHeader(sharding.MetachainShardId, prHdrs)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)

	notarizedHdrs := base.NotarizedHdrs()
	assert.Equal(t, uint64(0), notarizedHdrs[sharding.MetachainShardId][0].GetNonce())
}

func TestBaseProcessor_SaveLastNoterizedHdrShardGood(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	base := blproc.NewBaseProcessor(shardCoordinator)
	hasher := mock.HasherMock{}
	base.SetHasher(hasher)
	marshalizer := &mock.MarshalizerMock{}
	base.SetMarshalizer(marshalizer)
	genesisBlcks := createGenesisBlocks(shardCoordinator)
	_ = base.SetLastNotarizedHeadersSlice(genesisBlcks)

	highestNonce := uint64(10)
	shardId := uint32(0)
	prHdrs := createShardProcessHeadersToSaveLastNoterized(highestNonce, genesisBlcks[shardId], hasher, marshalizer)

	err := base.SaveLastNotarizedHeader(shardId, prHdrs)
	assert.Nil(t, err)

	assert.Equal(t, highestNonce, base.LastNotarizedHdrForShard(shardId).GetNonce())
}

func TestBaseProcessor_SaveLastNoterizedHdrMetaGood(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	base := blproc.NewBaseProcessor(shardCoordinator)
	hasher := mock.HasherMock{}
	base.SetHasher(hasher)
	marshalizer := &mock.MarshalizerMock{}
	base.SetMarshalizer(marshalizer)
	genesisBlcks := createGenesisBlocks(shardCoordinator)
	_ = base.SetLastNotarizedHeadersSlice(genesisBlcks)

	highestNonce := uint64(10)
	prHdrs := createMetaProcessHeadersToSaveLastNoterized(highestNonce, genesisBlcks[sharding.MetachainShardId], hasher, marshalizer)

	err := base.SaveLastNotarizedHeader(sharding.MetachainShardId, prHdrs)
	assert.Nil(t, err)

	assert.Equal(t, highestNonce, base.LastNotarizedHdrForShard(sharding.MetachainShardId).GetNonce())
}
