package unsigned_test

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/unsigned"
	"github.com/stretchr/testify/assert"
)

var senderShard = uint32(2)
var recvShard = uint32(3)
var senderAddress = []byte("sender")
var recvAddress = []byte("receiver")

func createInterceptedScrFromPlainScr(scr *smartContractResult.SmartContractResult) (*unsigned.InterceptedUnsignedTransaction, error) {
	marshalizer := &mock.MarshalizerMock{}
	txBuff, _ := marshalizer.Marshal(scr)

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.CurrentShard = 6
	shardCoordinator.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		if bytes.Equal(address.Bytes(), senderAddress) {
			return senderShard
		}
		if bytes.Equal(address.Bytes(), recvAddress) {
			return recvShard
		}

		return shardCoordinator.CurrentShard
	}

	return unsigned.NewInterceptedUnsignedTransaction(
		txBuff,
		marshalizer,
		mock.HasherMock{},
		&mock.AddressConverterStub{
			CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
				return mock.NewAddressMock(pubKey), nil
			},
		},
		shardCoordinator,
	)
}

func TestNewInterceptedUnsignedTransaction_NilBufferShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := unsigned.NewInterceptedUnsignedTransaction(
		nil,
		&mock.MarshalizerMock{},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilBuffer, err)
}

func TestNewInterceptedUnsignedTransaction_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := unsigned.NewInterceptedUnsignedTransaction(
		make([]byte, 0),
		nil,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedUnsignedTransaction_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := unsigned.NewInterceptedUnsignedTransaction(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		nil,
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptedUnsignedTransaction_NilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := unsigned.NewInterceptedUnsignedTransaction(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		mock.HasherMock{},
		nil,
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewInterceptedUnsignedTransaction_NilCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := unsigned.NewInterceptedUnsignedTransaction(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		nil,
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewInterceptedUnsignedTransaction_UnmarshalingTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")

	txi, err := unsigned.NewInterceptedUnsignedTransaction(
		make([]byte, 0),
		&mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return errExpected
			},
		},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, errExpected, err)
}

func TestNewInterceptedUnsignedTransaction_AddrConvFailsShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := unsigned.NewInterceptedUnsignedTransaction(
		[]byte("{}"),
		&mock.MarshalizerMock{},
		mock.HasherMock{},
		&mock.AddressConverterStub{
			CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
				return nil, errors.New("expected error")
			},
		},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrInvalidSndAddr, err)
}

func TestNewInterceptedUnsignedTransaction_NilTxHashShouldErr(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:   1,
		Value:   big.NewInt(2),
		Data:    "data",
		RcvAddr: recvAddress,
		SndAddr: senderAddress,
		TxHash:  nil,
	}

	txi, err := createInterceptedScrFromPlainScr(tx)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilTxHash, err)
}

func TestNewInterceptedUnsignedTransaction_NilSenderAddressShouldErr(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:   1,
		Value:   big.NewInt(2),
		Data:    "data",
		RcvAddr: recvAddress,
		SndAddr: nil,
		TxHash:  []byte("TX"),
	}

	txi, err := createInterceptedScrFromPlainScr(tx)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilSndAddr, err)
}

func TestNewInterceptedUnsignedTransaction_NilRecvAddressShouldErr(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:   1,
		Value:   big.NewInt(2),
		Data:    "data",
		RcvAddr: nil,
		SndAddr: senderAddress,
		TxHash:  []byte("TX"),
	}

	txi, err := createInterceptedScrFromPlainScr(tx)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilRcvAddr, err)
}

func TestNewInterceptedUnsignedTransaction_NilValueShouldErr(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:   1,
		Value:   nil,
		Data:    "data",
		RcvAddr: recvAddress,
		SndAddr: senderAddress,
		TxHash:  []byte("TX"),
	}

	txi, err := createInterceptedScrFromPlainScr(tx)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilValue, err)
}

func TestNewInterceptedUnsignedTransaction_NilNegativeValueShouldErr(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:   1,
		Value:   big.NewInt(-2),
		Data:    "data",
		RcvAddr: recvAddress,
		SndAddr: senderAddress,
		TxHash:  []byte("TX"),
	}

	txi, err := createInterceptedScrFromPlainScr(tx)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNegativeValue, err)
}

func TestNewInterceptedUnsignedTransaction_InvalidSenderShouldErr(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:   1,
		Value:   big.NewInt(2),
		Data:    "data",
		RcvAddr: recvAddress,
		SndAddr: []byte(""),
		TxHash:  []byte("TX"),
	}

	txi, err := createInterceptedScrFromPlainScr(tx)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilSndAddr, err)
}

func TestNewInterceptedUnsignedTransaction_ShouldWork(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:   1,
		Value:   big.NewInt(2),
		Data:    "data",
		RcvAddr: recvAddress,
		SndAddr: senderAddress,
		TxHash:  []byte("TX"),
	}

	txi, err := createInterceptedScrFromPlainScr(tx)

	assert.NotNil(t, txi)
	assert.Nil(t, err)
	assert.Equal(t, tx, txi.UnsignedTransaction())
}

func TestNewInterceptedUnsignedTransaction_OkValsGettersShouldWork(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:   1,
		Value:   big.NewInt(2),
		Data:    "data",
		RcvAddr: recvAddress,
		SndAddr: senderAddress,
		TxHash:  []byte("TX"),
	}

	txi, _ := createInterceptedScrFromPlainScr(tx)

	assert.Equal(t, senderShard, txi.SndShard())
	assert.Equal(t, recvShard, txi.RcvShard())
	assert.True(t, txi.IsAddressedToOtherShards())
	assert.Equal(t, tx, txi.UnsignedTransaction())
}
