## SpDown是一个基于golang的命令行下载器

* 支持设置多个请求参数Cookie，Referer，Origin，User-Agent
* 全类型文件支持多线程下载，包括m3u8文件
* 支持下载m3u8资源
* 支持下载含有加密m3u8资源

```
GLOBAL OPTIONS:
   --url value, -u value           下载链接。
   --file value, -f value          文件名,可以加全路径。
   --thread value, -t value        指定下载线程数,默认8个线程,最大64个。
   --cookie value, --ck value      指定Cookie。
   --referer value, --rf value     指定Referer。
   --origin value, --og value      指定Origin。
   --user-agent value, --ua value  指定User-Agent。
   --help, -h                      show help
```

能下载如阿里云盘（添加Referer），夸克云盘（添加Cookie），百度云盘（修改User-Agent）等