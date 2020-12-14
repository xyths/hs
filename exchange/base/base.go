package base

type Authentication interface {
	Init(apiKey, secretKey string)
	Build() (string, error)
}
