package main

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/xuri/excelize/v2"
)

//页数
var TotalPage int
//excel行数
var excelSet = 2
//失败记录
var falseSet []int
//初始化excle
func initExcel() *excelize.File {
	f := excelize.NewFile()
	err := f.SetColWidth("Sheet1", "A","A", 70)
	if err != nil {
		log.Fatal(err)
	}
	err = f.SetCellValue("Sheet1", "A1", "文章标题")
	if err != nil {
		log.Fatal(err)
	}
	err = f.SetCellValue("Sheet1", "B1", "文章日期")
	if err != nil {
		log.Fatal(err)
	}
	err = f.SetCellValue("Sheet1", "C1", "文章访问量")
	if err != nil {
		log.Fatal(err)
	}
	err = f.SetCellValue("Sheet1", "D1", "封面图")
	if err != nil {
		log.Fatal(err)
	}
	return f
}
//爬虫
func crawler(page int, f *excelize.File) {
	//获取html
	doc := getDoc(page)
	// 
	nodes := doc.Find("._1rGJJd-K0-f7qJoR9CzyeL ._1sC8pER1GUhouLkB66Mb0I").Nodes
	l := len(nodes)
	// 失败最多重试5次
	for i := 0; i < 5 && l == 0; i++ {
		// 等待3秒再进行请求，以防qps过高对请求进行拦截
		time.Sleep(3 * time.Second)
		doc = getDoc(page)
		nodes = doc.Find("._1rGJJd-K0-f7qJoR9CzyeL ._1sC8pER1GUhouLkB66Mb0I").Nodes
		l = len(nodes)
	}

	if l == 0 {
		log.Printf("页面 %d 获取失败", page)
		falseSet = append(falseSet, page)
		return
	}

	// 设置总页数
	if TotalPage == 0 {
		b := nodes[l-1].Attr[1].Val
		b = strings.TrimLeft(b, "/blog/page/")
		TotalPage, _ = strconv.Atoi(b)
	}
	parse(doc, f)
}
//获取html
func getDoc(page int) *goquery.Document {
	res, err := http.Get("https://www.aquanliang.com/blog/page/" + strconv.Itoa(page))
	if err != nil {
		log.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(res.Body)
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	//加载HTML文档
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	return doc
}
//解析文档并写入excel
func parse(doc *goquery.Document, f *excelize.File) {

	// 查找
	doc.Find("._1ySUUwWwmubujD8B44ZDzy span ._3gcd_TVhABEQqCcXHsrIpT").Each(func(i int, s *goquery.Selection) {
		// 图片
		img := s.Find("a").Find("._1wTUfLBA77F7m-CM6YysS6").Find("._2ahG-zumH-g0nsl6xhsF0s").
			Find("noscript").Nodes[0].FirstChild.Data
		img = trimImg(img)
		
		// 标题
		s = s.Find("._3HG1uUQ3C2HBEsGwDWY-zw")
		title := s.Find("._3_JaaUmGUCjKZIdiLhqtfr").Text()

		// 日期
		date := s.Find("._3TzAhzBA-XQQruZs-bwWjE").Nodes[0].LastChild.Data

		// 访问量
		view := s.Find("._2gvAnxa4Xc7IT14d5w8MI1").Nodes[0].LastChild.Data

		insertExcel(excelSet, title, date, view, img, f)
		excelSet++
	})

}
//过滤Img信息
func trimImg(img string) string {
	img = strings.TrimLeft(img, "<img src=\"")
	i :=strings.Index(img,"\" decoding=")
	img=img[0:i]
	return img
}
//插入excle
func insertExcel(i int, title string, date string, view string, img string, f *excelize.File) {
	index := strconv.Itoa(i)
	a := "A" + index
	b := "B" + index
	c := "C" + index
	d := "D" + index
	err := f.SetCellValue("Sheet1", a, title)
	if err != nil {
		log.Fatal(err)
	}
	err = f.SetCellValue("Sheet1", b, date)
	if err != nil {
		log.Fatal(err)
	}
	err = f.SetCellValue("Sheet1", c, view)
	if err != nil {
		log.Fatal(err)
	}
	err = f.SetCellValue("Sheet1", d, img)
	if err != nil {
		log.Fatal(err)
	}
}


func main() {
	log.Println("开始执行爬虫")
	f := initExcel()
	crawler(1, f)
	for i := 2; i <= TotalPage; i++ {
		//超过40页停止20秒后请求
		if TotalPage%40 == 0 {
			time.Sleep(20 * time.Second)
		}
		crawler(i, f)
	}
	for _, v := range falseSet {
		crawler(v, f)
	}
	if err := f.SaveAs("爬取到的信息.xlsx"); err != nil {
		log.Println(err)
	}
}
