package routespublish

import "time"

func (p *PayRoutesPublish) loop() {

	readUpdateFileTick := time.NewTicker(time.Minute * 15)
	//readUpdateFileTick := time.NewTicker(time.Second * 15)

	for {
		select {
		case <-readUpdateFileTick.C:
			// 读取更新文件
			p.DoUpdateByReadLogFile()

		}
	}
}
