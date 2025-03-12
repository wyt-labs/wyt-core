package okxswap

type SupportedChain[T any] struct {
	ChainId                T      `json:"chainId"`
	ChainName              string `json:"chainName"`
	DexTokenApproveAddress string `json:"dexTokenApproveAddress"`
}

type Token struct {
	Decimals             string `json:"decimals,omitempty"`
	Decimal              string `json:"decimal,omitempty"`
	TokenContractAddress string `json:"tokenContractAddress,omitempty"`
	TokenLogoUrl         string `json:"tokenLogoUrl,omitempty"`
	TokenName            string `json:"tokenName,omitempty"`
	TokenSymbol          string `json:"tokenSymbol,omitempty"`
	TokenUnitPrice       string `json:"tokenUnitPrice,omitempty"`
}

type CrossChainToken struct {
	ChainId              string `json:"chainId,omitempty"`
	Decimals             int    `json:"decimals,omitempty"`
	TokenContractAddress string `json:"tokenContractAddress,omitempty"`
	TokenName            string `json:"tokenName,omitempty"`
	TokenSymbol          string `json:"tokenSymbol,omitempty"`
}

type CrossChainTokenPair struct {
	FromChainId      string `json:"fromChainId"`
	ToChainId        string `json:"toChainId"`
	FromTokenAddress string `json:"fromTokenAddress"`
	ToTokenAddress   string `json:"toTokenAddress"`
	FromTokenSymbol  string `json:"fromTokenSymbol"`
	ToTokenSymbol    string `json:"toTokenSymbol"`
}

// Get Quotes
type DexProtocol struct {
	DexName string `json:"dexName"`
	Percent string `json:"percent"`
}

type SubRouter struct {
	DexProtocol []DexProtocol `json:"dexProtocol"`
	FromToken   Token         `json:"fromToken"`
	ToToken     Token         `json:"toToken"`
}

type DexRouter struct {
	Router        string      `json:"router"`
	RouterPercent string      `json:"routerPercent"`
	SubRouterList []SubRouter `json:"subRouterList"`
}

type QuoteCompare struct {
	AmountOut string `json:"amountOut"`
	DexLogo   string `json:"dexLogo"`
	DexName   string `json:"dexName"`
	TradeFee  string `json:"tradeFee"`
}

type QuotesData struct {
	ChainId           string         `json:"chainId"`
	DexRouterList     []DexRouter    `json:"dexRouterList"`
	EstimateGasFee    string         `json:"estimateGasFee"`
	EstimateGasFeeUI  string         `json:"estimateGasFeeUI"`
	FromToken         Token          `json:"fromToken"`
	FromTokenAmount   string         `json:"fromTokenAmount"`
	FromTokenUIAmount string         `json:"fromTokenUIAmount"`
	QuoteCompareList  []QuoteCompare `json:"quoteCompareList"`
	ToToken           Token          `json:"toToken"`
	ToTokenAmount     string         `json:"toTokenAmount"`
	ToTokenUIAmount   string         `json:"toTokenUIAmount"`
}

type TransactionData struct {
	Data               string `json:"data"`
	DexContractAddress string `json:"dexContractAddress"`
	GasLimit           string `json:"gasLimit"`
	GasPrice           string `json:"gasPrice"`
}

type RouterResult struct {
	ChainId           string         `json:"chainId"`
	DexRouterList     []DexRouter    `json:"dexRouterList"`
	EstimateGasFee    string         `json:"estimateGasFee"`
	EstimateGasFeeUI  string         `json:"estimateGasFeeUI"`
	FromToken         Token          `json:"fromToken"`
	FromTokenAmount   string         `json:"fromTokenAmount"`
	FromTokenUIAmount string         `json:"fromTokenUIAmount"`
	QuoteCompareList  []QuoteCompare `json:"quoteCompareList"`
	ToToken           Token          `json:"toToken"`
	ToTokenAmount     string         `json:"toTokenAmount"`
	ToTokenUIAmount   string         `json:"toTokenUIAmount"`
}

