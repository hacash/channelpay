package servicer

import "time"

// 启动事件循环
func (s *Servicer) loop() {

	// 8 小时更新一次路由更改
	loadUpdateFileTicker := time.NewTicker(time.Hour * 8)
	checkCustomerActiveTicker := time.NewTicker(time.Second * 35)

	for {
		select {
		case <-loadUpdateFileTicker.C:
			// 自动更新路由
			s.LoadRoutesUpdate()
		case <-checkCustomerActiveTicker.C:
			// 检查客户端心跳
			s.checkCustomerActive()
		}
	}
}
