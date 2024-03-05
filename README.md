# SSHVPN

[![GitHub Workflow][1]](https://github.com/wencaiwulue/sshvpn/actions)
[![Go Version][2]](https://github.com/wencaiwulue/sshvpn/blob/master/go.mod)
[![Go Report][3]](https://goreportcard.com/report/github.com/wencaiwulue/sshvpn)
[![GitHub License][4]](https://github.com/wencaiwulue/sshvpn/blob/main/LICENSE)
[![Releases][5]](https://github.com/wencaiwulue/sshvpn/releases)

[1]: https://img.shields.io/github/actions/workflow/status/wencaiwulue/sshvpn/release.yml?logo=github

[2]: https://img.shields.io/github/go-mod/go-version/wencaiwulue/sshvpn?logo=go

[3]: https://goreportcard.com/badge/github.com/wencaiwulue/sshvpn?style=flat

[4]: https://img.shields.io/github/license/wencaiwulue/sshvpn

[5]: https://img.shields.io/github/v/release/wencaiwulue/sshvpn?logo=smartthings

[中文](README_ZH.md) | [English](README.md)

A safety virtual personal network over SSH and Gvisor

## Content

1. [QuickStart](./README.md#quickstart)
2. [Functions](./README.md#functions)
3. [Architecture](./README.md#architecture)

## QuickStart

### Install server(On ssh server)

SSH login to your ssh server

#### Download sshvpn binary

```shell
curl -Lo sshvpn.zip https://github.com/wencaiwulue/sshvpn/releases/download/v1.0.0/sshvpn_v1.0.0_linux_amd64.zip && unzip -d sshvpn sshvpn.zip && mv ./sshvpn/bin/sshvpn /usr/local/bin
```

#### Run server

```shell
nohup sshvpn server &
```

### Install client(On local computer)

Download from GitHub release
[release](https://github.com/wencaiwulue/sshvpn/releases/latest)

```shell
➜  sshvpn version
SSHVPN: CLI
    Version: v1.0.0
    Branch: HEAD
    Git commit: 459082458113b828a5d73c718be42631699b44b2
    Built time: 2024-02-19 15:25:26
    Built OS/Arch: linux/amd64
    Built Go version: go1.21.7
```

## Functions

### Connect to remote ssh server network

```shell
➜  ~ sshvpn client --ssh-addr xxx.xxx.xxx.xxx:22 --ssh-username root --ssh-password xxx
DEBU[0000] [sudo --preserve-env sshvpn client --ssh-addr xxx.xxx.xxx.xxx:22 --ssh-username root --ssh-password xxx]
DEBU[0001] [tun] ifconfig utun9 inet 223.253.0.1/32 223.253.0.1 mtu 1500 up
DEBU[0001] [tun] ifconfig utun9 inet6 ::1 prefixlen 128 alias
DEBU[0001] [tun] route add -net xxx.xxx.xxx.xxx/32 -interface utun9
DEBU[0001] [tun] route add -net xxx.xxx.xxx.xxx/32 -interface utun9
DEBU[0001] [tun] 223.253.0.1: name: utun9, mtu: 1500, addrs: [223.253.0.1/32 fe80::bed0:74ff:fe4c:9790/64 ::1/128]
DEBU[0001] networksetup -getdnsservers Wi-Fi
DEBU[0001] networksetup -setdnsservers Wi-Fi xxx.xxx.xxx.xxx
INFO[0001] you can use VPN now~
DEBU[0001] [TUN-RAW] IP-Protocol: IPv6HopByHop, SRC: fe80::bed0:74ff:fe4c:9790, DST: ff02::16, Length: 96
DEBU[0001] [TUN-RAW] IP-Protocol: IPv6HopByHop, SRC: fe80::bed0:74ff:fe4c:9790, DST: ff02::16, Length: 76
DEBU[0001] [TUN-UDP] Debug: LocalPort: 53, LocalAddress: xxx.xxx.xxx.xxx, RemotePort: 60605, RemoteAddress 223.253.0.1
DEBU[0001] [TUN-UDP] IP-Protocol: UDP, SRC: 223.253.0.1, DST: xxx.xxx.xxx.xxx, Length: 85
DEBU[0001] [TUN-UDP] Debug: LocalPort: 53, LocalAddress: xxx.xxx.xxx.xxx, RemotePort: 54483, RemoteAddress 223.253.0.1
...
```

Leave this terminal alone

```shell
➜  ~ curl www.google.com -L
<!doctype html><html itemscope="" itemtype="http://schema.org/WebPage" lang="en-SG"><head><meta content="text/html; charset=UTF-8" http-equiv="Content-Type"><meta content="/images/branding/googleg/1x/googleg_standard_color_128dp.png" itemprop="image"><title>Google</title><script nonce="KGbHmdii_0HKvWBmaQ_OAQ">...
```

### Open local PC browser to visit web via SSH server network

### Multiple Protocol

- TCP
- UDP

### Cross-platform

- macOS
- Linux
- Windows

## Architecture

Architecture can be found [here](/docs/en/Architecture.md).