type Tx struct {
	Data                 string   `json:"data"`
	From                 string   `json:"from"`
	Gas                  string   `json:"gas"`
	GasPrice             string   `json:"gasPrice"`
	MaxPriorityFeePerGas string   `json:"maxPriorityFeePerGas"`
	MinReceiveAmount     string   `json:"minReceiveAmount"`
	SignatureData        []string `json:"signatureData"`
	To                   string   `json:"to"`
	Value                string   `json:"value"`
}

type SwapResponseData struct {
	RouterResult RouterResult `json:"routerResult"`
	Tx           Tx           `json:"tx"`
}

type Router struct {
	BridgeId                  int    `json:"bridgeId"`
	BridgeName                string `json:"bridgeName"`
	CrossChainFee             string `json:"crossChainFee"`
	CrossChainFeeTokenAddress string `json:"crossChainFeeTokenAddress"`
	OtherNativeFee            string `json:"otherNativeFee"`
}

type CrossSubRouter struct {
	DexProtocol []DexProtocol   `json:"dexProtocol"`
	FromToken   CrossChainToken `json:"fromToken"`
	ToToken     CrossChainToken `json:"toToken"`
}

type CrossDexRouter struct {
	Router        string           `json:"router"`
	RouterPercent string           `json:"routerPercent"`
	SubRouterList []CrossSubRouter `json:"subRouterList"`
}

type RouterList struct {
	EstimateTime        string           `json:"estimateTime"`
	EstimateGasFee      string           `json:"estimateGasFee"`
	FromChainNetworkFee string           `json:"fromChainNetworkFee,omitempty"`
	ToChainNetworkFee   string           `json:"toChainNetworkFee,omitempty"`
	MinimumReceived     string           `json:"minimumReceived,omitempty"`
	FromDexRouterList   []CrossDexRouter `json:"fromDexRouterList"`
	NeedApprove         int              `json:"needApprove"`
	Router              Router           `json:"router"`
	ToDexRouterList     []CrossDexRouter `json:"toDexRouterList"`
	ToTokenAmount       string           `json:"toTokenAmount"`
}

type CrossChainQuoteData struct {
	FromChainId        string          `json:"fromChainId"`
	FromToken          CrossChainToken `json:"fromToken"`
	FromTokenAmount    string          `json:"fromTokenAmount"`
	FromTokenUIAmount  string          `json:"fromTokenUIAmount"`
	FromTokenUnitPrice float64         `json:"fromTokenUnitPrice"`
	RouterList         []RouterList    `json:"routerList"`
	EstimateGasFeeUI   string          `json:"estimateGasFeeUI"`
	ToChainId          string          `json:"toChainId"`
	ToToken            CrossChainToken `json:"toToken"`
	ToTokenUIAmount    string          `json:"toTokenUIAmount"`
	ToTokenUnitPrice   float64         `json:"toTokenUnitPrice"`
}

type CrossChainTx struct {
	FromTokenAmount string `json:"fromTokenAmount"`
	Router          Router `json:"router"`
	ToTokenAmount   string `json:"toTokenAmount"`
	MinmumReceive   string `json:"minmumReceive"`
	Tx              Tx     `json:"tx"`
}

type TokenPriceReq struct {
	ChainIndex   string `json:"chainIndex"`
	TokenAddress string `json:"tokenAddress"`
}

type TokenPrice struct {
	ChainIndex   string `json:"chainIndex"`
	TokenAddress string `json:"tokenAddress"`
	Time         string `json:"time"`
	Price        string `json:"price"`
}

type OkxApiResponse[T any] struct {
	Code string `json:"code"`
	Data []T    `json:"data"`
	Msg  string `json:"msg"`
}

type MessariErrorResponse struct {
	Status struct {
		Elapsed      int    `json:"elapsed"`
		Timestamp    string `json:"timestamp"`
		ErrorCode    int    `json:"error_code"`
		ErrorMessage string `json:"error_message"`
	} `json:"status"`
}
