包含golang原生httpClient引擎,和phantomjs的下载引擎,你可以把它看成一个爬虫浏览器,后期打算添加Selenium
It contains golang httpClient engine, and phantomjs download engine, you can use it as a crawler browser ,the next idea to add the Selenium

```
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	. "github.com/k4s/webrowser"
)

func main() {
	p := &Param{
		Method:       "GET",
		Url:          "http://weibo.com/kasli/home?wvr=5",
		Header:       http.Header{"Cookie": []string{"your cookie"}},
		UsePhantomJS: true,
	}
	brower := NewWebrowse()
	resp, err := brower.Download(p)
	if err != nil {
		fmt.Println(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	fmt.Println(resp.Cookies())
}

```