# NJU 书童

书童嘛，侍候你读书。

## 现在支持的功能

✅ 监测寝室电量，电量不足时执行脚本 `electric.sh`（再也不用担心寝室突然<del>失明</del>停电啦～）

## 使用方法

1. 准备一台服务器，也可以是个人电脑或者机房电脑等，安装 Docker 和 Docker Compose
2. 在合适的位置 `mkdir nju-shutong && cd nju-shutong` 创建并进入文件夹
3. 下载应用程序配置 `curl https://raw.githubusercontent.com/chenghaopeng/NJU-ShuTong/master/docker-compose.yaml -o docker-compose.yaml`
4. 准备环境变量 `echo "EAI_SESS=【你的 EAI_SESS】" > .env`。`EAI_SESS` 的获取方法见下文
5. `mkdir scripts`，准备脚本，以供在条件满足时调用，如电量提醒脚本 `vim ./scripts/electric.sh`，内容自定义
6. 启动程序 `docker-compose up -d` 完成！

## EAI_SESS 的获取方法

EAI_SESS 是你在访问信息门户时的“身份证”，有了这个，书童才能代替你去访问信息门户

1. 访问 `https://wx.nju.edu.cn/njucharge/wap/electric/index`
2. 在开发者工具中随便点开一个请求，在 Cookie 里就能找到这个字段 ![EAI_SESS](images/eai_sess.png)

## 各功能配置

### 寝室电量

1. 配置所监测的房间
   1. 访问 `https://wx.nju.edu.cn/njucharge/wap/electric/index`
   2. 选择你要监测的寝室
   3. 点击“去充值”，跳转到新的页面
   4. 网址是这种形式 `https://wx.nju.edu.cn/njucharge/wap/electric/charge?area_id=【校区编号】&area_name=【校区名称】&build_id=【楼栋编号】&build_name=【楼栋名称】&room_name=【房间名称】&room_id=【房间号】`
   5. 这里我们用到校区编号、楼栋编号和房间号
   6. `echo 'ROOM="【校区编号】,【楼栋编号】,【房间号】" >> .env'` 注意是英文逗号隔开。例子：`ROOM="01,gl43,0101"` 表示鼓楼陶三 101 房
   7. 接下来配置电量提醒的脚本 `vim ./scripts/electric.sh`。书童在调用该脚本时，会添加一个叫做 `CONTENT` 的环境变量，表示提醒的文本内容，可以在你的脚本中直接使用。例子：你可以在脚本中这么写 `curl --request POST https://mou.ge.wang.zhi/fasong/xiaoxi --header 'Content-Type: application/json' --data-raw "{\"data\":{\"content\":\"$CONTENT\"}}"` 表示把 CONTENT 发送到某个特定网址，让它来后续处理，比如由它来将 CONTENT 发送给某人 ![电量提醒](images/electric_notify.png)
   8. 这样就大功告成了
