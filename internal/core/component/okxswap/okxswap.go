package okxswap

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/wyt-labs/wyt-core/internal/core/component/httpclient"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/internal/pkg/errcode"
	"github.com/wyt-labs/wyt-core/pkg/basic"
	"github.com/wyt-labs/wyt-core/pkg/cache"
)

func init() {
	basic.RegisterComponents(NewOkxSwapApi)
}

type OkxSwapApi struct {
	baseComponent *base.Component
	apiClient     *httpclient.Client
}

func NewOkxSwapApi(baseComponent *base.Component) (*OkxSwapApi, error) {
	client, err := httpclient.NewHttpClient(
		httpclient.WithBaseURL(baseComponent.Config.Okx.Endpoint),
	)
	if err != nil {
		baseComponent.Logger.WithField("err", err).Error("failed to create http client")
		return nil, err
	}
	return &OkxSwapApi{
		baseComponent: baseComponent,
		apiClient:     client,
	}, nil
}

func (oapi *OkxSwapApi) preReqGenHeader(method, path string, body string) map[string]string {
	// 获取当前时间的ISO字符串
	now := time.Now().UTC()
	isoString := now.Format(time.RFC3339)
	secretKey := oapi.baseComponent.Config.Okx.SecretKey

	// 生成签名
	message := isoString + method + path + body
	hash := hmac.New(sha256.New, []byte(secretKey))
	hash.Write([]byte(message))
	sign := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	headers := map[string]string{
		"Content-Type":         "application/json",
		"OK-ACCESS-PROJECT":    oapi.baseComponent.Config.Okx.ProjectId,
		"OK-ACCESS-KEY":        oapi.baseComponent.Config.Okx.APIKey,
		"OK-ACCESS-PASSPHRASE": oapi.baseComponent.Config.Okx.Passphrase,
		"OK-ACCESS-SIGN":       sign,
		"OK-ACCESS-TIMESTAMP":  isoString,
	}
	return headers
}

// solana 501
func (oapi *OkxSwapApi) GetSupportedChains(chainId int, isCrossChain bool) ([]SupportedChain[string], error) {
	var path string
	if isCrossChain {
		path = "/api/v5/dex/cross-chain/supported/chain"
	} else {
		path = "/api/v5/dex/aggregator/supported/chain"
	}
	queries := make(map[string]string, 0)
	if chainId != 0 {
		queries["chainId"] = fmt.Sprint(chainId)
	}
	parsedUrl, err := oapi.apiClient.ParseURL(path, queries)
	if err != nil {
		return nil, err
	}
	pathQuery := strings.TrimPrefix(parsedUrl, oapi.apiClient.GetBaseURL())
	headers := oapi.preReqGenHeader("GET", pathQuery, "")

	resp, err := oapi.apiClient.GetV2(parsedUrl, headers)
	if err != nil {
		oapi.baseComponent.Logger.WithField("err", err).Error("failed to get supported chains")
		return nil, err
	}
	if isCrossChain {
		var response OkxApiResponse[SupportedChain[string]]
		if err := json.Unmarshal(resp, &response); err != nil {
			return nil, err
		}
		return response.Data, nil
	} else {
		var response OkxApiResponse[SupportedChain[int]]
		if err := json.Unmarshal(resp, &response); err != nil {
			return nil, err
		}
		chains := lo.Map(response.Data, func(item SupportedChain[int], index int) SupportedChain[string] {
			return SupportedChain[string]{
				ChainId:                fmt.Sprint(item.ChainId),
				ChainName:              item.ChainName,
				DexTokenApproveAddress: item.DexTokenApproveAddress,
			}
		})
		return chains, nil
	}
}

var TokenNs = "tokens"

