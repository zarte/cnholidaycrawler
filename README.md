# cnholidaycrawler

中国大陆节假日放假与补班解析程序

## 说明
数据源来自国务院发布公告  
公告列表地址页：  
http://sousuo.gov.cn/s.htm?q=%E9%83%A8%E5%88%86%E8%8A%82%E5%81%87%E6%97%A5%E5%AE%89%E6%8E%92%E7%9A%84%E9%80%9A%E7%9F%A5&t=govall&timetype=timeqb&mintime=&maxtime=&sort=pubtime&sortType=1&nocorrect=  
公告详情页：  
http://www.gov.cn/zhengce/content/2021-10/25/content_5644835.htm  
推荐每年十月开始定期更新下一年缓存，公告内容格式发生改变会导致解析失败。  

## 快速上手
### 无脑更新
直接执行 main.exe 会自动获取最新一份公告  
显示create cache success即代表成功  
在程序所在目录生成对应.data缓存文件  

### 更新指定公告的内容
执行 main.exe  -url http://www.gov.cn/fuwu/2020-11/25/content_5564533.htm

### 字段说明
```cassandraql
[{
    "Date":"2022-01-01",    //日期
    "Status":1,             //1代表假日2代表补班
    "Name":"元旦"           //对应节日名称  
}]
```

