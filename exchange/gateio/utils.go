package gateio

import (
	"fmt"
	"strings"
)

func (g GateIO) FormatSymbol(base, quote string) string {
	return fmt.Sprintf("%s_%s", strings.ToLower(base), strings.ToLower(quote))
}