// GetTokens 获取source chainId支持的token列表
//
// chainId: source chainId, 0: all chains,
func (oapi *OkxSwapApi) GetTokens(chainId int) ([]Token, error) {
	path := "/api/v5/dex/aggregator/all-tokens"
	oapi.baseComponent.Logger.Info("get supported all tokens, ", "chainId: ", chainId)
	tokens, ok := cache.GetFromMemCache[[]Token](oapi.baseComponent.MemCache, TokenNs, fmt.Sprint(chainId))
	if ok {
		oapi.baseComponent.Logger.WithField("chainId", chainId).Info("get tokens from memcache")
		return tokens, nil
	}
	queries := make(map[string]string, 0)
	if chainId != 0 {
		queries["chainId"] = fmt.Sprint(chainId)
	}
	parsedUrl, err := oapi.apiClient.ParseURL(path, queries)
	if err != nil {
		return nil, err
	}
	pathQuery := strings.TrimPrefix(parsedUrl, oapi.apiClient.GetBaseURL())
	headers := oapi.preReqGenHeader("GET", pathQuery, "")

	resp, err := oapi.apiClient.GetV2(parsedUrl, headers)
	if err != nil {
		oapi.baseComponent.Logger.WithField("err", err).Error("failed to get supported chains")
		return nil, err
	}
	var response OkxApiResponse[Token]
	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, err
	}
	// put data to memcache
	cache.PutToMemCache(oapi.baseComponent.MemCache, TokenNs, fmt.Sprint(chainId), response.Data)
	oapi.baseComponent.Logger.Info("get supported all tokens successfully")
	return response.Data, nil
}

// GetCrossChainTokens, List of tokens available for traded directly across the cross-chain bridge.
func (oapi *OkxSwapApi) GetCrossChainTokens(chainId int) ([]CrossChainToken, error) {
	path := "/api/v5/dex/cross-chain/supported/tokens"
	queries := make(map[string]string, 0)
	if chainId != 0 {
		queries["chainId"] = fmt.Sprint(chainId)
	}
	parsedUrl, err := oapi.apiClient.ParseURL(path, queries)
	if err != nil {
		return nil, err
	}
	pathQuery := strings.TrimPrefix(parsedUrl, oapi.apiClient.GetBaseURL())
	headers := oapi.preReqGenHeader("GET", pathQuery, "")

	resp, err := oapi.apiClient.GetV2(parsedUrl, headers)
	if err != nil {
		oapi.baseComponent.Logger.WithField("err", err).Error("failed to get supported chains")
		return nil, err
	}
	var response OkxApiResponse[CrossChainToken]
	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

// GetCrossChainTokens, List of tokens available for traded directly across the cross-chain bridge.
//
// fromChainId: from chainId
func (oapi *OkxSwapApi) GetBridgeTokensPairs(fromChainId int) ([]CrossChainTokenPair, error) {
	path := "/api/v5/dex/cross-chain/supported/bridge-tokens-pairs"
	queries := make(map[string]string, 0)
	if fromChainId != 0 {
		queries["fromChainId"] = fmt.Sprint(fromChainId)
	}
	parsedUrl, err := oapi.apiClient.ParseURL(path, queries)
	if err != nil {
		return nil, err
	}
	pathQuery := strings.TrimPrefix(parsedUrl, oapi.apiClient.GetBaseURL())
	headers := oapi.preReqGenHeader("GET", pathQuery, "")

	resp, err := oapi.apiClient.GetV2(parsedUrl, headers)
	if err != nil {
		oapi.baseComponent.Logger.WithField("err", err).Error("failed to get supported chains")
		return nil, err
	}
	var response OkxApiResponse[CrossChainTokenPair]
	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, err
	}
	// 检查是否返回错误信息
	if response.Code != "0" {
		code, err := strconv.ParseUint(response.Code, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid error code: %s", response.Code)
		}
		return nil, errcode.NewCustomError(uint32(code), response.Msg)
	}

	if len(response.Data) == 0 {
		return nil, errcode.NewCustomError(uint32(10001), response.Msg)
	}
	return response.Data, nil
}

