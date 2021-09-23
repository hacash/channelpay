package accui

import (
	"io/ioutil"
	"path"
	"strings"
)

/**
 * 拼装UI文件
 */

func AssembleUIfiles(fpath string) {

	// 读取文件
	p1 := path.Join(fpath, "pay.html")
	p2 := path.Join(fpath, "pay.css")
	p3 := path.Join(fpath, "pay.js")
	con1, _ := ioutil.ReadFile(p1)
	con2, _ := ioutil.ReadFile(p2)
	con3, _ := ioutil.ReadFile(p3)

	htmlcon := strings.Replace(string(con1), "/*=*csscon*=*/", string(con2), 1)
	htmlcon = strings.Replace(htmlcon, "/*=*jscon*=*/", string(con3), 1)

	// 写入文件
	bdir := path.Dir(fpath)
	rfp := path.Join(bdir, "accui.go")

	htmlcon = "package client\nconst AccUIhtmlContent = `\n" + htmlcon + "\n`;"

	ioutil.WriteFile(rfp, []byte(htmlcon), 0777)

}
