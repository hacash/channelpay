package client

func GetLoginResolutionApiDomain() string {
	apiurl := "https://hcpu.hacash.org"
	if DevDebug {
		apiurl = "http://127.0.0.1:3355" // 测试
	}
	return apiurl
}
