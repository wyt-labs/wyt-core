package datasource

import (
	"github.com/wyt-labs/wyt-core/pkg/basic"
)

func init() {
	basic.RegisterComponents(NewMarket, NewMetrics, NewSocial)
}
