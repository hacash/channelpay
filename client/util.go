package client

var ROUTE_API_URL = "https://hcpu.hacash.org"

func GetLoginResolutionApiDomain() string {
	//apiurl := "http://54.219.80.127:8077" // test
	if DevDebug {
		return "http://127.0.0.1:3355" // 测试
	}
	return ROUTE_API_URL
}

func SetLoginResolutionApiDomain(url string) {
	ROUTE_API_URL = url
}
