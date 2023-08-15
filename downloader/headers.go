// Package downloader /*
package downloader

import (
	"github.com/msterzhang/SpDown/config"
	"net/http"
)

//AddHeaders 添加请求头
func (dl *Downloader) AddHeaders(req *http.Request) *http.Request {
	req.Header.Set("User-Agent", dl.UA)
	if len(dl.UA) == 0 {
		req.Header.Set("User-Agent", config.UA)
	}
	if len(dl.CK) > 0 {
		req.Header.Set("Cookie", dl.CK)
	}
	if len(dl.OG) > 0 {
		req.Header.Set("Origin", dl.OG)
	}
	if len(dl.RF) > 0 {
		req.Header.Set("Referer", dl.RF)
	}
	return req
}
