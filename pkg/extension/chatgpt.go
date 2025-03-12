package extension

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wyt-labs/wyt-core/internal/pkg/azure"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/internal/pkg/config"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/wyt-labs/wyt-core/internal/core/component/datapuller"
	dpmodel "github.com/wyt-labs/wyt-core/internal/core/component/datapuller/model"
	"github.com/wyt-labs/wyt-core/internal/core/component/httpclient"
	"github.com/wyt-labs/wyt-core/internal/core/model"
	"github.com/wyt-labs/wyt-core/internal/pkg/entity"
)

type ChatgptConfig struct {
	Endpoint string
	// EndpointFull with api version and model name baked in
	EndpointFull    string
	APIKey          string
	Model           string
	Temperature     float32
	PresencePenalty float32
}

type WytAIClient struct {
	BaseComponent *base.Component
	ProjectId     string
	UniProjectId  string
	APIKey        string
	APIClient     *httpclient.Client
}

func NewWYTAIClient(baseComponent *base.Component, envUrl string, projectId, uniProjectId, apiKey string) (*WytAIClient, error) {
	client, err := httpclient.NewHttpClient(
		httpclient.WithBaseURL(envUrl),
	)
	if err != nil {
		baseComponent.Logger.Error("failed to create http client", err)
		return nil, err
	}
	return &WytAIClient{
		BaseComponent: baseComponent,
		ProjectId:     projectId,
		UniProjectId:  uniProjectId,
		APIKey:        apiKey,
		APIClient:     client,
	}, nil
}

type ChatgptDriver struct {
	cfg    *ChatgptConfig
	client *azopenai.Client
	//azureRestClient supports json structured payload
	azureRestClient *azure.ChatAPI
	pumpDataService *datapuller.PumpDataService
	wytAIClient     *WytAIClient
}

func NewChatgptDriver(cfg *ChatgptConfig, baseComponent *base.Component, pumpDataService *datapuller.PumpDataService) (*ChatgptDriver, error) {
	keyCredential := azcore.NewKeyCredential(cfg.APIKey)
	client, err := azopenai.NewClientWithKeyCredential(cfg.Endpoint, keyCredential, nil)
	azureRestClient := azure.NewAzureChatAPI(
		cfg.EndpointFull,
		cfg.APIKey,
		azure.RequestResponseFormat{
			Type:       "json_schema",
			JsonSchema: config.ChatSchemaWithStructureOutputModeEnabled,
		},
		lo.Map(SupportedFunctions(), func(item azopenai.FunctionDefinition, _ int) azure.RequestTool {
			return azure.RequestTool{
				Type:     "function",
				Function: item,
			}
		}))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create chatgpt client")
	}
	var envUrl, pjId, uniPjId, apiKey string
	if baseComponent.Config.AIBackend.AIEnv == "prod" {
		envUrl = baseComponent.Config.AIBackend.Endpoint
		pjId = baseComponent.Config.AIBackend.ProjectId
		uniPjId = baseComponent.Config.AIBackend.UniProjectId
		apiKey = baseComponent.Config.AIBackend.APIKey
	} else if baseComponent.Config.AIBackend.AIEnv == "dev" {
		envUrl = baseComponent.Config.AIBackend.DevEndpoint
		pjId = baseComponent.Config.AIBackend.DevProjectId
		uniPjId = baseComponent.Config.AIBackend.DevUniProjectId
		apiKey = baseComponent.Config.AIBackend.DevAPIKey
	}
	wytAIClient, err := NewWYTAIClient(
		baseComponent,
		envUrl,
		pjId,
		uniPjId,
		apiKey,
	)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create wyt ai client")
	}

	return &ChatgptDriver{
		cfg:             cfg,
		client:          client,
		azureRestClient: azureRestClient,
		pumpDataService: pumpDataService,
		wytAIClient:     wytAIClient,
	}, nil
}

type ChatgptMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DocsFaq connect wyt ai backend
func (d *ChatgptDriver) DocsFaq(ctx context.Context, msgs []*ChatgptMsg, projectId string) (string, *model.FuncCallingRet, error) {
	var pjId string
	if projectId == "" {
		pjId = d.wytAIClient.ProjectId
	} else {
		pjId = projectId
	}
	headers := map[string]string{
		"Content-Type": "application/json",
		"apikey":       d.wytAIClient.APIKey,
	}
	msgs2post := lo.Map(msgs, func(item *ChatgptMsg, index int) entity.Message {
		return entity.Message{
			Type: "text",
			Text: item.Content,
		}
	})
	pjRelated := []string{}
	// history messages
	hisMsgs := []entity.HistoryMessage{}
	history := entity.History{Messages: hisMsgs}
	historys := []entity.History{history}

	body := entity.Request{
		Messages:        msgs2post,
		RelatedProjects: pjRelated,
		History:         historys,
	}
	bodyBytes, _ := json.Marshal(body)
	// request wyt ai backend
	d.wytAIClient.BaseComponent.Logger.Info("project id:", pjId, ",WYT AI request:", string(bodyBytes))
	resp, err := d.wytAIClient.APIClient.Post("/doc-search/search/"+pjId, bodyBytes, headers, nil)
	if err != nil {
		d.wytAIClient.BaseComponent.Logger.Error("failed to post to wyt ai ", err)
		return "", nil, err
	}
	var res entity.Response
	err = json.Unmarshal(resp, &res)
	if err != nil {
		d.wytAIClient.BaseComponent.Logger.Error("failed to unmarshal wyt ai response ", err)
		return "", nil, err
	}
	// check any unresolved tool results
	for _, rest := range res.ToolResults {
		if !rest.Result.Resolved {
			// handle unresolved tool results
			bytesResult, err := json.Marshal(rest.Result.Result)
			if err != nil {
				d.wytAIClient.BaseComponent.Logger.Error("failed to unmarshal wyt ai response", err)
				return "", nil, err
			}
			d.wytAIClient.BaseComponent.Logger.Info("wyt ai response(function call):", string(bytesResult))
			callRet, err := d.handleFunctionCall(ctx, &azopenai.FunctionCall{Name: &rest.ToolName, Arguments: to.Ptr(string(bytesResult))})
			if err != nil {
				return "", nil, err
			}
			return "", callRet, nil
		} else {
			ret, err := mapRemoteFunctionToFuncCallingRet(
				rest.ToolName,
				rest.Result.Result,
			)
			if err != nil {
				d.wytAIClient.BaseComponent.Logger.Error("failed to map remote function to func calling ret", err)
				return "", nil, err
			}
			jsonResult, _ := json.Marshal(ret)
			d.wytAIClient.BaseComponent.Logger.Info("wyt ai response(function call):", string(jsonResult))
			if ret.FCType == model.FCUniswap {
				return "", &model.FuncCallingRet{
					FCType:          model.FCUniswap,
					FCUniswapResult: ret.FCUniswapResult,
				}, nil
			}
			return "", ret, nil
		}
	}
	bytes, err := json.Marshal(res.Object)
	if err != nil {
		d.wytAIClient.BaseComponent.Logger.Error("failed to marshal wyt ai response", err)
		return "", nil, err
	}
	d.wytAIClient.BaseComponent.Logger.Info("WYT AI response:", string(bytes))
	return string(bytes), nil, nil
}

func (d *ChatgptDriver) ChatCompletions(msgs []*ChatgptMsg) (string, error) {
	res, err := d.client.GetChatCompletions(
		context.Background(),
		azopenai.ChatCompletionsOptions{
			Messages: lo.Map(msgs, func(item *ChatgptMsg, index int) azopenai.ChatRequestMessageClassification {
				if item.Role == "user" {
					return &azopenai.ChatRequestUserMessage{
						Content: azopenai.NewChatRequestUserMessageContent(item.Content),
					}
				} else {
					return &azopenai.ChatRequestSystemMessage{
						Content: to.Ptr(item.Content),
					}
				}
			}),

			DeploymentName:  &d.cfg.Model,
			Temperature:     &d.cfg.Temperature,
			PresencePenalty: &d.cfg.PresencePenalty,
		},
		&azopenai.GetChatCompletionsOptions{},
	)
	if err != nil {
		return "", err
	}

	if len(res.Choices) == 0 {
		return "", errors.New("not found choices")
	}

	return *res.Choices[0].Message.Content, nil
}

