package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"joke/set"
	"joke/util"
	"net/http"
	"net/url"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	iconv "github.com/djimenez/iconv-go"
)

type Config struct {
	GoCount         int
	DirPath         string
	ImgFolder       string
	DescriptionFile string
	RecordFile      string
}

type urlMq struct {
	title string
	url   string
}

// ConfigObj 存储配置json信息
var ConfigObj Config

// UrlSet 路径集合
var UrlSet = set.New()

// RecodFile 去重url日志
var RecodFile *os.File

// MQ gorotine之间的消息队列，传输路径
var MQ = make(chan urlMq, 100)

func initConfig() {
	file, err := os.Open("./config.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	err = json.NewDecoder(file).Decode(&ConfigObj)
	if err != nil {
		fmt.Println("json decode error", err.Error())
		panic(err)
	}
	// 设置保存目录
	ConfigObj.DirPath = strings.TrimSpace(ConfigObj.DirPath)
	if len(ConfigObj.DirPath) == 0 {
		ConfigObj.DirPath = "./data"
	}
	ConfigObj.ImgFolder = path.Join(ConfigObj.DirPath, ConfigObj.ImgFolder)
	ConfigObj.RecordFile = path.Join(ConfigObj.DirPath, ConfigObj.RecordFile)
	os.MkdirAll(ConfigObj.ImgFolder, 0766)
}

func initURLSet(path string, us *set.Set) {
	file, err := os.Open(path)
	if err != nil {
		println("first")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		us.Add(scanner.Text())
	}
}

// 获取记录日志文件
func getRecordFile() *os.File {
	if !util.IsPathExists(ConfigObj.RecordFile) {
		f, err := os.OpenFile(ConfigObj.RecordFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		return f
	}
	f, err := os.OpenFile(ConfigObj.RecordFile, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	return f
}

// 获取每个类目下的类别id
func getTypeNum(doc *goquery.Document) string {
	converter, _ := iconv.NewConverter("gb2312", "utf-8")
	typeStr := ""
	doc.Find(".NewPages ul li a[href]:first-child").Each(func(i int, s *goquery.Selection) {
		if len(typeStr) > 0 {
			return
		}
		tmp, ok := s.Attr("href")
		if !ok {
			return
		}
		out, _ := converter.ConvertString(tmp)
		prefix := strings.Split(out, ".")[0]
		typeStr = strings.Split(prefix, "_")[1]
	})
	return typeStr
}

// 下载具体相册的图片
func loadImage(url string, dir string) {
	prefix := strings.Split(url, ".html")[0]
	count := 1
	reqUrl := url
	client := &http.Client{}
	for {
		response, err := http.Get(reqUrl)
		if err != nil {
			print("come in err", err.Error())
			return
		}
		if response.StatusCode != 200 {
			return
		}
		doc, err := goquery.NewDocumentFromReader(response.Body)
		if err != nil {
			println("哎呀，goquery 解析出错啦", err.Error())
			return
		}
		imgUrl := ""
		doc.Find("#picBody p a img:first-child").Each(func(i int, s *goquery.Selection) {
			src, ok := s.Attr("src")
			if !ok {
				return
			}
			imgUrl = src
		})
		if len(imgUrl) > 0 {
			var req *http.Request
			req, _ = http.NewRequest("GET", imgUrl, nil)
			req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.121 Safari/537.36")
			req.Header.Add("Host", "t1.hddhhn.com")
			req.Header.Add("Referer", reqUrl)
			response, err := client.Do(req)
			if err != nil {
				println("图片请求失败了", imgUrl, err.Error())
				return
			}
			if response.StatusCode != 200 {
				return
			}

			filename := path.Join(dir, strconv.Itoa(count)+".jpg")
			out, _ := os.Create(filename)
			io.Copy(out, response.Body)
			response.Body.Close()
			println("图片下载完成", filename)
		}
		response.Body.Close()
		count++
		reqUrl = prefix + "_" + strconv.Itoa(count) + ".html"
	}
}

func generateProject(wg *sync.WaitGroup) {
	defer wg.Done()
	for v := range MQ {
		ok := UrlSet.AddNX(v.url)
		if !ok {
			continue
		}
		imgFolder := path.Join(ConfigObj.ImgFolder, v.title)
		os.MkdirAll(imgFolder, 0766)
		loadImage(v.url, imgFolder)
		RecodFile.Write([]byte(v.url + "\n"))
	}
}

// 发送消息到通道
func sendDataToMQ(domain string, doc *goquery.Document, converter *iconv.Converter) {
	doc.Find(".MeinvTuPianBox ul li a[class=MMPic]").Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok {
			return
		}
		href, _ = converter.ConvertString(href)
		u, err := url.Parse(domain)
		if err != nil {
			return
		}
		// 拼接每个相册的地址
		u.Path = path.Join(u.Path, href)
		href = u.String()

		if UrlSet.Has(href) {
			return
		}
		title, ok := s.Attr("title")
		if !ok {
			return
		}
		title, _ = converter.ConvertString(title)
		MQ <- urlMq{title: title, url: href}
	})
}

func spider(wg *sync.WaitGroup) {
	defer wg.Done()
	domain := "https://www.2717.com/"
	module := "https://www.2717.com/ent/meinvtupian/"

	defer close(MQ)
	response, err := http.Get(module)
	if err != nil {
		panic(err)
	}
	if response.StatusCode != 200 {
		fmt.Println("哎呀，好像结束了")
	}
	converter, _ := iconv.NewConverter("gb2312", "utf-8")
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		println("哎呀，goquery 解析出错啦")
		panic(err)
	}
	typeStr := getTypeNum(doc)
	fmt.Printf("%v\n", typeStr)
	response.Body.Close()
	sendDataToMQ(domain, doc, converter)
}

func main() {
	runtime.GOMAXPROCS(1)
	var wait sync.WaitGroup
	initConfig()
	initURLSet(ConfigObj.RecordFile, UrlSet)
	RecodFile = getRecordFile()

	wait.Add(ConfigObj.GoCount + 1)
	for i := 0; i < ConfigObj.GoCount; i++ {
		go generateProject(&wait)
	}
	go spider(&wait)
	wait.Wait()
}
