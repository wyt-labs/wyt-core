package datapuller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
)

func TestMetabaseDataSource_Auth(t *testing.T) {
	baseComponent := base.NewMockBaseComponent(t)
	metabaseDataSource, err := NewMetabaseDataSource(baseComponent)
	if err != nil {
		t.Fatal(err)
	}
	token, err := metabaseDataSource.Auth(
		baseComponent.Config.Backends.MetabaseUserName,
		baseComponent.Config.Backends.MetabasePassword,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token)
}

func Test_HttpReq(t *testing.T) {
	url := "http://119.12.174.42/metabase/api/session"
	method := "POST"

	payload := strings.NewReader(`{
    	"username": "wyt@axiomesh.io",
    	"password": "VvBQSJbvig31U7"
    }`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}

func TestMetabaseDataSource_DailyLaunchedTokenInfo(t *testing.T) {
	baseComponent := base.NewMockBaseComponent(t)
	metabaseDataSource, err := NewMetabaseDataSource(baseComponent)
	if err != nil {
		t.Fatal(err)
	}
	duration := 7
	ret, err := metabaseDataSource.DailyLaunchedTokenInfo(duration, "UTC")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, ret)
	assert.Equal(t, duration, len(ret.Data.Rows))
	rows, _ := json.Marshal(ret.Data.Rows)
	t.Log(string(rows))
	cols, _ := json.Marshal(ret.Data.Cols)
	t.Log(string(cols))
}

func TestMetabaseDataSource_LaunchedTokenTimeDistribution(t *testing.T) {
	baseComponent := base.NewMockBaseComponent(t)
	metabaseDataSource, err := NewMetabaseDataSource(baseComponent)
	if err != nil {
		t.Fatal(err)
	}
	duration := 7
	ret, err := metabaseDataSource.LaunchedTokenTimeDistribution(duration, "UTC")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, ret)
	assert.Equal(t, duration, len(ret.Data.Rows))
	rows, _ := json.Marshal(ret.Data.Rows)
	t.Log(string(rows))
	cols, _ := json.Marshal(ret.Data.Cols)
	t.Log(string(cols))
}

func TestMetabaseDataSource_DailyTradeCounts(t *testing.T) {
	baseComponent := base.NewMockBaseComponent(t)
	metabaseDataSource, err := NewMetabaseDataSource(baseComponent)
	if err != nil {
		t.Fatal(err)
	}
	duration := 3
	ret, err := metabaseDataSource.DailyTradeCounts(duration, "UTC")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, ret)
	assert.Equal(t, duration, len(ret.Data.Rows))
	rows, _ := json.Marshal(ret.Data.Rows)
	t.Log(string(rows))
	cols, _ := json.Marshal(ret.Data.Cols)
	t.Log(string(cols))
}

func TestMetabaseDataSource_TraderOverview(t *testing.T) {
	baseComponent := base.NewMockBaseComponent(t)
	metabaseDataSource, err := NewMetabaseDataSource(baseComponent)
	if err != nil {
		t.Fatal(err)
	}
	ret, err := metabaseDataSource.TraderOverview("BwnYE9xp5aBKQSfBAr2VXEpcyrw5sYyaeiSLUMKFv1YZ")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, ret)
	rows, _ := json.Marshal(ret.Data.Rows)
	t.Log(string(rows))
	cols, _ := json.Marshal(ret.Data.Cols)
	t.Log(string(cols))
}

func TestMetabaseDataSource_TraderOverviewV2(t *testing.T) {
	baseComponent := base.NewMockBaseComponent(t)
	metabaseDataSource, err := NewMetabaseDataSource(baseComponent)
	if err != nil {
		t.Fatal(err)
	}
	ret, err := metabaseDataSource.TraderOverviewV2("9YqDWbEpKME1pjM91FYDMYfuTbFCqmWT8GLitfe19Ngr", "CST", 7)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, ret)
	rows, _ := json.Marshal(ret.Data.Rows)
	t.Log(string(rows))
	cols, _ := json.Marshal(ret.Data.Cols)
	t.Log(string(cols))
}

func TestMetabaseDataSource_TraderTxTimeDistribution(t *testing.T) {
	baseComponent := base.NewMockBaseComponent(t)
	metabaseDataSource, err := NewMetabaseDataSource(baseComponent)
	if err != nil {
		t.Fatal(err)
	}
	ret, err := metabaseDataSource.TraderTxTimeDistribution("74tYkMYmwnmi44PQo6L6QpkxmdNTdX5AZaiKMrMAncwW", 7, "UTC")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, ret)
	rows, _ := json.Marshal(ret.Data.Rows)
	t.Log(string(rows))
	cols, _ := json.Marshal(ret.Data.Cols)
	t.Log(string(cols))
}

func TestMetabaseDataSource_TraderProfitTokenDistribution(t *testing.T) {
	baseComponent := base.NewMockBaseComponent(t)
	metabaseDataSource, err := NewMetabaseDataSource(baseComponent)
	if err != nil {
		t.Fatal(err)
	}
	ret, err := metabaseDataSource.TraderProfitTokenDistribution("74tYkMYmwnmi44PQo6L6QpkxmdNTdX5AZaiKMrMAncwW", 7, "UTC")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, ret)
	rows, _ := json.Marshal(ret.Data.Rows)
	t.Log(string(rows))
	cols, _ := json.Marshal(ret.Data.Cols)
	t.Log(string(cols))
}

func TestMetabaseDataSource_TraderProfitDistribution(t *testing.T) {
	baseComponent := base.NewMockBaseComponent(t)
	metabaseDataSource, err := NewMetabaseDataSource(baseComponent)
	if err != nil {
		t.Fatal(err)
	}
	ret, err := metabaseDataSource.TraderProfitDistribution("GHeFBkkxCjwi35RnNZuvoQarCZ9DV6ZmZFqaS9h4noob", 7, "UTC")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, ret)
	rows, _ := json.Marshal(ret.Data.Rows)
	t.Log(string(rows))
	cols, _ := json.Marshal(ret.Data.Cols)
	t.Log(string(cols))
}

func TestMetabaseDataSource_TopTrader(t *testing.T) {
	baseComponent := base.NewMockBaseComponent(t)
	metabaseDataSource, err := NewMetabaseDataSource(baseComponent)
	if err != nil {
		t.Fatal(err)
	}
	ret, err := metabaseDataSource.TopTrader(1, 1)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, ret)
	rows, _ := json.Marshal(ret.Data.Rows)
	t.Log(string(rows))
	cols, _ := json.Marshal(ret.Data.Cols)
	t.Log(string(cols))
}

func TestMetabaseDataSource_Common(t *testing.T) {

}
