<h1 align="center">
    <img width="300" src="./assets/logo.png" alt="">
</h1>

<p align="center">
   <a href="https://github.com/getcharzp/light-proxy/fork" target="blank">
      <img src="https://img.shields.io/github/forks/getcharzp/light-proxy?style=for-the-badge" alt="LocalAI forks"/>
   </a>
   <a href="https://github.com/getcharzp/light-proxy/stargazers" target="blank">
      <img src="https://img.shields.io/github/stars/getcharzp/light-proxy?style=for-the-badge" alt="LocalAI stars"/>
   </a>
   <a href="https://github.com/getcharzp/light-proxy/pulls" target="blank">
      <img src="https://img.shields.io/github/issues-pr/getcharzp/light-proxy?style=for-the-badge" alt="LocalAI pull-requests"/>
   </a>
   <a href='https://github.com/getcharzp/light-proxy/releases'>
      <img src='https://img.shields.io/github/release/getcharzp/light-proxy?&label=Latest&style=for-the-badge'>
   </a>
</p>

+ **LightProxy** 提供了桥接，代理上网等功能。适用于在同一局域网内，设备A能上网，设备B不能上网的情况（此时，设备B可以通过LightProxy借助设备A上网）。
+ 特别是工厂的环境，很多检测服务器为了安全是不能上网的，只有用于远程的 PC 能上网，此时通过 **LightProxy** 可以将 PC 的代理服务桥接到设备B上，让设备B也能临时上网。

## relay (代理上网)

+ 流程图：

![LightProxy-Relay.png](assets/LightProxy-Relay.png)

+ 命令：

```shell
# 代理转发（默认端口：8080）
proxy relay

# 指定端口
proxy relay --port 8000

# 指定端口、证书
proxy relay --port 8000 --cert light-proxy.crt --key light-proxy.key

# 指定端口、证书、认证token
proxy relay --port 8000 --cert light-proxy.crt --key light-proxy.key --auth 123
```

## bridge (桥接转发)

+ 流程图：

![LightProxy-Bridge.png](assets/LightProxy-Bridge.png)

+ 命令：

```shell
# 1. 桥接转发（默认端口：8080）
proxy bridge --relay 192.168.1.3:8080

# 2. 桥接转发，指定端口
proxy bridge --relay 192.168.1.3:8080 --port 8000

# PS：桥接转发 Clash 的网络
proxy bridge --relay 127.0.0.1:7890
```

## config (配置管理)

说明：如果需要在终端中使用配置的代理地址，需要打开新的终端

```shell
# 清除配置
sudo proxy config --set 0

# 设置HTTP代理（set 后面的参数为 [relay 所在服务器的IP]:[relay 指定的端口]）
sudo proxy config --set http://192.168.1.3:8080

# 设置TLS代理
sudo proxy config --set https://192.168.1.3:8080

# 设置带认证token的TLS代理
sudo proxy config --set https://token@192.168.1.3:8080
```

## 帮助命令

```bash
# 自建证书的生成命令
# openssl req -x509 -newkey rsa:4096 -keyout light-proxy.key -out light-proxy.crt -days 365 -nodes -subj "/CN=ip" -addext "subjectAltName = IP:ip"
# 示例：
openssl req -x509 -newkey rsa:4096 -keyout light-proxy.key -out light-proxy.crt -days 365 -nodes -subj "/CN=192.168.1.3" -addext "subjectAltName = IP:192.168.1.3"
```

### Ubuntu 配置自建证书

1. 拷贝生成的 `light-proxy.crt` 自建证书到 `/usr/local/share/ca-certificates/` 目录下

2. 执行 `sudo update-ca-certificates` 命令更新证书
