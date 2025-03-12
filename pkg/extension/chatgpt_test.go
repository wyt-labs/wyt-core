package extension

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/wyt-labs/wyt-core/internal/core/component/datapuller"
	"github.com/wyt-labs/wyt-core/internal/core/model"
	"github.com/wyt-labs/wyt-core/internal/pkg/azure"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
)

func anyToString(value any) string {
	return fmt.Sprintf("%v", value)
}

func BuildChatGpt(t *testing.T) *ChatgptDriver {
	baseComponent := base.NewMockBaseComponent(t)
	// metabase
	metabaseDataSource, err := datapuller.NewMetabaseDataSource(baseComponent)
	if err != nil {
		baseComponent.Logger.Error(err)
		t.Fatal(err)
	}
	pumpDataSource := datapuller.NewPumpDataService(baseComponent, metabaseDataSource)
	cfg := baseComponent.Config.Extension.Chatgpt
	gptcfg := &ChatgptConfig{
		Endpoint:        cfg.Endpoint,
		EndpointFull:    cfg.EndpointFull,
		APIKey:          cfg.APIKey,
		Model:           cfg.Model,
		Temperature:     cfg.Temperature,
		PresencePenalty: cfg.PresencePenalty,
	}

	gpt, err := NewChatgptDriver(gptcfg, baseComponent, pumpDataSource)
	if err != nil {
		baseComponent.Logger.Error(err)
		t.Fatal(err)
	}
	return gpt

}

func TestChatgptDriver_DocFaq(t *testing.T) {
	gpt := BuildChatGpt(t)
	msgs := []*ChatgptMsg{
		{"user", "What is the difference between a public and private blockchain?"},
	}
	content, funct, err := gpt.DocsFaq(context.Background(), msgs, "")
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(content)
	assert.True(t, len(content) > 0)
	assert.Nilf(t, funct, "Should not return function")
}

func TestChatgptDriver_LocalFunction(t *testing.T) {
	gpt := BuildChatGpt(t)

	tests := []struct {
		userMessage string
	}{
		{
			userMessage: "pump.fun daily create tokens, for 7 consecutive days at UTC timezone",
		},
		{
			userMessage: "pump.fun token creation time distribution (by half hour) of 7 days, at UTC timezone?",
		},
		{
			userMessage: "pump.fun daily statistics of token exchange (swap) transactions of 7 days, at UTC timezone",
		},
		{
			userMessage: "Overview information of pump.fun's trader 8i57XsS3E4iuw2qy2cPbKDWnW4pwx6yaBc7N7UQzG3MJ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.userMessage, func(t *testing.T) {
			msgs := []*ChatgptMsg{
				{"user", tt.userMessage},
			}
			content, funct, err := gpt.DocsFaq(context.Background(), msgs, "cm1g8vjdj0000h2nu2iyscpwx")
			if err != nil {
				t.Fatal(err)
			}
			assert.NotNil(t, funct)
			assert.True(t, len(content) == 0)
			t.Logf("content: %v", content)
			bytes, _ := json.Marshal(funct)
			t.Logf("funct: %v", string(bytes))
		})
	}
}

func TestChatgptDriver_uniswap(t *testing.T) {
	gpt := BuildChatGpt(t)

	tests := []struct {
		userMessage string
	}{
		{
			"Ethereum",
		},
		{
			"project Bitcoin",
		},
		{
			"what is Uniswap?",
		},
		{
			"How to swap tokens on Uniswap?",
		},
		{
			"swap tokens on Uniswap",
		},
		{
			"swap 2 eth to doge on uniswap",
		},
		{
			"bridge 2 eth from mainnet to bsc on uniswap",
		},
		{
			"swap 0x78E3b1A21744868BF7c102ee5d9B02341f7dCe73 on uniswap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.userMessage, func(t *testing.T) {
			msgs := []*ChatgptMsg{
				{"user", tt.userMessage},
			}
			// cm1g8vjdj0000h2nu2iyscpwx uniswap
			intentContent, funcCallRet, err := gpt.DocsFaq(context.Background(), msgs, "cm1g8vjdj0000h2nu2iyscpwx")
			if err != nil {
				t.Fatal(err)
			}
			if intentContent != "" {
				t.Logf("content: %v", intentContent)
			}
			if funcCallRet != nil {
				bytes, _ := json.Marshal(funcCallRet)
				t.Logf("func ret: %v", string(bytes))
			}
		})
	}
}

