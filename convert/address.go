package convert

func ShortAddress(address string) string {
	l := len(address)
	if l > 10 {
		return address[0:6] + "..." + address[l-4:l]
	}
	return address
}
