package servicer

import "time"

// 启动事件循环
func (s *Servicer) loop() {

	// 8 小时更新一次路由更改
	loadUpdateFileTick := time.NewTicker(time.Hour * 8)
	//loadUpdateFileTick := time.NewTicker(time.Second * 8)

	for {
		select {
		case <-loadUpdateFileTick.C:
			// 自动更新路由
			s.LoadRoutesUpdate()

		}
	}
}
