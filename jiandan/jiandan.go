package jiandan

import (
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"meizi/provider"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 煎蛋站点的实现
type Jiandan struct {
	wg          *sync.WaitGroup
	dir         string // 下载到哪个文件夹下
	currentPage int    // 当前页
	userCookie  string
	client      *http.Client
	p           provider.Provider // 定义站点和对应到翻页以及解析逻辑
	nextPageUrl string            // 下一页的地址
	maxPage     int               // 最多翻多少页
}

func Create(dir string, startPage int, p provider.Provider, cookie string, client *http.Client, nextPageUrl string, wg *sync.WaitGroup, maxPage int) *Jiandan {
	return &Jiandan{dir: dir, currentPage: startPage, userCookie: cookie, p: p, client: client, nextPageUrl: nextPageUrl, wg: wg, maxPage: maxPage}
}

func createRequest(url string) *http.Request {
	req, err := http.NewRequest("GET", url, nil) // 没参数 body传nil
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.152 Safari/537.36")
	req.Header.Set("Cookie", "")
	return req
}

func (j *Jiandan) parseContent(url string) (res []string) {
	req := createRequest(url)
	resp, err := j.client.Do(req)
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

	j.p.ImageParser(doc, func(image string) {
		res = append(res, image)
	})
	j.p.PagePagination(doc, func(nextPage string) {
		j.nextPageUrl = nextPage
	})
	return res
}

func isExist(dir string) bool {
	_, err := os.Stat(dir)
	if err == nil {
		return true
	}
	return os.IsExist(err)
}

func (j *Jiandan) saveImage(imageUrls []string) {
	for _, item := range imageUrls {
		// 开多线程（协程）下载 需要注意 主线程不能退出，否则子线程还没工作完 程序就结束了 这里使用waitGroup做控制
		// 类似java中的latchdowncount 开始的时候计数器加1 做完了就减1 最后等到计数器为0 的时候主线程就不等了
		j.wg.Add(1)
		go j.downloadImage(item, j.dir)
	}
}

func (j *Jiandan) downloadImage(url string, dir string) {
	arr := strings.Split(url, "/")
	fileName := arr[len(arr)-1]
	if isExist(dir + fileName) {
		log.Printf("file name =%s has already download\n", fileName)
		return
	}
	req := createRequest(url)
	resp, err := j.client.Do(req)
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
	j.wg.Done()
}

func (j *Jiandan) parseAndSave(url string) {
	urls := j.parseContent(url)
	j.saveImage(urls)
}

func (j *Jiandan) Go() {
	if !isExist(j.dir) {
		if err := os.Mkdir(j.dir, 0777); err != nil {
			panic("can not mkdir " + j.dir)
		}
	}

	for j.currentPage > 0 {
		time.Sleep(1e9)
		var nextPageUrl = j.nextPageUrl
		if len(j.nextPageUrl) == 0 {
			nextPageUrl = j.p.UrlProvider()
		}
		log.Printf("next page url is %s\n", nextPageUrl)
		j.parseAndSave(nextPageUrl)
		j.currentPage++
		if j.currentPage > j.maxPage {
			break
		}
	}
	j.wg.Wait()
	log.Printf("All tasks have been finished...")

}
