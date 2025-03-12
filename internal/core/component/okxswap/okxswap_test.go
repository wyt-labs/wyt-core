package okxswap

import (
	"testing"

	"github.com/wyt-labs/wyt-core/internal/pkg/base"
)

func GetOkxSwapApi(t *testing.T) *OkxSwapApi {
	baseComponent := base.NewMockBaseComponent(t)
	oapi, err := NewOkxSwapApi(baseComponent)
	if err != nil {
		t.Fatal(err)
	}
	return oapi
}

func TestOkxSwapApi_GetSupportedChains(t *testing.T) {
	oapi := GetOkxSwapApi(t)
	chains, err := oapi.GetSupportedChains(0, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("same chain swap, supported chains:", chains)
	chains, err = oapi.GetSupportedChains(0, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("cross-chain swap, supported chains:", chains)
}

func TestOkxSwapApi_UIAmount2ContractAmount(t *testing.T) {
	f, err := UIAmount2ContractAmount("5.11231", "6")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(f)
}

func TestOkxSwapApi_ContractAmount2UIAmount(t *testing.T) {
	f, err := ContractAmount2UIAmount("618025728476", "8")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(f)
}

func TestOkxSwapApi_GetTokens(t *testing.T) {
	oapi := GetOkxSwapApi(t)
	chains, err := oapi.GetTokens(1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("all tokens:", chains)
}

func TestOkxSwapApi_GetCrossChainTokens(t *testing.T) {
	oapi := GetOkxSwapApi(t)
	chains, err := oapi.GetCrossChainTokens(1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("list of tokens supported by the cross-chain bridges:", chains)
}

func TestOkxSwapApi_GetBridgeTokensPairs(t *testing.T) {
	oapi := GetOkxSwapApi(t)
	pairs, err := oapi.GetBridgeTokensPairs(1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("list of token pairs supported by from chain:", pairs)
}

func TestOkxSwapApi_GetCrossChainQuote(t *testing.T) {
	oapi := GetOkxSwapApi(t)
	quotes, err := oapi.GetCrossChainQuote(324, 42161, "0x3355df6d4c9c3035724fd0e3914de96a5a83aaf4", "0xff970a61a04b1ca14834a43f5de4533ebddb5cc8", "788390", "0.1")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("cross-chain quote:", quotes)
}

func TestOkxSwapApi_GetTokenPrice(t *testing.T) {
	oapi := GetOkxSwapApi(t)
	quotes, err := oapi.GetTokenPrice(501, "11111111111111111111111111111111")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("token prices:", quotes)
}

func TestOkxSwapApi_GetQuotes(t *testing.T) {
	oapi := GetOkxSwapApi(t)
	chains, err := oapi.GetQuotes(1, "10000000", "0xdac17f958d2ee523a2206206994597c13d831ec7", "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(chains)
}

func TestOkxSwapApi_ApproveTransaction(t *testing.T) {
	oapi := GetOkxSwapApi(t)
	chains, err := oapi.ApproveTransaction(1, "0xdac17f958d2ee523a2206206994597c13d831ec7", "1000000")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(chains)
}

func TestOkxSwapApi_Swap(t *testing.T) {
	oapi := GetOkxSwapApi(t)
	chains, err := oapi.Swap(501, "100000", "11111111111111111111111111111111", "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", "0x78E3b1A21744868BF7c102ee5d9B02341f7dCe73", "0.2", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(chains)
}

func TestOkxSwapApi_CrosschainSwap(t *testing.T) {
	oapi := GetOkxSwapApi(t)
	chains, err := oapi.CrosschainSwap(324, 42161, "505900", "0x3355df6d4c9c3035724fd0e3914de96a5a83aaf4", "0xff970a61a04b1ca14834a43f5de4533ebddb5cc8", "0x22497668Fb12BA21E6A132de7168D0Ecc69cDF7d", "0.2")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(chains)
}

func Test_GetTokenPriceV2(t *testing.T) {
	oapi := GetOkxSwapApi(t)
	chains, err := oapi.GetTokenPriceV2("eth")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(chains)
}
