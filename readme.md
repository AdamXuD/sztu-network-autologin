# SZTU 宿舍网络自动登录脚本

> 为了解决 SZTU 宿舍的校园网络无法使用路由器 PPPoE 自动登录、长时间无流量或超时会自动掉线的问题而抓包设计 Golang 版本。
> ~~这网络确实有点溺智~~

## 功能特点

- 自动登录
- 断线自动重连
- 支持容器化部署
- 支持单文件运行，无需其他环境

## 使用方法

在宿舍路由器局域网内设备运行该程序，在程序运行期间即可实现上述功能。

### 1. 不使用容器

需要先编辑`.env`文件，补全校园网账号、密码，以及路由器`wan`口`mac`，随后执行

```shell
go mod download # 运行安装必备软件包
go build -ldflags="-s -w" -o /app/main /app/main.go # 运行脚本
```

### 2. 使用容器

在项目目录下运行

```shell
docker build --tag sztu-network-autologin .
docker run -d --name sztu-network-autologin \
           -e USER_ID=你的校园网账号 \
           -e   PASSWORD=你的校园网密码 \
           -e   DEVICE_MAC=你的路由器mac \
           -e  CHECK_INTERVAL=5 \
           -e   RETRY_MAXCOUNT=5 \
           sztu-network-autologin
```
