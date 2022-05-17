/*
 * @Time    : 2022年04月21日 18:23:00
 * @Author  : root
 * @Project : SpDown
 * @File    : main.go
 * @Software: GoLand
 * @Describe:
 */
package main

import (
	"SpDown/downloader"
	"SpDown/utils"
	"errors"
	"github.com/urfave/cli"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func server() {
	app := &cli.App{
		Name:  "SpDown",
		Usage: "SpDown是一个基于golang的多线程下载引擎!",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "url, u",
				Usage: "下载链接。",
			},
			&cli.StringFlag{
				Name:  "file, f",
				Usage: "文件名,可以加全路径。",
			},
			&cli.StringFlag{
				Name:  "thread, t",
				Usage: "指定下载线程数,默认8个线程,最大64个。",
			},
			&cli.StringFlag{
				Name:  "cookie, ck",
				Usage: "指定Cookie。",
			},
			&cli.StringFlag{
				Name:  "referer, rf",
				Usage: "指定Referer。",
			},
			&cli.StringFlag{
				Name:  "origin, og",
				Usage: "指定Origin。",
			},
			&cli.StringFlag{
				Name:  "user-agent, ua",
				Usage: "指定User-Agent。",
			},
		},
		Action: func(c *cli.Context) error {
			url := c.String("u")
			file := c.String("f")
			thread := c.Int("t")
			ck := c.String("ck")
			rf := c.String("rf")
			ua := c.String("ua")
			og := c.String("og")
			if len(url) == 0 && len(file) == 0 {
				utils.Err("下载参数不足，无法下载!")
				return errors.New("下载参数不足，无法下载")
			}
			if thread == 0 {
				thread = 8
			}
			if thread > 100 {
				thread = 64
			}
			if len(rf) == 0 {
				rf = url
			}
			client := &http.Client{}
			download := downloader.NewDownloader(client, file, url, ua, og, rf, ck, thread)

			startTime := time.Now()
			if strings.Contains(download.Url, ".m3u8") {
				return download.DownloadM3u8(startTime)
			}
			if download.Threads <= 1 {
				return download.SingleThreaded(startTime)
			}
			resp, err := download.Client.Head(url)
			if err != nil {
				utils.Info("资源错误!")
				return err
			}
			contentLength := resp.Header.Get("Content-Length")
			ranges := resp.Header.Get("Accept-Ranges")
			if contentLength == "" {
				utils.Info("资源不支持多线程下载,使用单线程下载!")
				return download.SingleThreaded(startTime)
			}
			if ranges != "bytes" {
				utils.Info("服务器不支持多线程下载,使用单线程下载!")
				return download.SingleThreaded(startTime)
			}
			length, err := strconv.Atoi(contentLength)
			if err != nil {
				return err
			}
			err = download.ThreadedDownload(length, startTime)
			if err != nil {
				return err
			}
			return nil
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		os.Exit(1)
	}
}

func main() {
	server()
}