func (d *ChatgptDriver) ChatCompletionsWithFuncCall(msgs []*ChatgptMsg) (string, *model.FuncCallingRet, error) {
	ctx := context.Background()
	res, err := d.client.GetChatCompletions(
		ctx,
		azopenai.ChatCompletionsOptions{
			Messages: lo.Map(msgs, func(item *ChatgptMsg, index int) azopenai.ChatRequestMessageClassification {
				if item.Role == "user" {
					return &azopenai.ChatRequestUserMessage{
						Content: azopenai.NewChatRequestUserMessageContent(item.Content),
					}
				} else {
					return &azopenai.ChatRequestSystemMessage{
						Content: to.Ptr(item.Content),
					}
				}
			}),
			DeploymentName:  &d.cfg.Model,
			Temperature:     &d.cfg.Temperature,
			PresencePenalty: &d.cfg.PresencePenalty,
			Functions:       SupportedFunctions(),
			FunctionCall:    &azopenai.ChatCompletionsOptionsFunctionCall{Value: to.Ptr("auto")},
		},
		&azopenai.GetChatCompletionsOptions{},
	)
	if err != nil {
		return "", nil, err
	}

	if len(res.Choices) == 0 {
		return "", nil, errors.New("not found choices")
	}

	fmt.Printf("res: %+v\n", res.Choices[0].Message.Content)

	if res.Choices[0].Message.FunctionCall != nil {
		fmt.Printf("function call, Name %+v, Arguments %+v\n", *res.Choices[0].Message.FunctionCall.Name, *res.Choices[0].Message.FunctionCall.Arguments)
		fcret, err := d.handleFunctionCall(ctx, res.Choices[0].Message.FunctionCall)
		if err != nil {
			return "", nil, err
		}
		return "", fcret, nil
	}

	return *res.Choices[0].Message.Content, nil, nil
}

// ChatWithStructuredOutput uses custom azure restful api making a direct API call to the azure api endpoint.
// This is useful since official Azure SDK doesn't support 2024-08-01 version yet.
func (d *ChatgptDriver) ChatWithStructuredOutput(msgs []*ChatgptMsg) (string, *model.FuncCallingRet, error) {
	ctx := context.Background()
	res, err := d.azureRestClient.Chat(lo.Map(msgs, func(item *ChatgptMsg, _ int) azure.Message {
		return azure.Message{
			Role:    item.Role,
			Content: item.Content,
		}
	}))

	if err != nil {
		return "", nil, err
	}

	if len(res.Choices) == 0 {
		return "", nil, errors.New("not found choices")
	}

	if len(res.Choices[0].Message.ToolCalls) > 0 {
		toolCall := res.Choices[0].Message.ToolCalls[0]
		fcret, err := d.handleFunctionCall(ctx, &azopenai.FunctionCall{
			Name:      &toolCall.Function.Name,
			Arguments: &toolCall.Function.Arguments,
		})
		if err != nil {
			return "", nil, err
		}

		// In the latest implementation, the function call response is returned even it is a function call
		//TODO: return the content
		return "", fcret, nil
	}

	return res.Choices[0].Message.Content, nil, nil
}

func (d *ChatgptDriver) handleFunctionCall(ctx context.Context, functionCall *azopenai.FunctionCall) (*model.FuncCallingRet, error) {
	if functionCall.Name == nil || functionCall.Arguments == nil {
		return nil, fmt.Errorf("invalid function call response")
	}
	switch *functionCall.Name {
	case "swap":
		ret, err := d.Swap(ctx, *functionCall.Arguments)
		if err != nil {
			return nil, err
		}
		return &model.FuncCallingRet{
			FCType:       model.FCSwap,
			FCSwapResult: *ret,
		}, nil
	case "daily_new_tokens":
		ret, err := d.DailyNewTokens(ctx, *functionCall.Arguments)
		if err != nil {
			return nil, err
		}
		return &model.FuncCallingRet{
			FCType:         model.FCDailyNewToken,
			NewTokenResult: *ret,
		}, nil
	case "token_launched_time_distribution":
		ret, err := d.TokenLaunchedTimeDistribution(ctx, *functionCall.Arguments)
		if err != nil {
			return nil, err
		}
		return &model.FuncCallingRet{
			FCType:   model.FCTokenLaunchedTimeDistribution,
			TokenLTD: *ret,
		}, nil
	case "daily_token_swap_counts":
		ret, err := d.DailyTokenSwapCounts(ctx, *functionCall.Arguments)
		if err != nil {
			return nil, err
		}
		return &model.FuncCallingRet{
			FCType:         model.FCDailyTokenSwapCount,
			TokenSwapCount: *ret,
		}, nil
	case "top_traders":
		ret, err := d.TopTraders(ctx, *functionCall.Arguments)
		if err != nil {
			return nil, err
		}
		return &model.FuncCallingRet{
			FCType:    model.FCTopTrader,
			TopTrader: *ret,
		}, nil
	case "trader_overview":
		ret, err := d.TraderOverview(ctx, *functionCall.Arguments)
		if err != nil {
			return nil, err
		}
		return &model.FuncCallingRet{
			FCType:         model.FCTraderOverview,
			TraderOverview: *ret,
		}, nil
	default:
		return nil, fmt.Errorf("unknown function: %s", *functionCall.Name)
	}
}

