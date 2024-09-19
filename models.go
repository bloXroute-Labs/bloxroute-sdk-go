package bloxroute_sdk_go

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/bloXroute-Labs/gateway/v2/jsonrpc"
)

const (
	RPCBSCGetBundlePrice jsonrpc.RPCRequestType = "bsc_get_bundle_price"
	RPCBSCPrivateTx      jsonrpc.RPCRequestType = "bsc_private_tx"
	RPCPolygonPrivateTx  jsonrpc.RPCRequestType = "polygon_private_tx"
)

// OnBlockNotification represents the result of an RPC call on published block
type OnBlockNotification struct {
	Name        string `json:"name,omitempty"`
	Response    string `json:"response,omitempty"`
	BlockHeight string `json:"block_height,omitempty"`
	Tag         string `json:"tag,omitempty"`
}

// NewTxNotification is the notification object for new transactions
// Note: the fields in this object differs in WebSockets and gRPC
type NewTxNotification struct {
	TxHash      string                       `json:"txHash"`
	TxContents  *NewTxNotificationTxContents `json:"txContents"`
	LocalRegion bool                         `json:"localRegion"`
	Time        string                       `json:"time"`
	RawTx       string                       `json:"rawTx"`
}

// NewTxNotificationTxContents is the transaction contents object for new transactions
type NewTxNotificationTxContents struct {
	AccessList           types.AccessList `json:"accessList"`
	ChainId              string           `json:"chainId"`
	From                 string           `json:"from"`
	Gas                  string           `json:"gas"`
	GasPrice             string           `json:"gasPrice"`
	Hash                 string           `json:"hash"`
	Input                string           `json:"input"`
	MaxFeePerGas         string           `json:"maxFeePerGas"`
	MaxFeePerBlobGas     string           `json:"maxFeePerBlobGas"`
	MaxPriorityFeePerGas string           `json:"maxPriorityFeePerGas"`
	Nonce                string           `json:"nonce"`
	R                    string           `json:"r"`
	S                    string           `json:"s"`
	To                   string           `json:"to"`
	Type                 string           `json:"type"`
	V                    string           `json:"v"`
	Value                string           `json:"value"`
	BlobVersionedHashes  []string         `json:"blobVersionedHashes"`
	YParity              string           `json:"yParity"`
}

// OnTxStatusNotification represents status of a transaction
type OnTxStatusNotification struct {
	TxHash string `json:"txHash"`
	Status string `json:"status"`
}

// OnTxReceiptNotification represents transaction receipt
type OnTxReceiptNotification struct {
	BlockHash         string                       `json:"block_hash"`
	BlockNumber       string                       `json:"block_number"`
	ContractAddress   interface{}                  `json:"contract_address"`
	CumulativeGasUsed string                       `json:"cumulative_gas_used"`
	EffectiveGasUsed  string                       `json:"effective_gas_used"`
	From              string                       `json:"from"`
	GasUsed           string                       `json:"gas_used"`
	Logs              []OnTxReceiptNotificationLog `json:"logs"`
	LogsBloom         string                       `json:"logs_bloom"`
	Status            string                       `json:"status"`
	To                string                       `json:"to"`
	TransactionHash   string                       `json:"transaction_hash"`
	TransactionIndex  string                       `json:"transaction_index"`
	Type              string                       `json:"type"`
	TxsCount          string                       `json:"txs_count"`
	BlobGasUsed       string                       `json:"blobGasUsed"`
	BlobGasPrice      string                       `json:"blobGasPrice"`
}

// OnTxReceiptNotificationLog represents transaction receipt log
type OnTxReceiptNotificationLog struct {
	Address          string   `json:"address"`
	Topics           []string `json:"topics"`
	Data             string   `json:"data"`
	BlockNumber      string   `json:"blockNumber"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
	BlockHash        string   `json:"blockHash"`
	LogIndex         string   `json:"logIndex"`
	Removed          bool     `json:"removed"`
}

// Header represents the header of a block
type Header struct {
	ParentHash       string       `json:"parentHash"`
	Sha3Uncles       string       `json:"sha3Uncles"`
	Miner            string       `json:"miner"`
	StateRoot        string       `json:"stateRoot"`
	TransactionsRoot string       `json:"transactionsRoot"`
	ReceiptsRoot     string       `json:"receiptsRoot"`
	LogsBloom        string       `json:"logsBloom"`
	Difficulty       string       `json:"difficulty"`
	Number           string       `json:"number"`
	GasLimit         string       `json:"gasLimit"`
	GasUsed          string       `json:"gasUsed"`
	Timestamp        string       `json:"timestamp"`
	ExtraData        string       `json:"extraData"`
	MixHash          string       `json:"mixHash"`
	Nonce            string       `json:"nonce"`
	BaseFeePerGas    *int         `json:"baseFeePerGas"`
	WithdrawalsRoot  *common.Hash `json:"withdrawalsRoot"`
	BlobGasUsed      string       `json:"blobGasUsed"`
	ExcessBlobGas    string       `json:"excessBlobGas"`
	ParentBeaconRoot *common.Hash `json:"parentBeaconBlockRoot"`
}

// FutureValidatorInfo represents the future validator info of a block
type FutureValidatorInfo struct {
	BlockHeight string `json:"block_height"`
	WalletId    string `json:"wallet_id"`
	Accessible  string `json:"accessible"`
}

// OnNewBlockTransaction differs for WebSockets and gRPC
type OnNewBlockTransaction struct {
	From                 string           `json:"from"`
	RawTx                []byte           `json:"rawTx"`
	AccessList           types.AccessList `json:"accessList"`
	BlobVersionedHashes  []common.Hash    `json:"blobVersionedHashes"`
	ChainID              string           `json:"chainId"`
	Gas                  string           `json:"gas"`
	GasPrice             string           `json:"gasPrice"`
	Hash                 string           `json:"hash"`
	Input                string           `json:"input"`
	MaxFeePerGas         string           `json:"maxFeePerGas"`
	MaxPriorityFeePerGas string           `json:"maxPriorityFeePerGas"`
	Nonce                string           `json:"nonce"`
	R                    string           `json:"r"`
	S                    string           `json:"s"`
	To                   string           `json:"to"`
	Type                 string           `json:"type"`
	V                    string           `json:"v"`
	Value                string           `json:"value"`
	YParity              string           `json:"yParity"`
}

// OnBlockWithdrawal represents the withdrawal object for a block
type OnBlockWithdrawal struct {
	Address        string `json:"address"`
	Amount         string `json:"amount"`
	Index          string `json:"index"`
	ValidatorIndex string `json:"validator_index"`
}

// OnBdnBlockNotification represents the block notification object for BDN
type OnBdnBlockNotification struct {
	Hash                string                  `json:"hash"`
	Header              *Header                 `json:"header"`
	FutureValidatorInfo []FutureValidatorInfo   `json:"future_validator_info"`
	Transactions        []OnNewBlockTransaction `json:"transactions"`
	Withdrawals         []OnBlockWithdrawal     `json:"withdrawals"`
}
