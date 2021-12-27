package main

import (
	_ "encoding/json"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/colly/proxy"
	_ "github.com/dlclark/regexp2"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	tenda_init = "https://www.tendacn.com/en/service/download-cata-11-1.html"
)

func main()  {
	tendaCrawler()
}


func tendaCrawler()  {
	proxy_url := "120.40.185.60:4232"
	c := colly.NewCollector(func(collector *colly.Collector){
		//这次在colly.NewCollector里面加了一项colly.Async(true)，表示抓取时异步的
		collector.Async=true
		extensions.RandomUserAgent(collector)
		collector.AllowURLRevisit = true
		collector.SetRequestTimeout(40*time.Second)
		//colly.MaxDepth(2)
	})

	rp, err := proxy.RoundRobinProxySwitcher("socks5://"+proxy_url)
	if err != nil {
		log.Fatal(err)
	}
	c.SetProxyFunc(rp)

	//err = c.Limit(&colly.LimitRule{DomainGlob: "nvd.nist.gov.*", Parallelism: 1})
	//if err != nil {
	//	log.Println(err)
	//}

	c1 := c.Clone()
	//c2 := c.Clone()

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		resp, _ := http.Get("http://http.tiqu.letecs.com/getip3?num=1&type=1&pro=&city=0&yys=0&port=2&time=1&ts=0&ys=0&cs=0&lb=1&sb=0&pb=4&mr=1&regions=&gm=4")
		body, err := ioutil.ReadAll(resp.Body)
		fmt.Println("[+]更换代理, ip为:"+string(body))
		proxy_url = string(body)
		rp, _ := proxy.RoundRobinProxySwitcher("socks5://"+proxy_url)
		c.SetProxyFunc(rp)
		err = c.Visit(tenda_init)
		if err != nil {
			log.Println(err)
		}
	})

	c1.OnResponse(func(r *colly.Response) {
		fmt.Println("Http状态: ",r.StatusCode)
		for r.StatusCode != 200{
			err := r.Request.Retry()
			if err != nil {
				log.Println(err)
			}
		}
		doc, err := htmlquery.Parse(strings.NewReader(string(r.Body)))
		if err != nil {
			log.Println("[!]",err)
		}
		firmware := htmlquery.FindOne(doc,`/html/body/div[5]/div/div/div[2]/a`)
		firmware_1 := "https:"+firmware.Attr[0].Val
		//fmt.Println(firmware_1)
		cmd := exec.Command("wget", firmware_1, "--no-check-certificate","-v")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil{
			log.Println(err)
		}
	})

	//c.OnHTML(".row .col-sm-12 .pagination", func(e *colly.HTMLElement) {
	//	fmt.Println(e.Response.StatusCode)
	//	for e.Response.StatusCode != 200{
	//		err := e.Request.Retry()
	//		if err != nil{
	//			log.Println(err)
	//		}
	//	}
	//	//lastPage, _ := e.DOM.Find("li>a").Last().Attr("href")
	//	//lastPageNumber, _ := strconv.Atoi(strings.Trim(lastPage,"/vuln/search/results?isCpeNameSearch=false&results_type=overview&form_type=Basic&search_type=all&startIndex="))
	//	//fmt.Println("[+]总页数: ",lastPageNumber)
	//	for count:=0;count<=8000;count+=20{
	//		fmt.Println("可以跳到下一页，链接是；",jump_u_3+strconv.Itoa(count))
	//		err := c1.Visit(jump_u_3+strconv.Itoa(count))
	//		if err != nil{
	//			log.Println(err)
	//		}
	//		c1.Wait()
	//	}
	//})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Http状态: ",r.StatusCode)
		for r.StatusCode != 200{
			err := r.Request.Retry()
			if err != nil {
				log.Println(err)
			}
		}
		fmt.Println("c---",r.Request.URL)
		doc, err := htmlquery.Parse(strings.NewReader(string(r.Body)))
		if err != nil {
			log.Println("[!]",err)
		}
		nodes := htmlquery.Find(doc, `//html/body/div[4]/div/div/table/tbody/tr`)
		for index,tr := range nodes{
			defer func() {
				err := recover()
				if err != nil {
					fmt.Println(err)
				}
			}()
			if index>2{
				url_1 := htmlquery.FindOne(tr,`//td/a`)
				download_url := "https:"+url_1.Attr[0].Val
				err := c1.Visit(download_url)
				if err != nil {
					log.Println(err)
				}
				c1.Wait()
			}
		}
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished", r.Request.URL)
	})

	err = c.Visit(tenda_init)
	if err != nil{
		log.Println(err)
	}
	c.Wait()
}

