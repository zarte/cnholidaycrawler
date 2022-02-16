package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/antchfx/htmlquery"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var Reg1 *regexp.Regexp
var Reg2 *regexp.Regexp
var Reg3 *regexp.Regexp
var Reg4 *regexp.Regexp
var Reg5 *regexp.Regexp
type Holiday struct {
	Date string
	Status int //1假日2补班
	Name string
}

func main()  {
	if !initReg(){
		println("initReg fail")
		return
	}
	//"http://www.gov.cn/zhengce/content/2021-10/25/content_5644835.htm"
	var  url string
	flag.StringVar(&url,"url","", "Use -url <url>")
	flag.Parse()
	if url !="" {
		//获取特定公告
		diycreatedata(url)
	}else {
		//自动获取最新
		autogetnewyear("http://sousuo.gov.cn/s.htm?q=%E9%83%A8%E5%88%86%E8%8A%82%E5%81%87%E6%97%A5%E5%AE%89%E6%8E%92%E7%9A%84%E9%80%9A%E7%9F%A5&t=govall&timetype=timeqb&mintime=&maxtime=&sort=pubtime&sortType=1&nocorrect=")
	}

}
/**
获取特定公告数据
 */
func diycreatedata(url string)  {
	analydetail(url)
}
/**
自动获取最新公告数据
 */
func autogetnewyear(url string)  {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, strings.NewReader(""))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.121 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		println("get web fail",err)
	}else {
		doc, err := htmlquery.Parse(resp.Body)
		if err != nil {
			println("goquery fail",err)
			return
		}
		//获取最新公告链接
		aitem :=  htmlquery.FindOne(doc, `/html/body/div[2]/div/div[1]/div[3]/ul/li[1]/h3/a`)
		if aitem==nil {
			println("autogetnewyear ",err)
			return
		}
		detailhref :=htmlquery.SelectAttr(aitem,"href")
		if detailhref==""{
			println("get detail href fail ")
			return
		}
		println("detail href:",detailhref)
		analydetail(detailhref)
	}


}
/**
公告详情分析
 */
func analydetail(url string)  {
	client := &http.Client{}
	println("page :" + url)
	req, err := http.NewRequest("GET", url, strings.NewReader(""))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.121 Safari/537.36")

	resp, err := client.Do(req)
	var year string
	if err != nil {
		println("get web fail",err)
	}else {
		doc, err := htmlquery.Parse(resp.Body)
		if err != nil {
			println("goquery fail",err)
			return
		}
		//尝试获取年
		title :=  htmlquery.FindOne(doc, `/html/body/div[3]/div[2]/div[1]/div[2]/h1`)
		if title!=nil {
			tcontent :=htmlquery.InnerText(title)
			matchres := Reg5.FindStringSubmatch(tcontent)
			if matchres != nil {
				year = matchres[1]
			}
		}
		plist :=  htmlquery.Find(doc, `//*[@id="UCAP-CONTENT"]/p`)
		var holidaydata []Holiday
		for index,item := range plist{
			if index==0 && year=="" {
				//获取年
				tcontent :=htmlquery.InnerText(item)
				matchres := Reg5.FindStringSubmatch(tcontent)
				if matchres == nil {
					println("get year fail")
					return
				}
				year = matchres[1]
			}
			if index< 3 {
				continue
			}
			spanone,err := htmlquery.Query(item,`.//span`)
			if err!=nil {
				print(err)
			}else{
				if spanone!=nil {
					title := htmlquery.InnerText(spanone)
					if strings.Index(title,"节") > 0 || strings.Index(title,"元旦") > 0 {
						//节日安排信息
						tcontent :=htmlquery.InnerText(item)
						tlist := getholidayarr(year,tcontent)
						fmt.Println(tlist)
						for _,item :=range tlist{
							holidaydata = append(holidaydata,item)
						}
					}
				}
			}
		}
		resp.Body.Close()
		//生成缓存文件
		jsondata, err := json.Marshal(holidaydata)
		if err != nil {
			println("json fail")
			println(err.Error())
			return
		}
		creatcache(year,string(jsondata))
	}
}

/**
初始化正则
 */
func initReg() bool {
	Reg,err := regexp.Compile(`(\d*?)年(\d*?)月(\d*?)日至(\d*?)日`)
	if err != nil{
		fmt.Println("reg err ...",err)
		return false
	}
	Reg1 = Reg

	Reg,err = regexp.Compile(`(\d*?)月(\d*?)日至(\d*?)月(\d*?)日`)
	if err != nil{
		fmt.Println("reg err ...",err)
		return false
	}
	Reg2 = Reg

	Reg,err = regexp.Compile(`(\d*?)月(\d*?)日至(\d*?)日`)
	if err != nil{
		fmt.Println("reg err ...",err)
		return false
	}
	Reg3 = Reg

	Reg,err = regexp.Compile(`(\d*?)月(\d*?)日`)
	if err != nil{
		fmt.Println("reg err ...",err)
		return false
	}
	Reg4 = Reg

	Reg,err = regexp.Compile(`(\d*?)年`)
	if err != nil{
		fmt.Println("reg err ...",err)
		return false
	}
	Reg5 = Reg
	return true
}

