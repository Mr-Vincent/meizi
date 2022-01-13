package main

import (
	"meizi/jiandan"
	"meizi/provider"
	"net/http"
	"sync"
)

var client = &http.Client{}
var wg sync.WaitGroup

func main() {
	// 最大采集10页数据
	jiandan.Create("./download_images/", 83, provider.DefaultProvider(), "", client, "", &wg, 10).Go()
}
