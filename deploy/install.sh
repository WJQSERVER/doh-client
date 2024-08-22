#! /bin/bash

# 检测是否安装了wget
if ! command -v wget >/dev/null 2>&1; then
  echo "wget is not installed, please install it first."
  exit 1
fi

#拉取最新版本的doh-client
mkdir -p /root/data/doh-client
VERSION=$(curl -s https://raw.githubusercontent.com/WJQSERVER/doh-client/main/VERSION)
wget -O /root/data/doh-client.tar.gz https://github.com/WJQSERVER/doh-client/releases/download/$VERSION/doh-client-linux-amd64.tar.gz

#解压并移动到/usr/bin目录
tar -zxvf /root/data/doh-client.tar.gz -C /root/data/doh-client
rm -rf /root/data/doh-client.tar.gz

#设置权限
chmod +x /root/data/doh-client/doh-client
chown root:root /root/data/doh-client/doh-client

#拉取配置文件
mkdir -p /root/data/doh-client/config
mkdir -p /root/data/doh-client/log
wget -O /root/data/doh-client/config/config.yaml https://raw.githubusercontent.com/WJQSERVER/doh-client/main/config/config.yaml

#拉取systemd service文件
wget -O /etc/systemd/system/doh-client.service https://raw.githubusercontent.com/WJQSERVER/doh-client/main/deploy/doh-client.service

#启动服务
systemctl enable doh-client
systemctl start doh-client