// swap
func (d *ChatgptDriver) Swap(ctx context.Context, params string) (*model.SwapFuncCallingResult, error) {
	swapParams := &entity.SwapParams{}
	err := json.Unmarshal([]byte(params), swapParams)
	if err != nil {
		return nil, err
	}
	var ret model.SwapFuncCallingResult
	ret.DEX = swapParams.DEX
	ret.AmountIn = swapParams.AmountIn
	ret.SwapInToken = swapParams.SwapInToken
	ret.SwapOutToken = swapParams.SwapOutToken
	ret.SwapOut = swapParams.SwapOut
	ret.SourceChain = swapParams.SourceChain
	// dest is not specified, use source chain as dest chain
	if len(swapParams.DestChain) == 0 {
		ret.DestChain = swapParams.SourceChain
	} else {
		ret.DestChain = swapParams.DestChain
	}
	return &ret, nil
}

// pump.fun 每日新创建代币数量以及上raydium的数量
func (d *ChatgptDriver) DailyNewTokens(ctx context.Context, params string) (*model.DailyNewTokensFuncCallingResult, error) {
	newTokenParams := &entity.DailyNewTokensParams{}
	err := json.Unmarshal([]byte(params), newTokenParams)
	if err != nil {
		return nil, err
	}
	reqParams := &dpmodel.CommonPumpDataQuery{
		Duration: newTokenParams.Duration,
		Timezone: newTokenParams.Timezone,
	}
	if reqParams.Timezone == "" {
		reqParams.Timezone = "CST"
	}
	resp, err := d.pumpDataService.NewTokens(ctx, reqParams)
	if err != nil {
		d.pumpDataService.BaseComponent.Logger.Error("failed to get new tokens", err)
		return nil, err
	}
	ret := &model.DailyNewTokensFuncCallingResult{
		DailyNewToken: resp,
	}
	return ret, nil
}

// pump.fun 每日新创建代币的时间分布，按半小时划分
func (d *ChatgptDriver) TokenLaunchedTimeDistribution(ctx context.Context, params string) (*model.TokenLaunchedTimeDtFuncCallingResult, error) {
	param := &entity.TokenLaunchedTimeDtParams{}
	err := json.Unmarshal([]byte(params), param)
	if err != nil {
		return nil, err
	}
	reqParams := &dpmodel.CommonPumpDataQuery{
		Duration: param.Duration,
		Timezone: param.Timezone,
	}
	if reqParams.Timezone == "" {
		reqParams.Timezone = "CST"
	}
	resp, err := d.pumpDataService.LaunchTime(ctx, reqParams)
	if err != nil {
		d.pumpDataService.BaseComponent.Logger.Error("failed to get new tokens", err)
		return nil, err
	}
	ret := &model.TokenLaunchedTimeDtFuncCallingResult{
		LaunchTimeDt: resp,
	}
	return ret, nil
}

// pump.fun 每日token交换（swap）交易统计
func (d *ChatgptDriver) DailyTokenSwapCounts(ctx context.Context, params string) (*model.DailyTokenSwapCountsFuncCallingResult, error) {
	param := &entity.DailyTokenSwapCountsParams{}
	err := json.Unmarshal([]byte(params), param)
	if err != nil {
		return nil, err
	}
	reqParams := &dpmodel.CommonPumpDataQuery{
		Duration: param.Duration,
		Timezone: param.Timezone,
	}
	if reqParams.Timezone == "" {
		reqParams.Timezone = "CST"
	}
	resp, err := d.pumpDataService.Transactions(ctx, reqParams)
	if err != nil {
		d.pumpDataService.BaseComponent.Logger.Error("failed to get daily token swap counts", err)
		return nil, err
	}
	ret := &model.DailyTokenSwapCountsFuncCallingResult{
		TxCounts: resp,
	}
	return ret, nil
}

