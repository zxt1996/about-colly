package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"log"
	"math/rand"
	"os"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandomString() string {
	b := make([]byte, rand.Intn(10)+10)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

//以小写字母开头的字段成员是无法被外部直接访问的，所以 struct 在进行 json、xml、gob 等格式的 encode 操作时，这些私有字段会被忽略，导出时得到零值
//所以字段都要以大写开头，不然到时转json的时候看不到
type MovieMes struct {
	Name string
	Img string
}

func toJson(m MovieMes)  {
	//将结构体数据转换为json格式
	b, err := json.Marshal(&m)

	if err != nil {
		fmt.Println(err)
		return
	}

	//生成JSON文件
	//在UNIX型系统中，文件的默认权限为0644，即所有者能够读取和写入，而其他人只能读取
	f, err := os.OpenFile("movie.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write(b); err != nil {
		log.Fatal(err)
	}

	f.WriteString("\n")

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func main()  {
	//Colly的主体是一个收集对象。Collector管理网络通信，并负责在收集器作业运行时执行附加的回调
	c := colly.NewCollector(
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{DomainGlob:  "*.douban.*", Parallelism: 5})

	//在请求前调用
	c.OnRequest(func(r *colly.Request) {
		//用户代理，是标明身份的一种标识, 在每个请求中更改用户代理
		r.Headers.Set("User-Agent", RandomString())
		log.Println("Visiting", r.URL)
	})

	//发生错误时调用
	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	//收到的是HTML后调用
	//第一参数是要匹配的元素
	c.OnHTML("li", func(e *colly.HTMLElement) {
		//利用ChildAttr获取某个标签中的特定元素
		var aboutImg string = e.ChildAttr(".pic a img", "src")
		//ChildText返回匹配内容的文本
		var movieName string = e.ChildText(".hd a .title")

		log.Println("电影名：", movieName, " 海报：", aboutImg)

		if movieName != "" {
			oneMovie := MovieMes {
				movieName,
				aboutImg,
			}
			toJson(oneMovie)
		}
	})

	c.OnHTML(".paginator a", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	c.Visit("https://movie.douban.com/top250")
	c.Wait()

}
