package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func ConnecttoDataBase() (databaseptr *sqlx.DB, err error) {
	databaseptr, err = sqlx.Open("mysql", "root:*****@tcp(127.0.0.1:3306)/fzunews")
	if err != nil {
		fmt.Println("Failed to connect database:")
		fmt.Println(err)
		return
	} else {
		fmt.Println("Database init successfully")
	}
	return
}

func filter(page string, rule string, i int) [][]string { //用于进行任意匹配
	regexRULE := regexp.MustCompile(rule)
	matched := regexRULE.FindAllStringSubmatch(page, i)

	return matched
}

func pageget(url string) (pagemessage string, err error) {
	//fmt.Println("Page get start")
	pageresp, err1 := http.Get(url)
	if err1 != nil {
		err = err1
		return
	}
	defer pageresp.Body.Close()
	buf := make([]byte, 4096)
	for {
		n, err2 := pageresp.Body.Read(buf)
		if n == 0 {
			break
		}
		if err2 != nil && err2 != io.EOF {
			err = err2
			return
		}
		pagemessage += string(buf[:n])
	}
	//fmt.Println("Page get end")
	//fmt.Println(pagemessage)
	return
}

func Subcrawl(i int, DataBase *sqlx.DB, PageCode chan int) {
	var pageurl string
	//fmt.Println(i)
	if i == 1 {
		pageurl = "https://www.fzu.edu.cn/index/fdyw.htm"
	} else {
		pageurl = "https://www.fzu.edu.cn/index/fdyw/" + strconv.Itoa(75-i) + ".htm"
	}
	page, err2 := pageget(pageurl)
	if err2 != nil {
		return
	}

	/*
		fileptr, errfile := os.Create("test.txt")
		if errfile != nil {
			continue
		}
	*/

	allurls := filter(page, `<a href="(https|http)://news.fzu.edu.cn/info/(?s:(.*?))/(?s:(.*?)).htm`, -1)
	if len(allurls) != 0 {
		for _, URL := range allurls { //URL here is page's index of every news, then use the index to filter out necessary information

			subpage, errsub := pageget("https://news.fzu.edu.cn/info/" + URL[2] + "/" + URL[3] + ".htm")

			if errsub != nil {
				fmt.Println("Failed to get page:")
				fmt.Println(errsub)
				continue
			}

			readcounterurl := filter(subpage, `<script>_showDynClicks\("wbnews", (?s:(.*?)), (?s:(.*?))\)</script></span>`, 1)

			author := filter(subpage, `<span>作者：(?s:(.*?))</span>`, 1)
			contnet := filter(subpage, `<div class="v_news_content">(?s:(.*?))<div id="div_vote_id"></div>`, 1)
			publishtime := filter(subpage, `<span>发布日期:(?s:(.*?))</span>`, 1)
			title := filter(subpage, `<div class="nav01">(?s:(.*?))</h3>`, 1)

			readcount_C, _ := pageget(`https://news.fzu.edu.cn/system/resource/code/news/click/dynclicks.jsp?clickid=` + readcounterurl[0][2] + `&owner=` + readcounterurl[0][1] + `&clicktype=wbnews`)

			author_C := author[0][1]
			contnet_C := contnet[0][1]
			publishtime_C := publishtime[0][1]
			title_C := title[0][1]

			title_C = strings.Replace(title_C, `<h3>`, "", -1)
			title_C = strings.Replace(title_C, ` `, "", -1)
			publishtime_C = strings.Replace(publishtime_C, ` `, "", -1)
			author_C = strings.Replace(author_C, ` `, "", -1)
			contnet_C = strings.Replace(contnet_C, `<p>`, "", -1)
			contnet_C = strings.Replace(contnet_C, `</p>`, "", -1)
			contnet_C = strings.Replace(contnet_C, `<strong>`, "", -1)
			contnet_C = strings.Replace(contnet_C, `</strong>`, "", -1)
			contnet_C = strings.Replace(contnet_C, `</div>`, "", -1)
			contnet_C = strings.Replace(contnet_C, `<p style="text-align: center;">`, "", -1)
			contnet_C = strings.Replace(contnet_C, `<p class="vsbcontent_start">`, "", -1)
			contnet_C = strings.Replace(contnet_C, `<p class="vsbcontent_end">`, "", -1)

			publishtime_temp := strings.Replace(publishtime_C, "-", "", -1)
			publishtime_INT, _ := strconv.Atoi(publishtime_temp)

			if publishtime_INT < 20211101 {
				return
			}
			if publishtime_INT <= 20220901 {
				FeedBack, err2 := DataBase.Exec("INSERT INTO News(Title,Date,Author,ReadCount,Content)VALUES(?,?,?,?,?)", title_C, publishtime_C, author_C, readcount_C, contnet_C)
				if err2 != nil {
					fmt.Println("Exec failed, skip")
					continue
				} else {
					fmt.Println("Done!")
					fmt.Println(FeedBack)
				}
			}

			fmt.Println(title_C + "\n")
			fmt.Println("阅读数" + readcount_C + "\n")
			fmt.Println(publishtime_C + "\n")
			fmt.Println(author_C + "\n")
			fmt.Println(contnet_C + "\n")

		}
	}

	allurls2 := filter(page, `<a href="http://news.fzu.edu.cn/news/info/(?s:(.*?))/(?s:(.*?)).htm" target="_blank" title="`, -1)
	if len(allurls2) != 0 {
		for _, URL := range allurls2 { //URL here is page's index of every news, then use the index to filter out necessary information
			//fmt.Println("http://news.fzu.edu.cn/news/info/" + URL[1] + "/" + URL[2] + ".htm")
			subpage, errsub := pageget("http://news.fzu.edu.cn/news/info/" + URL[1] + "/" + URL[2] + ".htm")
			if errsub != nil {
				fmt.Println("Failed to get page:")
				fmt.Println(errsub)
				continue
			}
			readcounterurl := filter(subpage, `\("wbnews", (?s:(.*?)), (?s:(.*?))\)`, 1)
			author := filter(subpage, `<span id="author">(?s:(.*?))</span>`, 1)
			contnet := filter(subpage, `<div id="vsb_content"><div class="v_news_content">(?s:(.*?))</div></div></div>`, 1)
			publishtime := filter(subpage, `</span> <span id="fbsj">(?s:(.*?))</span>`, 1)
			title := filter(subpage, `<title>(?s:(.*?))</title>`, 1)

			readcount_C, _ := pageget(`https://news.fzu.edu.cn/system/resource/code/news/click/dynclicks.jsp?clickid=` + readcounterurl[0][2] + `&owner=` + readcounterurl[0][1] + `&clicktype=wbnews`)
			author_C := author[0][1]
			contnet_C := contnet[0][1]
			publishtime_C := publishtime[0][1]
			title_C := title[0][1]

			author_C = strings.Replace(author_C, `供稿`, "", -1)
			title_C = strings.Replace(title_C, ` `, "", -1)
			title_C = strings.Replace(title_C, "-福州大学新闻网", "", -1)
			publishtime_C = strings.Replace(publishtime_C, ` `, "", -1)
			author_C = strings.Replace(author_C, ` `, "", -1)
			contnet_C = strings.Replace(contnet_C, `<p>`, "", -1)
			contnet_C = strings.Replace(contnet_C, `</p>`, "", -1)
			contnet_C = strings.Replace(contnet_C, `<strong>`, "", -1)
			contnet_C = strings.Replace(contnet_C, `</strong>`, "", -1)
			contnet_C = strings.Replace(contnet_C, `</div>`, "", -1)
			contnet_C = strings.Replace(contnet_C, `<p style="text-align: center;">`, "", -1)
			contnet_C = strings.Replace(contnet_C, `<p class="vsbcontent_start">`, "", -1)
			contnet_C = strings.Replace(contnet_C, `<p class="vsbcontent_end">`, "", -1)

			publishtime_temp := strings.Replace(publishtime_C, "-", "", -1)
			publishtime_INT, _ := strconv.Atoi(publishtime_temp)

			if publishtime_INT < 20211101 {
				return
			}
			if publishtime_INT <= 20220901 {
				FeedBack, err2 := DataBase.Exec("INSERT INTO News(Title,Date,Author,ReadCount,Content)VALUES(?,?,?,?,?)", title_C, publishtime_C, author_C, readcount_C, contnet_C)
				if err2 != nil {
					fmt.Println("Exec failed, skip")
					continue
				} else {
					fmt.Println("Done!")
					fmt.Println(FeedBack)
				}
			}

			fmt.Println(title_C + "\n")
			fmt.Println("阅读数" + readcount_C + "\n")
			fmt.Println(publishtime_C + "\n")
			fmt.Println(author_C + "\n")
			fmt.Println(contnet_C + "\n")

		}
	}

	allurls3 := filter(page, `<a href="../../info/(?s:(.*?))/(?s:(.*?)).htm" target="_blank" title="`, -1)
	fmt.Println("------------------------------------------------------------------------------")
	fmt.Println(allurls3)
	if len(allurls3) != 0 {
		for _, URL := range allurls3 { //URL here is page's index of every news, then use the index to filter out necessary information

			//fmt.Println("https://www.fzu.edu.cn/info/" + URL[1] + "/" + URL[2] + ".htm")

			subpage, errsub := pageget(`https://www.fzu.edu.cn/info/` + URL[1] + `/` + URL[2] + `.htm`)
			fmt.Println(`https://www.fzu.edu.cn/info/` + URL[1] + `/` + URL[2] + `.htm`)
			//https://www.fzu.edu.cn/info/1062/4102.htm

			if errsub != nil {
				fmt.Println("Failed to get page:")
				fmt.Println(errsub)
				continue
			}

			readcounterurl := filter(subpage, `<script>_showDynClicks\("wbnews", (?s:(.*?)), (?s:(.*?))\)</script></span>`, 1)
			//https://www.fzu.edu.cn/system/resource/code/news/click/dynclicks.jsp?clickid=4176&owner=1779491084&clicktype=wbnews
			author := filter(subpage, `<span>作者: (?s:(.*?))</span>`, 1)
			contnet := filter(subpage, `<div class="v_news_content">(?s:(.*?))<div id="div_vote_id"></div>`, 1)
			publishtime := filter(subpage, `<span>发布时间：(?s:(.*?))</span>`, 1)
			title := filter(subpage, `<div class="nav01">(?s:(.*?))</h3>`, 1)

			readcount_C, _ := pageget(`https://www.fzu.edu.cn/system/resource/code/news/click/dynclicks.jsp?clickid=` + readcounterurl[0][2] + `&owner=` + readcounterurl[0][1] + `&clicktype=wbnews`)

			author_C := author[0][1]
			contnet_C := contnet[0][1]
			publishtime_C := publishtime[0][1]
			title_C := title[0][1]

			title_C = strings.Replace(title_C, `<h3>`, "", -1)
			title_C = strings.Replace(title_C, ` `, "", -1)
			title_C = strings.Replace(title_C, "\n", "", -1)
			publishtime_C = strings.Replace(publishtime_C, ` `, "", -1)
			author_C = strings.Replace(author_C, ` `, "", -1)
			contnet_C = strings.Replace(contnet_C, `<p>`, "", -1)
			contnet_C = strings.Replace(contnet_C, `</p>`, "", -1)
			contnet_C = strings.Replace(contnet_C, `<strong>`, "", -1)
			contnet_C = strings.Replace(contnet_C, `</strong>`, "", -1)
			contnet_C = strings.Replace(contnet_C, `</div>`, "", -1)
			contnet_C = strings.Replace(contnet_C, `<p style="text-align: center;">`, "", -1)
			contnet_C = strings.Replace(contnet_C, `<p class="vsbcontent_start">`, "", -1)
			contnet_C = strings.Replace(contnet_C, `<p class="vsbcontent_end">`, "", -1)

			publishtime_temp := strings.Replace(publishtime_C, "-", "", -1)
			publishtime_INT, _ := strconv.Atoi(publishtime_temp)

			if publishtime_INT < 20211101 {
				return
			}
			if publishtime_INT <= 20220901 {
				FeedBack, err2 := DataBase.Exec("INSERT INTO News(Title,Date,Author,ReadCount,Content)VALUES(?,?,?,?,?)", title_C, publishtime_C, author_C, readcount_C, contnet_C)
				if err2 != nil {
					fmt.Println("Exec failed, skip")
					continue
				} else {
					fmt.Println("Done!")
					fmt.Println(FeedBack)
				}
			}
			fmt.Println(title_C + "\n")
			fmt.Println("阅读数" + readcount_C + "\n")
			fmt.Println(publishtime_C + "\n")
			fmt.Println(author_C + "\n")
			fmt.Println(contnet_C + "\n")
		}
	}

	PageCode <- i
}

func crawl(DataBase *sqlx.DB) {
	PageCode := make(chan int)
	for i := 1; i <= 74; i++ {
		go Subcrawl(i, DataBase, PageCode)
	}

	for i := 1; i <= 74; i++ {
		fmt.Printf("%d Page Done", <-PageCode)
	}
}

func main() {
	DataBase, err := ConnecttoDataBase()
	if err != nil {
		return
	}

	/*
		var startpage int
		var endpage int

			fmt.Printf("请输入起始页数")
			fmt.Scan(&startpage)
			fmt.Printf("请输入结束页数")
			fmt.Scan(&endpage)
	*/

	fmt.Println("Press any key to start")
	fmt.Scanln()

	crawl(DataBase)

}