// pump.fun top trader
func (d *ChatgptDriver) TopTraders(ctx context.Context, params string) (*model.TopTradersFuncCallingResult, error) {
	param := &entity.TopTraderParams{}
	err := json.Unmarshal([]byte(params), param)
	if err != nil {
		return nil, err
	}
	reqParams := &dpmodel.CommonPumpDataQuery{
		Duration:   param.Duration,
		Timezone:   param.Timezone,
		MaxWinRate: float64(param.WinRatio),
	}
	if reqParams.Timezone == "" {
		reqParams.Timezone = "CST"
	}
	resp, err := d.pumpDataService.TopTraders(ctx, reqParams)
	if err != nil {
		d.pumpDataService.BaseComponent.Logger.Error("failed to get top traders,", err)
		return nil, err
	}
	ret := &model.TopTradersFuncCallingResult{
		TopTraders: resp,
	}
	return ret, nil
}

// pump.fun trader 概览信息
func (d *ChatgptDriver) TraderOverview(ctx context.Context, params string) (*model.TraderOverviewFuncCallingResult, error) {
	param := &entity.TraderOverviewParams{}
	err := json.Unmarshal([]byte(params), param)
	if err != nil {
		return nil, err
	}
	reqParams := &dpmodel.CommonPumpDataQuery{
		Address: param.TraderAddr,
	}
	resp, err := d.pumpDataService.TraderDetail(ctx, reqParams)
	if err != nil {
		d.pumpDataService.BaseComponent.Logger.Error("failed to get trader overview info,", err)
		return nil, err
	}
	ret := &model.TraderOverviewFuncCallingResult{
		TraderDetails: resp,
	}
	return ret, nil
}

type Function struct {
	Name string

	Description string

	Parameters map[string]any
}

func (d *ChatgptDriver) ChatCompletionsFunctionCall(ctx context.Context, userMsg string, functions []Function) (string, error) {
	resp, err := d.client.GetChatCompletions(ctx, azopenai.ChatCompletionsOptions{
		DeploymentName: &d.cfg.Model,
		Messages: []azopenai.ChatRequestMessageClassification{
			&azopenai.ChatRequestUserMessage{
				Content: azopenai.NewChatRequestUserMessageContent(userMsg),
			},
		},
		FunctionCall: &azopenai.ChatCompletionsOptionsFunctionCall{
			Value: to.Ptr("auto"),
		},
		Functions: lo.Map(functions, func(item Function, index int) azopenai.FunctionDefinition {
			return azopenai.FunctionDefinition{
				Name:        &item.Name,
				Description: &item.Description,
				Parameters:  item.Parameters,
			}
		}),
		Temperature: to.Ptr[float32](0.0),
	}, nil)
	if err != nil {
		return "", err
	}

	replyContent := ""
	if resp.Choices[0].Message.Content != nil {
		replyContent = *resp.Choices[0].Message.Content
	}
	funcCall := resp.Choices[0].Message.FunctionCall
	if funcCall == nil {
		if replyContent != "" {
			return replyContent, nil
		}
		return "", errors.New("not a function call response")
	}
	if funcCall.Name == nil {
		return "", fmt.Errorf("parse function name is missing: %v", replyContent)
	}
	if funcCall.Arguments == nil {
		return fmt.Sprintf("your intent is parsed, call-function: %s, no arguments", *funcCall.Name), nil
	}

	return fmt.Sprintf("your intent is parsed, call-function: %s, arguments: %s", *funcCall.Name, *funcCall.Arguments), nil
}

// mapRemoteFunctionToFuncCallingRet maps the remote function call result to FuncCallingRet
func mapRemoteFunctionToFuncCallingRet(name string, result map[string]any) (*model.FuncCallingRet, error) {
	ret := &model.FuncCallingRet{}
	switch name {
	case "uniswap":
		ret.FCUniswapResult = mapToStruct[model.UniswapFuncCallingResult](result)
		ret.FCType = model.FCUniswap
	case "swap":
		ret.FCSwapResult = mapToStruct[model.SwapFuncCallingResult](result)
		ret.FCType = model.FCSwap
	case "daily_new_tokens":
		ret.NewTokenResult = mapToStruct[model.DailyNewTokensFuncCallingResult](result)
		ret.FCType = model.FCDailyNewToken
	case "token_launched_time_distribution":
		ret.TokenLTD = mapToStruct[model.TokenLaunchedTimeDtFuncCallingResult](result)
		ret.FCType = model.FCTokenLaunchedTimeDistribution
	case "daily_token_swap_counts":
		ret.TokenSwapCount = mapToStruct[model.DailyTokenSwapCountsFuncCallingResult](result)
		ret.FCType = model.FCDailyTokenSwapCount
	case "top_traders":
		ret.TopTrader = mapToStruct[model.TopTradersFuncCallingResult](result)
		ret.FCType = model.FCTopTrader
	case "trader_overview":
		ret.TraderOverview = mapToStruct[model.TraderOverviewFuncCallingResult](result)
		ret.FCType = model.FCTraderOverview
	default:
		ret.RemoteFunctionResult = result
	}
	return ret, nil
}

