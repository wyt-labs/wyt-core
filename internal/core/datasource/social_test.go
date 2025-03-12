package datasource

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
)

func TestParseGithub(t *testing.T) {
	component := base.NewMockBaseComponent(t)
	metrics, err := NewSocial(component, nil, nil)
	require.Nil(t, err)
	githubInfo := &GhInfo{}
	// good url
	err = metrics.parseGithub("https://github.com/zenlinkpro/", githubInfo)
	require.Nil(t, err)

	// bad url
	githubInfo = &GhInfo{}
	err = metrics.parseGithub("https://github.com/baofinance1323", githubInfo)
	require.Contains(t, err.Error(), invalidUrl)
}
