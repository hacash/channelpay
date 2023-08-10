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
	fmt.Println("[Config] Read update log file:", upfilen)
	return ioutil.ReadFile(upfilen)
}

// Read log file and update
// Returns the number of pages that were last successfully updated
func (p *PayRoutesPublish) DoUpdateByReadLogFile() uint32 {

	// locking
	p.routingManager.UpdateLock()
	defer p.routingManager.UpdateUnlock()

	lastestPageNum := p.routingManager.GetUpdateLastestPageNum()
	curnum := lastestPageNum + 1

	for {
		// read file
		upfbts, e := p.ReadUpdateLogFile(curnum)
		if e != nil {
			curnum--
			break // non-existent
		}
		// to update
		fmt.Println("readUpdateLogFile:", curnum)
		p.routingManager.ForceUpdataNodesAndRelationshipByJsonBytesUnsafe(upfbts, curnum)
		curnum++ // next page
	}
	if curnum > lastestPageNum {
		lastestPageNum = curnum
	} else {
		// No updates
		fmt.Println("not find any new update log file.")
		return lastestPageNum
	}

	// Write to disk
	fmt.Println("flushAllNodesAndRelationshipToDisk.")
	e := p.routingManager.FlushAllNodesAndRelationshipToDiskUnsafe(
		p.config.DataSourceDir,
		&p.dataAllNodes, // Copy data
		&p.dataAllGraph, // Copy data
	)
	if e != nil {
		fmt.Println(e.Error())
	}

	// Returns the last valid pages
	return lastestPageNum
}
