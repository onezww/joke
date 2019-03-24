# 小姐姐图片爬取

## 爬虫基本信息
- 使用http标准库处理网络请求和下载
- goquery解析爬取的html文件
- 使用channel模拟消息队列，传输下载信息
- 使用文件记载已爬取过的相册,逻辑去重(个人项目就不引进Redis之类的,布隆过滤器稍显麻烦,这样简单)

## 配置信息
```
{
    // goroutine 数量
    "goCount": 10,

    // 保存目录
    "dirPath": "/destination/yourpath",

    // 图片保存文件夹名称
    "imgFolder": "imgs",

    // 相册描述文件
    "descriptionFile": "desc.txt",

    // 去重文件
    "recordFile": "record.log"
}

```