package accui

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"
)

/**
 * 拼装UI文件
 */

func AssembleUIfiles(fpath string) {

	// read file
	p1 := path.Join(fpath, "pay.html")
	p2 := path.Join(fpath, "pay.css")
	p3 := path.Join(fpath, "pay.js")
	con1, _ := ioutil.ReadFile(p1)
	con2, _ := ioutil.ReadFile(p2)
	con3, _ := ioutil.ReadFile(p3)

	htmlcon := strings.Replace(string(con1), "/*=*csscon*=*/", string(con2), 1)
	htmlcon = strings.Replace(htmlcon, "/*=*jscon*=*/", string(con3), 1)

	// write file
	bdir := path.Dir(fpath)
	rfp := path.Join(bdir, "accui.go")

	htmlcon = "package client\nconst AccUIhtmlContent = `\n" + htmlcon + "\n`;"

	e := ioutil.WriteFile(rfp, []byte(htmlcon), 0777)
	if e != nil {
		fmt.Println(e)
	}
}
