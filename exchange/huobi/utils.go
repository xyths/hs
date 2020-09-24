package huobi

import (
	"fmt"
	"strings"
)

func (c Client) FormatSymbol(base, quote string) string {
	return fmt.Sprintf("%s%s", strings.ToLower(base), strings.ToLower(quote))
}
