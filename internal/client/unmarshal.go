package client

var (
	GlobalTokenInfo TokenInfo
)

func InitUnmarshalData(info TokenInfo) {
	GlobalTokenInfo = info
}
