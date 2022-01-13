package main

import (
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

func newRequest(url string) *http.Request {
	req, err := http.NewRequest("GET", url, nil) // 没参数 body传nil
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.152 Safari/537.36")
	req.Header.Set("Cookie", "")
	return req
}

func parseContent(url string, config *Config) (res []string) {
	var client = &http.Client{}
	req := newRequest(url)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("%s-->%s", url, strconv.Itoa(resp.StatusCode))
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	doc.Find("#comments > ol >li > div > div.row > div.text > p > img").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		if err == nil {
			img := "http:" + src
			log.Printf("parse element : image url -> %s\n", img)
			res = append(res, img)
		}
	})
	return res
}

func saveImage(imageUrl []string, config *Config) {
	for _, item := range imageUrl {
		// 开多线程（协程）下载 需要注意 主线程不能退出，否则子线程还没工作完 程序就结束了 这里使用waitGroup做控制
		// 类似java中的latchdowncount 开始的时候计数器加1 做完了就减1 最后等到计数器为0 的时候主线程就不等了
		config.wg.Add(1)
		go downloadImage(item, config.dir, config)
	}
}

func downloadImage(url string, dir string, config *Config) {
	// 这里重新又new出一个实例了 不合理
	var client = &http.Client{}
	arr := strings.Split(url, "/")
	fileName := arr[len(arr)-1]
	if isExist(dir + fileName) {
		log.Printf("file name =%s has already download\n", fileName)
		return
	}
	req := newRequest(url)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("download failed")
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("download failed")
		return
	}
	defer resp.Body.Close()

	log.Printf("begin download %s\n", fileName)
	localFile, _ := os.OpenFile(dir+fileName, os.O_CREATE|os.O_RDWR, 0777)
	if _, err := io.Copy(localFile, resp.Body); err != nil {
		panic("failed save " + fileName)
	}
	log.Printf("download %s success \n", fileName)
	// 做完了 标记一下结束
	config.wg.Done()
}

func isExist(dir string) bool {
	_, err := os.Stat(dir)
	if err == nil {
		return true
	}
	return os.IsExist(err)
}

type Config struct {
	dir         string
	currentPage int
	wg          *sync.WaitGroup
}

func newConfig(dir string, startPage int, wg *sync.WaitGroup) *Config {
	return &Config{dir: dir, currentPage: startPage, wg: wg}
}

func main() {
	var wg sync.WaitGroup
	url := "http://jandan.net/girl"
	dir := "./download_images/"
	config := newConfig(dir, 1, &wg)

	if !isExist(config.dir) {
		if err := os.Mkdir(config.dir, 0777); err != nil {
			panic("can not mkdir " + config.dir)
		}
	}

	imgUrls := parseContent(url, config)

	saveImage(imgUrls, config)

	wg.Wait() //阻塞直到所有任务完成
	log.Printf("All task has been finished...")
}
