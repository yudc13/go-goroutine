package utils

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

// 模拟生成日志 方便后面的分析

const BASR_URL = "http://localhost:8080"

func makeLog(timeStr string, url string, refer string) string {
	logTemplate := "127.0.0.1 - - [09/Jan/2021:21:18:17 +0800] \"GET /dig?time={time}&url={url}&refer={refer}&cookie=user%3DMTYxMDE3Nzg0OHxEdi1CQkFFQ180SUFBUkFCRUFBQVpQLUNBQUVHYzNSeWFXNW5EQVlBQkhWelpYSVlaMjlzWVc1blYyVmlRbTl2YXk5dGIyUmxiQzVWYzJWeV80TURBUUVFVlhObGNnSF9oQUFCQkFFQ1NXUUJCQUFCQ0ZWelpYSnVZVzFsQVF3QUFRTlFkMlFCREFBQkJVVnRZV2xzQVF3QUFBQWdfNFFkQVFZQkJIbDFaR01DRW5sMVpHRmphR0Z2UUdkdFlXbHNMbU52YlFBPXx9JoL9w1LbtgQo8NC-LorbBNSnqCaHY-QCOZTHe1kWQA%3D%3D&userAgent=Mozilla%2F5.0+(Macintosh%3B+Intel+Mac+OS+X+11_1_0)+AppleWebKit%2F537.36+(KHTML,+like+Gecko)+Chrome%2F87.0.4280.88+Safari%2F537.36 HTTP/1.1\" 200 43 \"{url}\" \"Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36\" \"127.0.0.1\"\n"
	log := strings.Replace(logTemplate, "{time}", timeStr, -1)
	log = strings.Replace(log, "{url}", BASR_URL+url, -1)
	log = strings.Replace(log, "{refer}", BASR_URL+refer, -1)
	return log
}

func randInt(min int, max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if min > max {
		return max
	}
	return r.Intn(max - min)
}

func GenerateLog() {
	// 日志路径
	logPath := "/usr/local/var/log/nginx/dig.log"

	urls := []string{
		"/home",
		"/book",
		"/cart",
		"/login",
		"/register",
	}
	log := ""
	for i := 0; i < 1000; i++ {
		timeStr := time.Now().String()
		url := urls[randInt(0, len(urls)-1)]
		refer := urls[randInt(0, len(urls)-1)]
		log += makeLog(timeStr, url, refer) + "\n"
	}

	fb, _ := os.OpenFile(logPath, os.O_RDWR|os.O_APPEND, 0644)
	defer fb.Close()
	fb.Write([]byte(log))
	fmt.Println("done...")
}
