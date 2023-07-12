package convert

import "github.com/ethereum/go-ethereum/common"

func ShortAddress(address string) string {
	l := len(address)
	if l > 10 {
		address = common.HexToAddress(address).Hex()
		return address[0:6] + "..." + address[l-4:l]
	}
	return address
}
