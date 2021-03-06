package process

// BlockHeaderState specifies which is the state of the block header received
type BlockHeaderState int

const (
	// BHReceived defines ID of a received block header
	BHReceived BlockHeaderState = iota
	// BHProcessed defines ID of a processed block header
	BHProcessed
	// BHProposed defines ID of a proposed block header
	BHProposed
	// BHNotarized defines ID of a notarized block header
	BHNotarized
)

// TransactionType specifies the type of the transaction
type TransactionType int

const (
	// MoveBalance defines ID of a payment transaction - moving balances
	MoveBalance TransactionType = iota
	// SCDeployment defines ID of a transaction to store a smart contract
	SCDeployment
	// SCInvoking defines ID of a transaction of type smart contract call
	SCInvoking
	// InvalidTransaction defines unknown transaction type
	InvalidTransaction
)

const ShardBlockFinality = 1
const MetaBlockFinality = 1
const ForkBlockFinality = 1
const MaxHeaderRequestsAllowed = 10
const MaxItemsInBlock = 15000

// TODO - calculate exactly in case of the VM, for every VM to have a similar constant, operations / seconds
const MaxGasLimitPerMiniBlock = uint64(100000)
const MaxRequestsWithTimeoutAllowed = 5