/**
获取放假时间数组
 */
func getholidayarr(year string, content string) []Holiday {
	tmparr := strings.Split(content,"：")
	if len(tmparr)!=2 {
		return nil
	}
	holidayname := tmparr[0][6:len(tmparr[0])]

	//获取放假时间
	var holidayarr []string

	matchres := Reg1.FindStringSubmatch(tmparr[1])
	if matchres!=nil {
		holidayarr = getdatearr(matchres[1],matchres[2],matchres[3],year,matchres[2],matchres[4])
	}

	if holidayarr==nil {
		matchres = Reg2.FindStringSubmatch(tmparr[1])
		if matchres!=nil {
			holidayarr = getdatearr(year,matchres[1],matchres[2],year,matchres[3],matchres[4])
		}
	}

	if holidayarr==nil {
		matchres = Reg3.FindStringSubmatch(tmparr[1])
		if matchres!=nil {
			holidayarr = getdatearr(year,matchres[1],matchres[2],year,matchres[1],matchres[3])
		}
	}
	if holidayarr==nil {
		println("获取放假日期失败")
		return nil
	}
	var reslult []Holiday
	for _, datestr := range holidayarr{
		var tholiday Holiday
		tholiday.Date = datestr
		tholiday.Name = holidayname
		tholiday.Status = 1
		reslult = append(reslult,tholiday)
	}
	tmparr =strings.Split(tmparr[1],"。")
	if len(tmparr)!=3 {
		//不用补班
		return reslult
	}
	//获取补班
	workarr := getworkdayarr(year,tmparr[1])
	for _, datestr := range workarr{
		var tholiday Holiday
		tholiday.Date = datestr
		tholiday.Name = holidayname
		tholiday.Status = 2
		reslult = append(reslult,tholiday)
	}
	return reslult
}
/**
获取补班时间数组
 */
func getworkdayarr(year string,content string) []string {
	matchlist := Reg4.FindAllStringSubmatch(content,-1)
	var datearr []string
	for _, matchitem := range matchlist{
		datestr :=year+"-"
		if len(matchitem[1])<=1{
			datestr +="0"
		}
		datestr += matchitem[1]+"-"
		if len(matchitem[2])<=1{
			datestr +="0"
		}
		datestr += matchitem[2]
		datearr =append(datearr,datestr)
	}
	return datearr
}

/**
获取时间区间
 */
func getdatearr(year string,month string,day string,endyear string,endmonth string,endday string)[]string  {
	var startday string
	startday +=year+"-"
	if len(month)<=1{
		startday +="0"
	}
	startday +=month+"-"
	if len(day)<=1{
		startday +="0"
	}
	startday +=day+" 00:00:00"


	var endtime string
	endtime +=endyear+"-"
	if len(endmonth)<=1{
		endtime +="0"
	}
	endtime +=endmonth+"-"
	if len(endday)<=1{
		endtime +="0"
	}
	endtime +=endday+" 00:00:00"
	starttimeset,err :=time.Parse("2006-01-02 15:04:05",startday)
	if err!=nil{
		fmt.Println("Parse startime fail:",err)
		return nil
	}
	endtimeset,err :=time.Parse("2006-01-02 15:04:05",endtime)
	if err!=nil{
		fmt.Println("Parse startime fail:",err)
		return nil
	}
	m, _ := time.ParseDuration("+24h")
	var datearr []string
	for {
		 datearr =append(datearr,starttimeset.Format("2006-01-02"))
		 starttimeset = starttimeset.Add(m)
		 if starttimeset.Unix()>=endtimeset.Unix() {
				break
		 }
	}
	datearr =append(datearr,endtimeset.Format("2006-01-02"))
	return datearr
}

/**
生成缓存
 */
func creatcache(year string,wireteString string){
	var f    *os.File
	var err   error

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println(err)
		dir ="./"
	}
	curpath := dir+"/"
	var filename = curpath+"/holiday_"+year+".data"

	if checkFileIsExist(filename) {  //如果文件存在
		f, err = os.OpenFile(filename, os.O_APPEND,0)  //打开文件
	}else {
		f, err = os.Create(filename)  //创建文件
	}
	if err!=nil {
		fmt.Println(err)
		return
	}

	_,err = io.WriteString(f, wireteString) //写入文件(字符串)
	if err!=nil {
		fmt.Println(err)
		return
	}
	println("create cache success:"+year)
}
func checkFileIsExist(path string) (bool) {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	fmt.Println(err)
	return false
}
