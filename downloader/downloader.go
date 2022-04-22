package downloader

import (
	"SpDown/utils"
	"fmt"
	"github.com/msterzhang/gpool"
	"github.com/schollz/progressbar/v3"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

var l sync.Mutex

// SingleThreaded 使用单线程下载
func (dl *Downloader) SingleThreaded(startTime time.Time) error {
	req, err := http.NewRequest("GET", dl.Url, nil)
	if err != nil {
		return err
	}
	req = dl.AddHeaders(req)
	resp, err := dl.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"正在下载",
	)
	file, err := os.OpenFile(dl.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	io.Copy(io.MultiWriter(file, bar), resp.Body)
	utils.Info("已保存至： %s 用时： %s", dl.Filename, time.Now().Sub(startTime))
	return nil
}

// DownLoad 下载多线程中的块文件
func (dl *Downloader) DownLoad(start, end int, file *os.File, bar *progressbar.ProgressBar, pool *gpool.Pool) {
	req, err := http.NewRequest("GET", dl.Url, nil)
	if err != nil {
		pool.Done()
		utils.Err(err.Error())
	}
	req = dl.AddHeaders(req)
	byteRange := fmt.Sprintf("bytes=%d-%d", start, end-1)
	req.Header.Add("Range", byteRange)
	resp, err := dl.Client.Do(req)
	if err != nil {
		pool.Done()
		utils.Err(err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		pool.Done()
		utils.Err(err.Error())
	}
	l.Lock()
	file.WriteAt(body, int64(start))
	l.Unlock()
	pool.Done()
	bar.Add(1)
}

// ThreadedDownload 使用多线程下载
func (dl *Downloader) ThreadedDownload(length int, startTime time.Time) error {
	//将文件分成100个块，使用多线程下载
	yuan := 100
	size := length / yuan
	remainder := length % dl.Threads
	bar := progressbar.Default(int64(yuan))
	pool := gpool.New(dl.Threads)
	file, err := os.OpenFile(dl.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		utils.Err(err.Error())
	}
	for i := 0; i < yuan; i++ {
		pool.Add(1)
		start := i * size
		end := (i + 1) * size
		if i == dl.Threads-1 {
			end += remainder
		}
		go dl.DownLoad(start, end, file, bar, pool)
	}
	pool.Wait()
	utils.Info("已保存至： %s 用时： %s", dl.Filename, time.Now().Sub(startTime))
	return nil
}
