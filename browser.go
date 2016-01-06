package webrowser

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/k4s/phantomgo"
)

//内部对外的下载接口
type Webrowser interface {
	Download(Request) (resp *http.Response, err error)
}

//浏览器
type Webrowse struct {
	userAgent string
	cookieJar http.CookieJar
}

//new一个浏览器
func NewWebrowse() Webrowser {
	CookiesMemory, _ := cookiejar.New(nil)
	webrowse := &Webrowse{
		userAgent: "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.101 Safari/537.36",
		cookieJar: CookiesMemory,
	}
	return webrowse
}

//浏览器参数
type WebrowseParam struct {
	method        string
	url           *url.URL
	header        http.Header
	postBody      string
	redirectTimes int           //重定向次数
	dialTimeout   time.Duration //拨号超时时间段
	connTimeout   time.Duration //链接超时时间
	tryTimes      int           //请求失败重新请求次数
	retryPause    time.Duration //请求失败时重复试时间段
	client        *http.Client
}

func (self *Webrowse) Download(req Request) (resp *http.Response, err error) {

	//是否执行PhontomJS下载器
	if req.GetusePhomtomJS() {
		phantombrowse := phantomgo.NewPhantom()
		return phantombrowse.Download(req)
	} else {
		var webrowseParam = new(WebrowseParam)

		//请求方法
		webrowseParam.method = strings.ToUpper(req.GetMethod())
		//请求地址
		webrowseParam.url, err = url.ParseRequestURI(req.GetUrl())
		if err != nil {
			return nil, err
		}
		//请求http头
		webrowseParam.header = req.GetHeader()
		//postDATA
		webrowseParam.postBody = req.GetPostBody()
		//获取重定向次数
		webrowseParam.redirectTimes = req.GetRedirectTimes()
		//请求尝试次数
		webrowseParam.tryTimes = req.GetTryTimes()
		//拨号超时时间
		webrowseParam.dialTimeout = req.GetDialTimeout()
		//链接超时时间
		webrowseParam.connTimeout = req.GetConnTimeout()
		//请求失败重新尝试的间隔时间
		webrowseParam.retryPause = req.GetRetryPause()
		//请求客户端
		webrowseParam.client = &http.Client{
			CheckRedirect: webrowseParam.checkRedirect,
			Jar:           self.cookieJar,
		}
		//请求和链接超时
		transport := &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				c, err := net.DialTimeout(network, addr, webrowseParam.dialTimeout)
				if err != nil {
					return nil, err
				}
				if webrowseParam.connTimeout > 0 {
					c.SetDeadline(time.Now().Add(webrowseParam.connTimeout))
				}
				return c, nil
			},
		}
		//判断是否为ssl协议
		if strings.ToLower(webrowseParam.url.Scheme) == "https" {
			transport.TLSClientConfig = &tls.Config{RootCAs: nil, InsecureSkipVerify: true}
			transport.DisableCompression = true
		}

		webrowseParam.client.Transport = transport
		return self.httpRequest(webrowseParam)
	}
	return nil, err

}

func (self *Webrowse) httpRequest(webrowseParam *WebrowseParam) (resp *http.Response, err error) {
	req, err := http.NewRequest(webrowseParam.method, webrowseParam.url.String(), strings.NewReader(webrowseParam.postBody))
	if err != nil {
		return nil, err
	}

	//设置头信息，包括cookie
	for k, v := range webrowseParam.header {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}

	//设置user-agent
	req.Header.Set("User-Agent", self.userAgent)

	//如果是POST方法，添加Content-Type
	if webrowseParam.method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if webrowseParam.tryTimes <= 0 {
		for {
			resp, err = webrowseParam.client.Do(req)
			if err != nil {
				time.Sleep(webrowseParam.retryPause)
				continue
			}
			break
		}
	} else {
		for i := 0; i < webrowseParam.tryTimes; i++ {
			resp, err = webrowseParam.client.Do(req)
			if err != nil {
				time.Sleep(webrowseParam.retryPause)
				continue
			}
			break
		}
	}
	return resp, err

}

// checkRedirect is used as the value to http.Client.CheckRedirect
// when redirectTimes equal 0, redirect times is ∞
// when redirectTimes less than 0, not allow redirects
func (self *WebrowseParam) checkRedirect(req *http.Request, via []*http.Request) error {
	if self.redirectTimes == 0 {
		return nil
	}
	if len(via) >= self.redirectTimes {
		if self.redirectTimes < 0 {
			return fmt.Errorf("not allow redirects.")
		}
		return fmt.Errorf("stopped after %v redirects.", self.redirectTimes)
	}
	return nil
}
