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
	"regexp"
	"strings"
	"time"
)

const (
	d_link_init = "https://ftp.dlink.ru/pub/Router/"
)

func main()  {
	Crawler()
}


func Crawler()  {
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
		err = c.Visit(d_link_init)
		if err != nil {
			log.Println(err)
		}
	})

	c1.OnResponse(func(r *colly.Response) {
		fmt.Println("Http状态: ",r.StatusCode)
		doc, err := htmlquery.Parse(strings.NewReader(string(r.Body)))
		if err != nil {
			log.Println("[!]",err)
		}
		nodes := htmlquery.Find(doc, `//html/body/table/tbody/tr`)
		for index,tr := range nodes {
			if index>2{
				firm_name := htmlquery.InnerText(tr)
				re := regexp.MustCompile(`(.+?)(bin|zip|Bin|rar|exe|EXE|BIN|bix|Zip|ZIP|tar|_RU|_ru|dlf|Ing|img|map)`)//bin|zip|Bin|rar
				firm_name_1 := re.FindString(firm_name)
				if len(firm_name_1)>0{
					firm_url := r.Request.URL.String()+firm_name_1
					//fmt.Println(r.Request.URL.String()+firm_name_1)
					cmd := exec.Command("wget", firm_url, "--no-check-certificate","-v")
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					err := cmd.Run()
					if err != nil{
						log.Println(err)
					}
				}
			}
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
		nodes := htmlquery.Find(doc, `//html/body/table/tbody/tr`)
		fmt.Println(len(nodes))
		for index,tr := range nodes{
			defer func() {
				err := recover()
				if err != nil {
					fmt.Println(err)
				}
			}()
			if index>3{
				url_1 := htmlquery.InnerText(tr)
				model := strings.Split(url_1,"/")
				firm_url := d_link_init+model[0]+"/Firmware"
				fmt.Println(index,firm_url)
				err := c1.Visit(firm_url)
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

	err = c.Visit(d_link_init)
	if err != nil{
		log.Println(err)
	}
	c.Wait()
}

