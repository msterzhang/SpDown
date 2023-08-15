// Package downloader /*
package downloader

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/msterzhang/SpDown/utils"
	"github.com/msterzhang/gpool"
	"github.com/schollz/progressbar/v3"
)

//HttpGet http请求函数
func (dl *Downloader) HttpGet(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req = dl.AddHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

//ChunkM3u8 检测是否嵌套了m3u8文件
func (dl *Downloader) ChunkM3u8(body string) error {
	if strings.Contains(body, ".m3u8") {
		return errors.New("该文件嵌套有m3u8链接，请手动处理!")
	}
	return nil
}

//GetM3u8Bash 当m3u8中没有完整链接时候需要获取m3u8文件前缀来拼接ts下载链接
func (dl *Downloader) GetM3u8Bash() string {
	u, err := url.Parse(dl.Url)
	if err != nil {
		utils.Err(err.Error())
		return ""
	}
	host := u.Scheme + "://" + u.Host + strings.Replace(filepath.Dir(u.EscapedPath()), "\\", "/", -1)
	return host
}

// GetTsList 解析ts下载链接
func (dl *Downloader) GetTsList(body string) []TsInfo {
	bash := dl.GetM3u8Bash()
	lines := strings.Split(body, "\n")
	index := 0
	tsList := []TsInfo{}
	for _, line := range lines {
		if !strings.HasPrefix(line, "#") && line != "" {
			index++
			ts := TsInfo{}
			if strings.HasPrefix(line, "http") {
				ts = TsInfo{
					Name: fmt.Sprintf("%06d.ts", index),
					Url:  line,
				}
			} else {
				ts = TsInfo{
					Name: fmt.Sprintf("%06d.ts", index),
					Url:  fmt.Sprintf("%s/%s", bash, line),
				}
			}
			tsList = append(tsList, ts)
		}
	}
	return tsList
}

// InitCachePath 新建用于缓存ts的文件夹
func (dl *Downloader) InitCachePath(id string) error {
	WorkPath, _ := os.Getwd()
	err := os.MkdirAll(WorkPath+"/cache/"+id, os.ModePerm)
	if err != nil {
		return err
	}
	os.Remove(dl.Filename)
	return nil
}

// GetM3u8Key 获取m3u8加密的密匙
func (dl *Downloader) GetM3u8Key(host, body string) (key string) {
	lines := strings.Split(body, "\n")
	key = ""
	var err error
	for _, line := range lines {
		if strings.Contains(line, "#EXT-X-KEY") {
			uriPos := strings.Index(line, "URI")
			quotationMarkPos := strings.LastIndex(line, "\"")
			keyUrl := strings.Split(line[uriPos:quotationMarkPos], "\"")[1]
			if !strings.Contains(line, "http") {
				keyUrl = fmt.Sprintf("%s/%s", host, keyUrl)
			}
			key, err = dl.HttpGet(keyUrl)
			if err != nil {
				utils.Err(err.Error())
			}
		}
	}
	return key
}

// DownloadTs 使用多线程下载ts文件
func (dl *Downloader) DownloadTs(url string, filepath string, key string, bar *progressbar.ProgressBar, pool *gpool.Pool) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		pool.Done()
		utils.Err(err.Error())
		return
	}
	req = dl.AddHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		pool.Done()
		utils.Err(err.Error())
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		pool.Done()
		utils.Err(err.Error())
		return
	}
	if key != "" {
		body, err = utils.AesDecrypt(body, []byte(key))
		if err != nil {
			pool.Done()
			utils.Err(err.Error())
			return
		}
	}
	// https://en.wikipedia.org/wiki/MPEG_transport_stream
	// Some TS files do not start with SyncByte 0x47, they can not be played after merging,
	// Need to remove the bytes before the SyncByte 0x47(71).
	syncByte := uint8(71) //0x47
	bLen := len(body)
	for j := 0; j < bLen; j++ {
		if body[j] == syncByte {
			body = body[j:]
			break
		}
	}
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		pool.Done()
		utils.Err(err.Error())
		return
	}
	defer file.Close()
	_, err = file.Write(body)
	if err != nil {
		pool.Done()
		utils.Err(err.Error())
		return
	}
	bar.Add(1)
	pool.Done()
}

// MergeFile 合并所有ts文件
func (dl *Downloader) MergeFile(tsPath string) {
	outFile, err := os.OpenFile(dl.Filename,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		utils.Err(err.Error())
		return
	}
	defer outFile.Close()
	tsFileList, _ := ioutil.ReadDir(tsPath)
	for _, f := range tsFileList {
		tsFilePath := tsPath + "/" + f.Name()
		tsFileContent, err := ioutil.ReadFile(tsFilePath)
		if err != nil {
			utils.Err(err.Error())
		}
		if _, err := outFile.Write(tsFileContent); err != nil {
			utils.Err(err.Error())
		}
		if err = os.Remove(tsFilePath); err != nil {
			utils.Err(err.Error())
		}
	}
	//删除ts缓存目录
	err = os.RemoveAll(tsPath)
	if err != nil {
		utils.Err(err.Error())
	}
	//删除缓存目录
	WorkPath, _ := os.Getwd()
	err = os.RemoveAll(WorkPath + "/cache")
	if err != nil {
		utils.Err(err.Error())
	}
}

// DownloadM3u8 m3u8下载主入口
func (dl *Downloader) DownloadM3u8(startTime time.Time) error {
	body, err := dl.HttpGet(dl.Url)
	if err != nil {
		utils.Err(err.Error())
		return err
	}
	if !strings.Contains(body, "#") {
		utils.Err("下载链接存在未知错误!")
		utils.Err(body)
		return err
	}
	err = dl.ChunkM3u8(body)
	if err != nil {
		utils.Err(err.Error())
		return err
	}
	bash := dl.GetM3u8Bash()
	key := dl.GetM3u8Key(bash, body)
	tsList := dl.GetTsList(body)
	id := uuid.New().String()
	err = dl.InitCachePath(id)
	if err != nil {
		utils.Err(err.Error())
		return err
	}
	//防止创建的线程数比任务总量多
	thSize := dl.Threads
	if len(tsList) < dl.Threads {
		thSize = len(tsList)
	}
	pool := gpool.New(thSize)
	WorkPath, _ := os.Getwd()
	tsPath := WorkPath + "/cache/" + id
	bar := progressbar.Default(int64(len(tsList)))
	for _, v := range tsList {
		FilePath := tsPath + "/" + v.Name
		pool.Add(1)
		go dl.DownloadTs(v.Url, FilePath, key, bar, pool)
	}
	pool.Wait()
	dl.MergeFile(tsPath)
	utils.Info("下载完成已保存至： %s 用时： %s", dl.Filename, time.Now().Sub(startTime))
	return nil
}