func TestChatgptDriver_wyt(t *testing.T) {
	gpt := BuildChatGpt(t)

	tests := []struct {
		userMessage string
	}{
		{
			"Search ETH",
		},
		{
			"Search Ethereum",
		},
		{
			"Project ETH",
		},
		{
			"Project Ethereum",
		},
		{
			"Find ETH",
		},
		{
			"ETH info",
		},
		{
			"project Bitcoin information",
		},
		{
			"Show me the project info of Ethereum",
		},
		{
			"Search ETH Team",
		},
		{
			"Project ETH Team",
		},
		{
			"Ethereum Team",
		},
		{
			"ETH Team",
		},
		{
			"team of ETH",
		},
		{
			"Show me the team info of Ethereum",
		},
		{
			"Compare Ethereum with Solana",
		},
		{
			"swap ETH for SOL, from Ethereum to Solana",
		},
		{
			"swap ETH for SOL",
		},
		{
			"Swap",
		},
		{
			"Daily new tokens launched on Pump.fun",
		},
		{
			"How many new tokens were launched daily on Pump.fun in the last week",
		},
		{
			"Time distribution of new token launches on Pump.fun",
		},
		{
			"Number of Newly Launched Tokens Over Time on Pump.fun",
		},
		{
			"Pump.fun daily transactions",
		},
		{
			"Daily transactions on Pump.fun",
		},
		{
			"swap 0x78E3b1A21744868BF7c102ee5d9B02341f7dCe73 on uniswap",
		},
		{
			"Pump.fun daily new tokens",
		},
		{
			"Pump.fun top traders",
		},
		{
			"Pump.fun daily token exchange transactions",
		},
		{
			"Pump.fun trader 8i57XsS3E4iuw2qy2cPbKDWnW4pwx6yaBc7N7UQzG3MJ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.userMessage, func(t *testing.T) {
			msgs := []*ChatgptMsg{
				{"user", tt.userMessage},
			}
			// cm18rlpic0000sjx74wq4nzkz wyt
			// dev cm29v8y9300002olm9cr079xg
			intentContent, funcCallRet, err := gpt.DocsFaq(context.Background(), msgs, "cm29v8y9300002olm9cr079xg")
			if err != nil {
				t.Fatal(err)
			}
			if intentContent != "" {
				t.Logf("content: %v", intentContent)
			}
			if funcCallRet != nil {
				bytes, _ := json.Marshal(funcCallRet)
				t.Logf("func ret: %s", string(bytes))
			}
		})
	}
}

func TestChatgptDriver_RemoteFunction(t *testing.T) {
	gpt := BuildChatGpt(t)
	tests := []struct {
		userMessage string
	}{
		{
			userMessage: "Swap 10 ETH for BSC on WYT Swap",
		},
	}

	for _, tt := range tests {
		content, funct, err := gpt.DocsFaq(context.Background(), []*ChatgptMsg{
			{"user", tt.userMessage},
		}, "")
		if err != nil {
			t.Fatal(err)
			return
		}
		t.Log(content)
		assert.True(t, len(content) == 0)
		assert.NotNil(t, funct)
		assert.True(t, len(anyToString(funct.RemoteFunctionResult["source_chain"])) > 0)
	}
}

func TestChatgptDriver_Swap(t *testing.T) {
	type fields struct {
		cfg             *ChatgptConfig
		client          *azopenai.Client
		azureRestClient *azure.ChatAPI
	}
	type args struct {
		ctx    context.Context
		params map[string]any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.SwapFuncCallingResult
		wantErr bool
	}{
		{
			name: "Should return dest_chain set if source_chain defined",
			args: args{
				params: map[string]any{
					"source_chain":   "BSC",
					"swap_in_token":  "ETH",
					"swap_out_token": "Ethereum",
					"dest_chain":     "",
				},
			},
			want: &model.SwapFuncCallingResult{
				SourceChain:  "BSC",
				DestChain:    "BSC",
				SwapInToken:  "ETH",
				SwapOutToken: "Ethereum",
				DEX:          "WYT Swap",
			},
		},
		{
			name: "Should return source_chain set if dest_chain defined",
			args: args{
				params: map[string]any{
					"source_chain":   "",
					"swap_in_token":  "ETH",
					"swap_out_token": "Ethereum",
					"dest_chain":     "BSC",
				},
			},
			want: &model.SwapFuncCallingResult{
				SourceChain:  "BSC",
				DestChain:    "BSC",
				SwapInToken:  "ETH",
				SwapOutToken: "Ethereum",
				DEX:          "WYT Swap",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &ChatgptDriver{
				cfg:             tt.fields.cfg,
				client:          tt.fields.client,
				azureRestClient: tt.fields.azureRestClient,
			}
			jsonParams, _ := json.Marshal(tt.args.params)
			got, err := d.Swap(tt.args.ctx, string(jsonParams))
			if (err != nil) != tt.wantErr {
				t.Errorf("Swap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Swap() got = %v, want %v", got, tt.want)
			}
		})
	}
}
