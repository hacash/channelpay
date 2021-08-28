package routespublish

import (
	"fmt"
	"io/ioutil"
	"path"
)

/**
 * 读取最新的更新
 */

func (p *PayRoutesPublish) ReadUpdateLogFile(pagenum uint32) ([]byte, error) {
	upfilen := path.Join(p.config.DataSourceDir, fmt.Sprintf("update%d.json", pagenum))
	return ioutil.ReadFile(upfilen)
}

// 读取日志文件并更新
// 返回最后更新成功的页数
func (p *PayRoutesPublish) DoUpdateByReadLogFile() uint32 {

	// 锁定
	p.routingManager.UpdateLock()
	defer p.routingManager.UpdateUnlock()

	lastestPageNum := p.routingManager.GetUpdateLastestPageNum()
	curnum := lastestPageNum + 1

	for {
		// 读取文件
		upfbts, e := p.ReadUpdateLogFile(curnum)
		if e != nil {
			curnum--
			break // 不存在
		}
		// 更新
		fmt.Println("readUpdateLogFile:", curnum)
		p.routingManager.ForceUpdataNodesAndRelationshipByJsonBytesUnsafe(upfbts, curnum)
		curnum++ // 下一页
	}
	if curnum > lastestPageNum {
		lastestPageNum = curnum
	} else {
		// 没有更新
		fmt.Println("not find any new update log file.")
		return lastestPageNum
	}

	// 写入磁盘
	fmt.Println("flushAllNodesAndRelationshipToDisk.")
	e := p.routingManager.FlushAllNodesAndRelationshipToDiskUnsafe(
		p.config.DataSourceDir,
		&p.dataAllNodes, // 拷贝数据
		&p.dataAllGraph, // 拷贝数据
	)
	if e != nil {
		fmt.Println(e.Error())
	}

	// 返回最后有效的页数
	return lastestPageNum
}
