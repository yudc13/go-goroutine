package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type cmdParams struct {
	logPath    string
	routineNum int
}

type digData struct {
	time      string
	url       string
	refer     string
	userAgent string
}

type urlData struct {
	data  digData
	uid   string
	unode UrlContent
}

type UrlContent struct {
	nuType string // 哪个页面
	nuRid  int    // 页面的id
	unUrl  string // 页面的url
	unTime string // 访问页面的时间
}

type storage struct {
	counterType  string
	storageModel string
	content      UrlContent
}

func formatLog(log string) digData {
	start := strings.Index(log, "/dig?")
	end := strings.Index(log, "HTTP/")
	if start == -1 || end == -1 {
		return digData{}
	}
	target := log[start:end]
	if len(target) == 0 {
		return digData{}
	}
	urlData, err := url.Parse("http://localhost:8000" + target)
	if err != nil {
		return digData{}
	}
	data := urlData.Query()
	return digData{
		time:      data.Get("time"),
		url:       data.Get("url"),
		refer:     data.Get("refer"),
		userAgent: data.Get("userAgent"),
	}
}

func formatUrlData(url, time string) UrlContent {
	var unType string
	if strings.Contains(url, "/home") {
		unType = "home"
	} else if strings.Contains(url, "/login") {
		unType = "login"
	} else if strings.Contains(url, "/register") {
		unType = "register"
	} else if strings.Contains(url, "/cart") {
		unType = "cart"
	}
	return UrlContent{
		nuType: unType,
		nuRid:  0,
		unUrl:  url,
		unTime: time,
	}
}

func getTime(logTime, timeType string) string {
	var format string
	switch timeType {
	case "date":
		format = "2020-01-01"
	case "hour":
		format = "2020-01-01 12"
	case "min":
		format = "2020-01-01 12:30"
	}
	t, _ := time.Parse(format, time.Now().Format(format))
	return strconv.FormatInt(t.Unix(), 10)
}

// 日志解析
func readLog(cmdParams cmdParams, logChan chan string) {
	// 打开日志文件
	fd, err := os.Open(cmdParams.logPath)
	defer fd.Close()
	if err != nil {
		log.Warnf("readLog Open file paht (%s) error: %v \n", cmdParams.logPath, err)
	}
	bufferRead := bufio.NewReader(fd)
	// 计数 读取多少行
	count := 0
	for {
		content, err := bufferRead.ReadString('\n')
		if err != nil {
			// 文件读完了
			if err == io.EOF {
				time.Sleep(time.Second * 3)
				log.Infof("readLog wait")
			} else {
				log.Warnf("readLog error: %v \n", err)
			}
		}
		// 将日志内容写入chan
		logChan <- content
		log.Infof("read log: %v", content)
		count++
		if count%(cmdParams.routineNum*1000) == 0 {
			log.Infof("readLog read count: %v \n", count)
		}
	}
}

// 消费日志
func consumeLog(logChan chan string, pvChan chan urlData, uvChan chan urlData) {
	for logStr := range logChan {
		data := formatLog(logStr)
		// 随机生成uid
		hash := md5.New()
		hash.Write([]byte(data.refer + data.userAgent)) //nolint:errcheck
		uid := hex.EncodeToString(hash.Sum(nil))
		uData := urlData{
			data:  data,
			uid:   uid,
			unode: formatUrlData(data.url, data.time),
		}
		log.Infof("uData: %v \n", uData)
		pvChan <- uData
		uvChan <- uData
	}
}

// pv统计
func pvCounter(pvChan chan urlData, storageChan chan storage) {
	for data := range pvChan {
		sItem := storage{
			counterType:  "pv",
			storageModel: "ZINCREBY",
			content:      data.unode,
		}
		storageChan <- sItem
	}
}

// uv统计
func uvCounter(uvChan chan urlData, storageChan chan storage) {
	for data := range uvChan {
		sItem := storage{"uv", "ZINCRBY", data.unode}
		storageChan <- sItem
	}
}

func dataStorage(storageChan chan storage) {
	for data := range storageChan {
		fmt.Println("存储的数据是：", data)
	}
}

var log = logrus.New()

func init() {
	log.Out = os.Stdout
	log.SetLevel(logrus.DebugLevel)
}

func main() {
	// 获取参数
	// 获取需要处理的日志路径
	logPath := flag.String("logPath", "/usr/local/var/log/nginx/dig.log", "log path")
	targetLog := flag.String("targetLog", "/usr/local/var/log/nginx/target.log", "target log path")
	routineNum := flag.Int("routineNum", 6, "goroutine number")
	flag.Parse()
	params := cmdParams{
		logPath:    *logPath,
		routineNum: *routineNum,
	}

	// 打印日志
	logFd, err := os.OpenFile(*targetLog, os.O_CREATE|os.O_WRONLY, 0644)
	defer logFd.Close()
	if err != nil {
		log.Errorf("OpenFile error: %v \n", err)
	}
	log.Out = logFd
	log.Infoln("start...")
	log.Infof("params: %+v", params)

	// 声明需要的chan
	// 日志chan
	logChan := make(chan string, *routineNum)
	// pv chan
	pvChan := make(chan urlData, *routineNum)
	// uv chan
	uvChan := make(chan urlData, *routineNum)
	// 存储 chan
	storageChan := make(chan storage, *routineNum)

	// 日志分析
	go readLog(params, logChan)

	// 消费日志的goroutine
	for i := 0; i < params.routineNum; i++ {
		go consumeLog(logChan, pvChan, uvChan)
	}

	// pv uv 统计的goroutine
	go pvCounter(pvChan, storageChan)
	go uvCounter(uvChan, storageChan)

	// 存储的goroutine
	go dataStorage(storageChan)

	time.Sleep(time.Second * 1000)
}
