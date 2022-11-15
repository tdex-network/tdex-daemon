package application

import (
	swap_parser "github.com/tdex-network/tdex-daemon/internal/core/application/swap-parser"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

func NewSwapParserService() domain.SwapParser {
	return swap_parser.NewService()
}