// GetCrossChainQuote, Find the best route for a cross-chain swap through OKX’s DEX cross-chain aggregator.
//
// fromChainId: from chainId
func (oapi *OkxSwapApi) GetCrossChainQuote(
	fromChainId int,
	toChainId int,
	fromTokenAddress string,
	toTokenAddress string,
	amount string,
	slippage string,
) ([]CrossChainQuoteData, error) {
	path := "/api/v5/dex/cross-chain/quote"
	decimal, err := oapi.getDecimals(fromChainId, fromTokenAddress)
	if err != nil {
		return nil, err
	}
	amountInWei, err := UIAmount2ContractAmount(amount, decimal)
	if err != nil {
		return nil, err
	}
	queries := make(map[string]string, 0)
	if fromChainId != 0 {
		queries["fromChainId"] = fmt.Sprint(fromChainId)
	}
	if toChainId != 0 {
		queries["toChainId"] = fmt.Sprint(toChainId)
	}
	queries["fromTokenAddress"] = fromTokenAddress
	queries["toTokenAddress"] = toTokenAddress
	queries["amount"] = amountInWei
	queries["slippage"] = slippage
	parsedUrl, err := oapi.apiClient.ParseURL(path, queries)
	if err != nil {
		return nil, err
	}
	pathQuery := strings.TrimPrefix(parsedUrl, oapi.apiClient.GetBaseURL())
	headers := oapi.preReqGenHeader("GET", pathQuery, "")

	resp, err := oapi.apiClient.GetV2(parsedUrl, headers)
	if err != nil {
		oapi.baseComponent.Logger.WithField("err", err).Error("failed to get supported chains")
		return nil, err
	}
	oapi.baseComponent.Logger.WithField("resp", string(resp)).Info("cross chain quote")
	var response OkxApiResponse[CrossChainQuoteData]
	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, err
	}

	// 检查是否返回错误信息
	if response.Code != "0" {
		code, err := strconv.ParseUint(response.Code, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid error code: %s", response.Code)
		}
		return nil, errcode.NewCustomError(uint32(code), response.Msg)
	}

	if len(response.Data) == 0 {
		return nil, errcode.NewCustomError(uint32(10001), response.Msg)
	}

	response.Data[0].FromTokenUIAmount, err = ContractAmount2UIAmount(
		response.Data[0].FromTokenAmount,
		strconv.Itoa(response.Data[0].FromToken.Decimals),
	)
	if err != nil {
		oapi.baseComponent.Logger.WithField("err", err).Error("failed to convert contract amount to UI amount")
	}
	// 价格
	fp, err := oapi.GetTokenPriceV2(response.Data[0].FromToken.TokenSymbol)
	if err != nil {
		oapi.baseComponent.Logger.WithField("err", err).Error("failed to get token price")
	}
	response.Data[0].FromTokenUnitPrice = fp
	response.Data[0].ToTokenUIAmount, err = ContractAmount2UIAmount(
		response.Data[0].RouterList[0].ToTokenAmount,
		strconv.Itoa(response.Data[0].ToToken.Decimals),
	)
	if err != nil {
		oapi.baseComponent.Logger.WithField("err", err).Error("failed to convert contract amount to UI amount")
	}
	gasFee, err := ContractAmount2UIAmount(
		response.Data[0].RouterList[0].EstimateGasFee,
		strconv.Itoa(response.Data[0].FromToken.Decimals),
	)
	if err != nil {
		oapi.baseComponent.Logger.WithField("err", err).Error("failed to convert contract amount to UI amount")
	}
	if fp != 0 {
		gasFeeFloat, err := strconv.ParseFloat(gasFee, 64)
		if err != nil {
			oapi.baseComponent.Logger.WithField("err", err).Error("failed to convert gas fee to float")
		}
		response.Data[0].EstimateGasFeeUI = strconv.FormatFloat(gasFeeFloat*fp, 'f', -1, 64)
	} else {
		response.Data[0].EstimateGasFeeUI = gasFee
	}

	// 价格
	fp, err = oapi.GetTokenPriceV2(response.Data[0].ToToken.TokenSymbol)
	if err != nil {
		oapi.baseComponent.Logger.WithField("err", err).Error("failed to get token price")
	}
	response.Data[0].ToTokenUnitPrice = fp
	return response.Data, nil
}

