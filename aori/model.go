package aori

const (
	EventRfqQuoteRequested    = "QuoteRequested"
	EventRfqQuoteReceived     = "QuoteReceived"
	EventRfqCallDataToExecute = "CalldataToExecute"
)

type RfqQuoteRequest struct {
	RfqId       string `json:"rfqId"`
	Address     string `json:"address"`
	InputToken  string `json:"inputToken"`
	OutputToken string `json:"outputToken"`
	InputAmount string `json:"inputAmount"`
	Zone        string `json:"zone"`
	ChainId     int    `json:"chainId"`
	Deadline    int    `json:"deadline"`
}

type RfqQuoteReceived struct {
	RfqId        string `json:"rfqId"`
	Address      string `json:"address"`
	ChainId      int    `json:"chainId"`
	InputToken   string `json:"inputToken"`
	InputAmount  string `json:"inputAmount"`
	OutputToken  string `json:"outputToken"`
	OutputAmount string `json:"outputAmount"`
	Zone         string `json:"zone"`
}

type RfqCallDataToExecute struct {
	RfqId        string `json:"rfqId"`
	MatchingHash string `json:"matchingHash"`
	Matching     struct {
		MakerOrder        *RfqSolution `json:"makerOrder"`
		TakerOrder        *RfqSolution `json:"takerOrder"`
		MakerSignature    string       `json:"makerSignature"`
		TakerSignature    string       `json:"takerSignature"`
		BlockDeadline     int          `json:"blockDeadline"`
		SeatNumber        int          `json:"seatNumber"`
		SeatHolder        string       `json:"seatHolder"`
		SeatPercentOfFees int          `json:"seatPercentOfFees"`
	} `json:"matching"`
	MatchingSignature string `json:"matchingSignature"`
	MakerOrderHash    string `json:"makerOrderHash"`
	TakerOrderHash    string `json:"takerOrderHash"`
	ChainId           int    `json:"chainId"`
	Zone              string `json:"zone"`
	To                string `json:"to"`
	Value             int    `json:"value"`
	Data              string `json:"data"`
	Maker             string `json:"maker"`
	Taker             string `json:"taker"`
	InputToken        string `json:"inputToken"`
	InputAmount       string `json:"inputAmount"`
	OutputToken       string `json:"outputToken"`
	OutputAmount      string `json:"outputAmount"`
}

type RfqSolution struct {
	Offerer       string `json:"offerer"`
	InputToken    string `json:"inputToken"`
	InputAmount   string `json:"inputAmount"`
	InputChainId  int    `json:"inputChainId"`
	InputZone     string `json:"inputZone"`
	OutputToken   string `json:"outputToken"`
	OutputAmount  string `json:"outputAmount"`
	OutputChainId int    `json:"outputChainId"`
	OutputZone    string `json:"outputZone"`
	StartTime     string `json:"startTime"`
	EndTime       string `json:"endTime"`
	Salt          string `json:"salt"`
	Counter       int    `json:"counter"`
	ToWithdraw    bool   `json:"toWithdraw"`
}

type RfqQuoteResponse struct {
	RfqId     string       `json:"rfqId"`
	Order     *RfqSolution `json:"order"`
	Signature string       `json:"signature"`
}

type aoriEvent struct {
	name     string
	intentID string
	data     []byte
}

type sendPayload struct {
	Id      int           `json:"id"`
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}
