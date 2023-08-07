package bloxroute_sdk_go

import "github.com/ethereum/go-ethereum/core/types"

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
	MaxPriorityFeePerGas string           `json:"maxPriorityFeePerGas"`
	Nonce                string           `json:"nonce"`
	R                    string           `json:"r"`
	S                    string           `json:"s"`
	To                   string           `json:"to"`
	Type                 string           `json:"type"`
	V                    string           `json:"v"`
	Value                string           `json:"value"`
}

// OnTxStatusNotification represents status of a transaction
type OnTxStatusNotification struct {
	TxHash string `json:"txHash"`
	Status string `json:"status"`
}

// OnTxReceiptNotification represents transaction receipt
type OnTxReceiptNotification struct {
	BlockHash         string                       `json:"blockHash"`
	BlockNumber       string                       `json:"blockNumber"`
	ContractAddress   interface{}                  `json:"contractAddress"`
	CumulativeGasUsed string                       `json:"cumulativeGasUsed"`
	From              string                       `json:"from"`
	GasUsed           string                       `json:"gasUsed"`
	Logs              []OnTxReceiptNotificationLog `json:"logs"`
	LogsBloom         string                       `json:"logsBloom"`
	Status            string                       `json:"status"`
	To                string                       `json:"to"`
	TransactionHash   string                       `json:"transactionHash"`
	TransactionIndex  string                       `json:"transactionIndex"`
	Type              string                       `json:"type"`
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
	ParentHash       string `json:"parentHash"`
	Sha3Uncles       string `json:"sha3Uncles"`
	Miner            string `json:"miner"`
	StateRoot        string `json:"stateRoot"`
	TransactionsRoot string `json:"transactionsRoot"`
	ReceiptsRoot     string `json:"receiptsRoot"`
	LogsBloom        string `json:"logsBloom"`
	Difficulty       string `json:"difficulty"`
	Number           string `json:"number"`
	GasLimit         string `json:"gasLimit"`
	GasUsed          string `json:"gasUsed"`
	Timestamp        string `json:"timestamp"`
	ExtraData        string `json:"extraData"`
	MixHash          string `json:"mixHash"`
	Nonce            string `json:"nonce"`
}

// FutureValidatorInfo represents the future validator info of a block
type FutureValidatorInfo struct {
	BlockHeight string `json:"block_height"`
	WalletId    string `json:"wallet_id"`
	Accessible  string `json:"accessible"`
}

// OnNewBlockTransaction differs for WebSockets and gRPC
type OnNewBlockTransaction struct {
	TxHash string `json:"txHash"`
	RawTx  []byte
	From   []byte
}

// OnBdnBlockNotification represents the block notification object for BDN
type OnBdnBlockNotification struct {
	Hash                string                  `json:"hash"`
	Header              *Header                 `json:"header"`
	FutureValidatorInfo []FutureValidatorInfo   `json:"future_validator_info"`
	Transactions        []OnNewBlockTransaction `json:"transactions"`
	// Uncles              []types.Block                 `json:"uncles"`
}