// GetQuotes fromTokenAddress, token address: 0xdac17f958d2ee523a2206206994597c13d831ec7
// toTokenAddress, token address: 0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48
// dexIds, 1,501,180
func (oapi *OkxSwapApi) GetQuotes(chainId int, amount string, fromTokenAddress, toTokenAddress string, dexIds string) ([]QuotesData, error) {
	path := "/api/v5/dex/aggregator/quote"
	decimal, err := oapi.getDecimals(chainId, fromTokenAddress)
	if err != nil {
		return nil, err
	}
	amountInWei, err := UIAmount2ContractAmount(amount, decimal)
	if err != nil {
		return nil, err
	}
	queries := make(map[string]string, 0)
	if chainId != 0 {
		queries["chainId"] = fmt.Sprint(chainId)
	}
	queries["amount"] = amountInWei
	queries["fromTokenAddress"] = fromTokenAddress
	queries["toTokenAddress"] = toTokenAddress
	if dexIds != "" {
		queries["dexIds"] = dexIds
	}
	parsedUrl, err := oapi.apiClient.ParseURL(path, queries)
	if err != nil {
		return nil, err
	}
	pathQuery := strings.TrimPrefix(parsedUrl, oapi.apiClient.GetBaseURL())
	headers := oapi.preReqGenHeader("GET", pathQuery, "")

	resp, err := oapi.apiClient.GetV2(parsedUrl, headers)
	if err != nil {
		oapi.baseComponent.Logger.WithField("err", err).Error("failed to get supported chains")
		return nil, err
	}
	var response OkxApiResponse[QuotesData]
	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, err
	}
	// 检查是否返回错误信息
	if response.Code != "0" {
		code, err := strconv.ParseUint(response.Code, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid error code: %s", response.Code)
		}
		return nil, errcode.NewCustomError(uint32(code), response.Msg)
	}

	if len(response.Data) == 0 {
		return nil, errcode.NewCustomError(uint32(10001), response.Msg)
	}
	for _, item := range response.Data {
		amtFrom, err := ContractAmount2UIAmount(item.FromTokenAmount, item.FromToken.Decimal)
		if err != nil {
			oapi.baseComponent.Logger.WithField("err", err).Error("failed to convert contract amount to UI amount")
			continue
		}
		response.Data[0].FromTokenUIAmount = amtFrom
		amtTo, err := ContractAmount2UIAmount(item.ToTokenAmount, item.ToToken.Decimal)
		if err != nil {
			oapi.baseComponent.Logger.WithField("err", err).Error("failed to convert contract amount to UI amount")
			continue
		}
		response.Data[0].ToTokenUIAmount = amtTo
		// 估算手续费
		gas, err := ContractAmount2UIAmount(item.EstimateGasFee, "6")
		if err != nil {
			oapi.baseComponent.Logger.WithField("err", err).Error("failed to convert contract amount to UI amount")
			continue
		}
		response.Data[0].EstimateGasFeeUI = gas
	}
	return response.Data, nil
}

func (oapi *OkxSwapApi) ApproveTransaction(chainId int, tokenContractAddress string, approveAmount string) ([]TransactionData, error) {
	path := "/api/v5/dex/aggregator/approve-transaction"
	queries := make(map[string]string, 0)
	if chainId != 0 {
		queries["chainId"] = fmt.Sprint(chainId)
	}
	queries["tokenContractAddress"] = tokenContractAddress
	queries["approveAmount"] = approveAmount
	parsedUrl, err := oapi.apiClient.ParseURL(path, queries)
	if err != nil {
		return nil, err
	}
	pathQuery := strings.TrimPrefix(parsedUrl, oapi.apiClient.GetBaseURL())
	headers := oapi.preReqGenHeader("GET", pathQuery, "")

	resp, err := oapi.apiClient.GetV2(parsedUrl, headers)
	if err != nil {
		oapi.baseComponent.Logger.WithField("err", err).Error("failed to get supported chains")
		return nil, err
	}
	var response OkxApiResponse[TransactionData]
	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (oapi *OkxSwapApi) getDecimals(chainId int, tokenContractAddress string) (string, error) {
	tokens, err := oapi.GetTokens(chainId)
	if err != nil {
		return "", err
	}
	for _, token := range tokens {
		if token.TokenContractAddress == tokenContractAddress {
			return token.Decimals, nil
		}
	}
	return "", fmt.Errorf("token not found")
}

func (oapi *OkxSwapApi) GetToken(chainId int, tokenContractAddress string) (*Token, error) {
	tokens, err := oapi.GetTokens(chainId)
	if err != nil {
		return nil, err
	}
	for _, token := range tokens {
		if token.TokenContractAddress == tokenContractAddress {
			return &token, nil
		}
	}
	return nil, fmt.Errorf("token not found")
}

// "amount": "0.1", decimal: "18"
func UIAmount2ContractAmount(amountStr string, decimalStr string) (string, error) {
	// 将 decimal 转换成 big.Int
	decimal, ok := new(big.Int).SetString(decimalStr, 10)
	if !ok {
		return "", fmt.Errorf("invalid decimal value: %s", decimalStr)
	}

	// 将 amount 转换成 float64
	amountFloat, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return "", fmt.Errorf("invalid amount value: %s", amountStr)
	}

	// 将 amountFloat 转换成 big.Float
	amountBigFloat := new(big.Float).SetFloat64(amountFloat)

	// 计算 10^decimal
	ten := big.NewInt(10)
	decimalPower := new(big.Int).Exp(ten, decimal, nil)

	// 将 decimalPower 转换成 big.Float
	decimalPowerFloat := new(big.Float).SetInt(decimalPower)

	// 计算 amount * 10^decimal
	amountInWeiFloat := new(big.Float).Mul(amountBigFloat, decimalPowerFloat)

	// 将结果转换成 big.Int
	amountInWei := new(big.Int)
	amountInWeiFloat.Int(amountInWei)

	return amountInWei.String(), nil
}

