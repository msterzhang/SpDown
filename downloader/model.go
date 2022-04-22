// Package downloader /*
package downloader

import "net/http"

// Downloader 文件下载对象主体属性
type Downloader struct {
	Client   *http.Client
	Filename string
	Url      string
	UA       string
	OG       string
	RF       string
	CK       string
	Threads  int
}

// TsInfo 用于保存 ts 文件的下载地址和文件名
type TsInfo struct {
	Name string
	Url  string
}