// / mapToStruct converts a map[string]any to a struct
func mapToStruct[T any](data map[string]any) T {
	bytes, _ := json.Marshal(data)
	var ret T
	_ = json.Unmarshal(bytes, &ret)
	return ret
}

func SupportedFunctions() []azopenai.FunctionDefinition {
	return []azopenai.FunctionDefinition{
		{
			Name:        to.Ptr("swap"),
			Description: to.Ptr("Use one crypto token swap for another token"),
			Parameters: map[string]any{
				"required": []string{},
				"type":     "object",
				"properties": map[string]any{
					"swap_in_token": map[string]any{
						"type":        "string",
						"description": "Swap input token symbol.",
					},
					"source_chain": map[string]any{
						"type":        "string",
						"description": "The chain where the input token is located. Same as the destination chain if no chain specified for source_chain (should not be null in this case!). If user specified source_chain, use it instead!",
					},
					"amount_in": map[string]any{
						"type":        "number",
						"description": "Swap input token amount.",
					},
					"dest_chain": map[string]any{
						"type":        "string",
						"description": "The chain where the output token is located. Same as the source chain if no chain specified for dest_chain (should not be null in this case!). If user specified dest_chain, use it instead!",
					},
					"swap_out_token": map[string]any{
						"type":        "string",
						"description": "Swap output token symbol.",
					},
					"dex": map[string]any{
						"type":        "string",
						"description": "Swap on which DEX.",
					},
				},
			},
		},
		// pump.fun 每日新创建代币数量以及上raydium的数量
		{
			Name:        to.Ptr("daily_new_tokens"),
			Description: to.Ptr("For several consecutive days, the number of new tokens created daily by pump.fun and the number of tokens listed on Reydium."),
			Parameters: map[string]any{
				"required": []string{},
				"type":     "object",
				"properties": map[string]any{
					"duration": map[string]any{
						"type":        "number",
						"description": "The number of consecutive days for which you want to view data.",
					},
					"timezone": map[string]any{
						"type":        "string",
						"description": "time zone.",
					},
				},
			},
		},
		// pump.fun 每日新创建代币的时间分布，按半小时划分
		{
			Name:        to.Ptr("token_launched_time_distribution"),
			Description: to.Ptr("Time distribution of dump.fun new Token creation (by half hour)."),
			Parameters: map[string]any{
				"required": []string{},
				"type":     "object",
				"properties": map[string]any{
					"duration": map[string]any{
						"type":        "number",
						"description": "The number of consecutive days for which you want to view data.",
					},
					"timezone": map[string]any{
						"type":        "string",
						"description": "time zone.",
					},
				},
			},
		},
		// pump.fun 每日token交换（swap）交易统计
		{
			Name:        to.Ptr("daily_token_swap_counts"),
			Description: to.Ptr("Daily statistics of pump.fun's token exchange (swap) transactions."),
			Parameters: map[string]any{
				"required": []string{},
				"type":     "object",
				"properties": map[string]any{
					"duration": map[string]any{
						"type":        "number",
						"description": "The number of consecutive days for which you want to view data.",
					},
					"timezone": map[string]any{
						"type":        "string",
						"description": "time zone.",
					},
				},
			},
		},
		// pump.fun top trader
		{
			Name:        to.Ptr("top_traders"),
			Description: to.Ptr("List of top (high net profit) traders on pump.fun."),
			Parameters: map[string]any{
				"required": []string{},
				"type":     "object",
				"properties": map[string]any{
					"duration": map[string]any{
						"type":        "number",
						"description": "The number of consecutive days for which you want to view data.",
					},
					"winRatio": map[string]any{
						"type":        "number",
						"description": "Trader's win rate.",
					},
					"timezone": map[string]any{
						"type":        "string",
						"description": "time zone.",
					},
				},
			},
		},
		// pump.fun trader 概览信息
		{
			Name:        to.Ptr("trader_overview"),
			Description: to.Ptr("Overview information for pump.fun traders."),
			Parameters: map[string]any{
				"required": []string{},
				"type":     "object",
				"properties": map[string]any{
					"address": map[string]any{
						"type":        "string",
						"description": "pump.fun trader’s address.",
					},
				},
			},
		},
	}
}