// "amount": "100000000000000000", decimal: "18"
func ContractAmount2UIAmount(toTokenAmount string, decimalStr string) (string, error) {
	// 将 decimal 转换成 big.Int
	decimal, ok := new(big.Int).SetString(decimalStr, 10)
	if !ok {
		return "", fmt.Errorf("invalid decimal value: %s", decimalStr)
	}

	// 将 toTokenAmount 转换成 big.Int
	toTokenAmountBigInt, ok := new(big.Int).SetString(toTokenAmount, 10)
	if !ok {
		return "", fmt.Errorf("invalid toTokenAmount value: %s", toTokenAmount)
	}

	// 计算 10^decimal
	ten := big.NewInt(10)
	decimalPower := new(big.Int).Exp(ten, decimal, nil)

	// 将 decimalPower 转换成 big.Float
	decimalPowerFloat := new(big.Float).SetInt(decimalPower)

	// 将 toTokenAmountBigInt 转换成 big.Float
	toTokenAmountBigFloat := new(big.Float).SetInt(toTokenAmountBigInt)

	// 计算 toTokenAmount / 10^decimal
	toTokenAmountFloat := new(big.Float).Quo(toTokenAmountBigFloat, decimalPowerFloat)

	// 将结果转换成 float64
	toTokenAmountFloat64, _ := toTokenAmountFloat.Float64()

	// 将 float64 转换成字符串
	return strconv.FormatFloat(toTokenAmountFloat64, 'f', -1, 64), nil
}

func (oapi *OkxSwapApi) Swap(chainId int, amount string, fromTokenAddress string, toTokenAddress string, userWalletAddress, slippage string, swapReceiverAddress string) ([]SwapResponseData, error) {
	path := "/api/v5/dex/aggregator/swap"
	queries := make(map[string]string, 0)
	if chainId != 0 {
		queries["chainId"] = fmt.Sprint(chainId)
	}
	decimal, err := oapi.getDecimals(chainId, fromTokenAddress)
	if err != nil {
		return nil, err
	}
	amountInWei, err := UIAmount2ContractAmount(amount, decimal)
	if err != nil {
		return nil, err
	}
	queries["amount"] = amountInWei
	queries["fromTokenAddress"] = fromTokenAddress
	queries["toTokenAddress"] = toTokenAddress
	queries["userWalletAddress"] = userWalletAddress
	queries["slippage"] = slippage
	queries["swapReceiverAddress"] = swapReceiverAddress
	parsedUrl, err := oapi.apiClient.ParseURL(path, queries)
	if err != nil {
		return nil, err
	}
	pathQuery := strings.TrimPrefix(parsedUrl, oapi.apiClient.GetBaseURL())
	headers := oapi.preReqGenHeader("GET", pathQuery, "")

	resp, err := oapi.apiClient.GetV2(parsedUrl, headers)
	if err != nil {
		oapi.baseComponent.Logger.WithField("err", err).Error("failed to get supported chains")
		return nil, err
	}
	var response OkxApiResponse[SwapResponseData]
	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, err
	}
	// 检查是否返回错误信息
	if response.Code != "0" {
		code, err := strconv.ParseUint(response.Code, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid error code: %s", response.Code)
		}
		return nil, errcode.NewCustomError(uint32(code), response.Msg)
	}

	if len(response.Data) == 0 {
		return nil, errcode.NewCustomError(uint32(10001), response.Msg)
	}
	data := response.Data
	for _, item := range data {
		amtFrom, err := ContractAmount2UIAmount(item.RouterResult.FromTokenAmount, item.RouterResult.FromToken.Decimal)
		if err != nil {
			oapi.baseComponent.Logger.WithField("err", err).Error("failed to convert contract amount to UI amount")
			continue
		}
		data[0].RouterResult.FromTokenUIAmount = amtFrom
		amtTo, err := ContractAmount2UIAmount(item.RouterResult.ToTokenAmount, item.RouterResult.ToToken.Decimal)
		if err != nil {
			oapi.baseComponent.Logger.WithField("err", err).Error("failed to convert contract amount to UI amount")
			continue
		}
		data[0].RouterResult.ToTokenUIAmount = amtTo
		// 估算手续费
		gas, err := ContractAmount2UIAmount(item.RouterResult.EstimateGasFee, item.RouterResult.FromToken.Decimal)
		if err != nil {
			oapi.baseComponent.Logger.WithField("err", err).Error("failed to convert contract amount to UI amount")
			continue
		}
		data[0].RouterResult.EstimateGasFeeUI = gas
	}

	return data, nil
}

func (oapi *OkxSwapApi) CrosschainSwap(fromChainId int, toChainId int, amount string, fromTokenAddress string, toTokenAddress string, userWalletAddress, slippage string) ([]CrossChainTx, error) {
	path := "/api/v5/dex/cross-chain/build-tx"
	queries := make(map[string]string, 0)
	if fromChainId != 0 {
		queries["fromChainId"] = fmt.Sprint(fromChainId)
	}
	if toChainId != 0 {
		queries["toChainId"] = fmt.Sprint(toChainId)
	}
	queries["amount"] = amount
	queries["fromTokenAddress"] = fromTokenAddress
	queries["toTokenAddress"] = toTokenAddress
	queries["userWalletAddress"] = userWalletAddress
	queries["slippage"] = slippage
	parsedUrl, err := oapi.apiClient.ParseURL(path, queries)
	if err != nil {
		return nil, err
	}
	pathQuery := strings.TrimPrefix(parsedUrl, oapi.apiClient.GetBaseURL())
	headers := oapi.preReqGenHeader("GET", pathQuery, "")

	resp, err := oapi.apiClient.GetV2(parsedUrl, headers)
	if err != nil {
		oapi.baseComponent.Logger.WithField("err", err).Error("failed to get supported chains")
		return nil, err
	}
	var response OkxApiResponse[CrossChainTx]
	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

// get token current-price
func (oapi *OkxSwapApi) GetTokenPrice(chainId int, tokenContractAddress string) ([]TokenPrice, error) {
	path := "/api/v5/wallet/token/current-price"
	parsedUrl, err := oapi.apiClient.ParseURL(path, nil)
	if err != nil {
		return nil, err
	}
	pathQuery := strings.TrimPrefix(parsedUrl, oapi.apiClient.GetBaseURL())
	bds := make([]TokenPriceReq, 0)
	bds = append(bds, TokenPriceReq{
		ChainIndex:   fmt.Sprint(chainId),
		TokenAddress: tokenContractAddress,
	})
	bdsJson, _ := json.Marshal(bds)
	headers := oapi.preReqGenHeader("POST", pathQuery, string(bdsJson))
	resp, err := oapi.apiClient.Post(path, bdsJson, headers, nil)
	if err != nil {
		oapi.baseComponent.Logger.WithField("err", err).Error("failed to get supported chains")
		return nil, err
	}
	var response OkxApiResponse[TokenPrice]
	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

// get token current-price
func (oapi *OkxSwapApi) GetTokenPriceV2(tokenSymbol string) (float64, error) {
	client, err := httpclient.NewHttpClient(
		httpclient.WithBaseURL("https://data.messari.io"),
	)
	if err != nil {
		oapi.baseComponent.Logger.WithField("err", err).Error("failed to create http client")
		return 0, err
	}
	path := "/api/v1/assets/" + tokenSymbol + "/metrics"
	resp, err := client.Get(path, nil, nil)
	if err != nil {
		oapi.baseComponent.Logger.WithField("err", err).Error("failed to get supported chains")
		return 0, err
	}

	// 检查是否返回错误信息
	var errorResponse MessariErrorResponse
	if err := json.Unmarshal(resp, &errorResponse); err == nil && errorResponse.Status.ErrorCode != 0 {
		return 0, fmt.Errorf("API error: %s (code: %d)", errorResponse.Status.ErrorMessage, errorResponse.Status.ErrorCode)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(resp, &response); err != nil {
		return 0, err
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid data format")
	}

	marketData, ok := data["market_data"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid market_data format")
	}

	priceUSD, ok := marketData["price_usd"].(float64)
	if !ok {
		return 0, fmt.Errorf("price_usd not found or invalid format")
	}

	return priceUSD, nil
}
